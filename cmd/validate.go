package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Result is the outcome of checking one asset listed in a PKL.
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

// Results is the full validation report. Errors counts failed assets.
type Results struct {
	Results []Result `json:"results"`
	Errors  int      `json:"errors"`
}

// Validate checks every asset listed in every PKL found under dir.
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

	var paths map[string]string
	if files.assetMap != "" {
		if paths, err = readAssetMap(files.assetMap); err != nil {
			return res, err
		}
	}

	for _, pkl := range files.packingLists {
		assets, err := readPackingList(pkl)
		if err != nil {
			return res, err
		}
		for _, a := range assets {
			r := checkAsset(dir, a, paths)
			if r.Error != "" || !r.HashValid || !r.SizeValid {
				res.Errors++
			}
			res.Results = append(res.Results, r)
		}
	}
	return res, nil
}

func checkAsset(dir string, a pklAsset, paths map[string]string) Result {
	id := strings.TrimSpace(a.ID)
	rel := strings.TrimSpace(a.Name)
	if p, ok := paths[id]; ok {
		rel = p
	}

	r := Result{
		Name:         rel,
		ReportedSize: strings.TrimSpace(a.Size),
		ReportedHash: strings.TrimSpace(a.Hash),
		ActualSize:   -1,
	}
	if rel == "" {
		r.Error = fmt.Sprintf("asset %s has no file name", id)
		return r
	}

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
		r.Error = "not a regular file"
		return r
	}

	r.ActualSize = info.Size()
	if reported, err := strconv.ParseInt(r.ReportedSize, 10, 64); err == nil {
		r.SizeValid = reported == r.ActualSize
	}

	actual, err := sha1File(full)
	if err != nil {
		r.Error = fmt.Sprintf("cannot hash file: %v", err)
		return r
	}
	r.ActualHash = actual
	r.HashValid = actual == r.ReportedHash
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
