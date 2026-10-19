package main

import (
	"encoding/xml"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// dcpFiles holds the files discovered under an input directory. All paths are absolute or input-relative as given.
type dcpFiles struct {
	packingLists []string // PKLs, detected by root element
	assetMaps    []string // every ASSETMAP found, one per DCP directory
	regular      []string // every regular file under the input
}

// findDcpFiles walks dir and locates PackingList files and ASSETMAPs.
// Matching is case-insensitive and only regular files are considered.
func findDcpFiles(dir string) (dcpFiles, error) {
	var out dcpFiles

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
			out.assetMaps = append(out.assetMaps, path)
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
	return out, err
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

// within reports whether p lies at or below root.
func within(root, p string) bool {
	r, err := filepath.Rel(root, p)
	return err == nil && r != ".." && !strings.HasPrefix(r, ".."+string(filepath.Separator))
}

// normRel canonicalises an ASSETMAP or PKL path for comparison.
func normRel(rel string) string {
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(rel))))
}
