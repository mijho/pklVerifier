package main

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Result is the outcome of checking one asset listed in a PKL, or one
// ASSETMAP entry that no PKL lists.
type Result struct {
	Name         string `json:"name"`
	ReportedSize string `json:"reported_size"`
	ActualSize   int64  `json:"actual_size"`
	ReportedHash string `json:"reported_hash"`
	ActualHash   string `json:"actual_hash"`
	HashValid    bool   `json:"hash_valid"`
	SizeValid    bool   `json:"size_valid"`
	Error        string `json:"error,omitempty"`
}

func (r Result) failed() bool {
	return r.Error != "" || !r.HashValid || !r.SizeValid
}

// Results is the full validation report. Errors counts every failure:
// failed assets, ASSETMAP entries missing from the PKL, and unlisted files.
type Results struct {
	Results  []Result `json:"results"`
	Unlisted []string `json:"unlisted,omitempty"`
	Errors   int      `json:"errors"`
}

// Validate checks every PKL under dir and cross-checks the ASSETMAP.
// Per-asset problems are recorded in the results rather than aborting the run.
func Validate(dir string) (Results, error) {
	res := Results{Results: []Result{}}

	files, err := findDcpFiles(dir)
	if err != nil {
		return res, err
	}
	if len(files.packingLists) == 0 {
		return res, fmt.Errorf("no PackingList (PKL) found in %s", dir)
	}

	var amap map[string][]string
	if files.assetMap != "" {
		if amap, err = readAssetMap(files.assetMap); err != nil {
			return res, err
		}
	}

	pklFiles := make(map[string]bool, len(files.packingLists))
	for _, p := range files.packingLists {
		pklFiles[relSlash(dir, p)] = true
	}

	covered := make(map[string]bool)
	for _, pkl := range files.packingLists {
		assets, err := readPackingList(pkl)
		if err != nil {
			return res, err
		}
		for _, a := range assets {
			rels := assetPaths(a, amap)
			for _, rel := range rels {
				covered[normRel(rel)] = true
			}
			r := checkAsset(dir, a, rels)
			if r.failed() {
				res.Errors++
			}
			res.Results = append(res.Results, r)
		}
	}

	if files.assetMap == "" {
		return res, nil
	}

	// ASSETMAP entries that no PKL lists. PKL files themselves are expected here.
	listed := make(map[string]bool)
	var listedRels []string
	for _, chunks := range amap {
		for _, c := range chunks {
			n := normRel(c)
			if !listed[n] {
				listed[n] = true
				listedRels = append(listedRels, n)
			}
		}
	}
	sort.Strings(listedRels)
	for _, rel := range listedRels {
		if covered[rel] || pklFiles[rel] {
			continue
		}
		res.Results = append(res.Results, Result{
			Name:       rel,
			ActualSize: -1,
			Error:      "listed in ASSETMAP but not in any PKL",
		})
		res.Errors++
	}

	// Files on disk that the ASSETMAP does not mention.
	assetMapRel := relSlash(dir, files.assetMap)
	for _, p := range files.regular {
		rel := relSlash(dir, p)
		if listed[rel] || pklFiles[rel] || rel == assetMapRel {
			continue
		}
		if strings.EqualFold(filepath.Base(rel), "VOLINDEX.xml") {
			continue
		}
		res.Unlisted = append(res.Unlisted, rel)
		res.Errors++
	}
	return res, nil
}

// assetPaths returns the file path(s) for an asset, in chunk order. It prefers
// the ASSETMAP and falls back to the PKL's OriginalFileName.
func assetPaths(a pklAsset, amap map[string][]string) []string {
	if chunks, ok := amap[strings.TrimSpace(a.ID)]; ok {
		return chunks
	}
	if name := strings.TrimSpace(a.Name); name != "" {
		return []string{name}
	}
	return nil
}

func checkAsset(dir string, a pklAsset, rels []string) Result {
	id := strings.TrimSpace(a.ID)
	r := Result{
		Name:         strings.Join(rels, " + "),
		ReportedSize: strings.TrimSpace(a.Size),
		ReportedHash: strings.TrimSpace(a.Hash),
		ActualSize:   -1,
	}
	if len(rels) == 0 {
		r.Error = fmt.Sprintf("asset %s has no file name", id)
		return r
	}

	h := sha1.New()
	var total int64
	for _, rel := range rels {
		full, err := resolvePath(dir, rel)
		if err != nil {
			r.Error = err.Error()
			return r
		}
		info, err := os.Stat(full)
		if err != nil {
			r.Error = fmt.Sprintf("cannot stat file: %v", err)
			return r
		}
		if !info.Mode().IsRegular() {
			r.Error = fmt.Sprintf("%s: not a regular file", rel)
			return r
		}
		total += info.Size()
		if err := hashFile(h, full); err != nil {
			r.Error = fmt.Sprintf("cannot hash %s: %v", rel, err)
			return r
		}
	}

	r.ActualSize = total
	if reported, err := strconv.ParseInt(r.ReportedSize, 10, 64); err == nil {
		r.SizeValid = reported == total
	}
	r.ActualHash = base64.StdEncoding.EncodeToString(h.Sum(nil))
	r.HashValid = r.ActualHash == r.ReportedHash
	return r
}

// resolvePath joins rel onto dir and refuses paths that leave the DCP.
func resolvePath(dir, rel string) (string, error) {
	native := filepath.FromSlash(rel)
	if filepath.IsAbs(native) {
		return "", fmt.Errorf("absolute path %q is not allowed", rel)
	}
	full := filepath.Join(dir, native)
	r, err := filepath.Rel(dir, full)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", errors.New("path " + strconv.Quote(rel) + " escapes the DCP directory")
	}
	return full, nil
}
