package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// dcpFiles holds the files discovered in a DCP directory. All paths are absolute.
type dcpFiles struct {
	packingLists []string // PKLs, detected by root element
	assetMap     string   // the ASSETMAP, or "" if absent
	regular      []string // every regular file under the DCP
}

// findDcpFiles walks dir and locates PackingList files and the ASSETMAP.
// Matching is case-insensitive and only regular files are considered.
func findDcpFiles(dir string) (dcpFiles, error) {
	var out dcpFiles
	var assetMaps []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		out.regular = append(out.regular, path)

		upper := strings.ToUpper(d.Name())
		if upper == "ASSETMAP" || upper == "ASSETMAP.XML" {
			assetMaps = append(assetMaps, path)
			return nil
		}
		if !strings.HasSuffix(upper, ".XML") {
			return nil
		}
		root, err := rootElement(path)
		if err != nil {
			return nil // not parseable XML, so not a PKL
		}
		if root == "PackingList" {
			out.packingLists = append(out.packingLists, path)
		}
		return nil
	})
	if err != nil {
		return out, err
	}

	switch len(assetMaps) {
	case 0:
	case 1:
		out.assetMap = assetMaps[0]
	default:
		return out, fmt.Errorf("multiple ASSETMAP files found: %v", assetMaps)
	}
	return out, nil
}

// rootElement returns the local name of the document's root element.
func rootElement(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		if se, ok := tok.(xml.StartElement); ok {
			return se.Name.Local, nil
		}
	}
}

// hashFile streams the file at path into w.
func hashFile(w io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

// relSlash returns p relative to dir using forward slashes.
func relSlash(dir, p string) string {
	r, err := filepath.Rel(dir, p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(r)
}

// normRel canonicalises an ASSETMAP or PKL path for comparison.
func normRel(rel string) string {
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(rel))))
}
