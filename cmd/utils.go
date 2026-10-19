package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// dcpFiles holds the metadata files discovered in a DCP directory.
type dcpFiles struct {
	packingLists []string
	assetMap     string
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

// sha1File streams a file through SHA-1 and returns the base64 digest,
// which is the encoding used by SMPTE PKL hashes.
func sha1File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
