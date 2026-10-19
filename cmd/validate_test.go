package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testAsset struct {
	id   string
	rel  string // path relative to the DCP root, as listed in the ASSETMAP
	data string
}

var sampleAssets = []testAsset{
	{id: "urn:uuid:1111", rel: "video.mxf", data: "MXF-DATA-0123456789"},
	{id: "urn:uuid:2222", rel: "CPL_test.xml", data: "<CPL>fake</CPL>"},
}

func sha1b64(s string) string {
	sum := sha1.Sum([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeAssetMap(t *testing.T, dir, name string, assets []testAsset) {
	t.Helper()
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><AssetMap><AssetList>`)
	for _, a := range assets {
		fmt.Fprintf(&b, `<Asset><Id>%s</Id><ChunkList><Chunk><Path>%s</Path></Chunk></ChunkList></Asset>`, a.id, a.rel)
	}
	b.WriteString(`</AssetList></AssetMap>`)
	writeFile(t, filepath.Join(dir, name), b.String())
}

func writePKL(t *testing.T, dir, name string, assets []testAsset) {
	t.Helper()
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><PackingList><AssetList>`)
	for _, a := range assets {
		fmt.Fprintf(&b, `<Asset><Id>%s</Id><OriginalFileName>%s</OriginalFileName><Hash>%s</Hash><Size>%d</Size></Asset>`,
			a.id, filepath.Base(a.rel), sha1b64(a.data), len(a.data))
	}
	b.WriteString(`</AssetList></PackingList>`)
	writeFile(t, filepath.Join(dir, name), b.String())
}

// buildDCP writes the asset files, an ASSETMAP and a matching PKL under dir.
func buildDCP(t *testing.T, dir string, assets []testAsset) string {
	t.Helper()
	for _, a := range assets {
		writeFile(t, filepath.Join(dir, a.rel), a.data)
	}
	writeAssetMap(t, dir, "ASSETMAP.xml", assets)
	writePKL(t, dir, "PKL_test.xml", assets)
	return dir
}

func mustValidate(t *testing.T, dir string) Results {
	t.Helper()
	res, err := Validate(dir)
	if err != nil {
		t.Fatalf("Validate(%s) returned error: %v", dir, err)
	}
	return res
}

func assertAllValid(t *testing.T, res Results) {
	t.Helper()
	if res.Errors != 0 {
		t.Fatalf("expected 0 errors, got %d: %+v", res.Errors, res.Results)
	}
	for _, r := range res.Results {
		if !r.HashValid || !r.SizeValid || r.Error != "" {
			t.Errorf("expected %s valid, got %+v", r.Name, r)
		}
	}
}

func TestValidDCP(t *testing.T) {
	dir := buildDCP(t, t.TempDir(), sampleAssets)
	res := mustValidate(t, dir)
	if len(res.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res.Results))
	}
	assertAllValid(t, res)
}

func TestTamperedSameSize(t *testing.T) {
	dir := buildDCP(t, t.TempDir(), sampleAssets)
	writeFile(t, filepath.Join(dir, "video.mxf"), "MXF-DATA-0123456780") // same length, different content
	res := mustValidate(t, dir)
	if res.Errors != 1 {
		t.Fatalf("expected 1 failing asset, got %d", res.Errors)
	}
	for _, r := range res.Results {
		if r.Name == "video.mxf" && (r.HashValid || !r.SizeValid) {
			t.Errorf("expected hash invalid and size valid, got %+v", r)
		}
	}
}

func TestTamperedSize(t *testing.T) {
	dir := buildDCP(t, t.TempDir(), sampleAssets)
	writeFile(t, filepath.Join(dir, "video.mxf"), "MXF-DATA-0123456789-extra")
	res := mustValidate(t, dir)
	if res.Errors != 1 {
		t.Fatalf("expected 1 failing asset, got %d", res.Errors)
	}
	for _, r := range res.Results {
		if r.Name == "video.mxf" && (r.HashValid || r.SizeValid) {
			t.Errorf("expected hash and size invalid, got %+v", r)
		}
	}
}

func TestMissingFileRecordedNotFatal(t *testing.T) {
	dir := buildDCP(t, t.TempDir(), sampleAssets)
	if err := os.Remove(filepath.Join(dir, "video.mxf")); err != nil {
		t.Fatal(err)
	}
	res := mustValidate(t, dir)
	if res.Errors != 1 || len(res.Results) != 2 {
		t.Fatalf("expected 1 error and 2 results, got %+v", res)
	}
	found := false
	for _, r := range res.Results {
		if r.Name == "video.mxf" {
			found = true
			if r.Error == "" {
				t.Errorf("expected error message for missing file")
			}
		}
	}
	if !found {
		t.Errorf("missing asset not present in results")
	}
}

func TestSubdirectoryViaAssetMap(t *testing.T) {
	assets := []testAsset{{id: "urn:uuid:1111", rel: "VIDEO/video.mxf", data: "MXF-DATA"}}
	dir := buildDCP(t, t.TempDir(), assets)
	res := mustValidate(t, dir)
	assertAllValid(t, res)
	if res.Results[0].Name != "VIDEO/video.mxf" {
		t.Errorf("expected ASSETMAP path, got %q", res.Results[0].Name)
	}
}

func TestNoAssetMapFallsBackToOriginalFileName(t *testing.T) {
	dir := buildDCP(t, t.TempDir(), sampleAssets)
	if err := os.Remove(filepath.Join(dir, "ASSETMAP.xml")); err != nil {
		t.Fatal(err)
	}
	assertAllValid(t, mustValidate(t, dir))
}

func TestUppercasePKLExtension(t *testing.T) {
	dir := buildDCP(t, t.TempDir(), sampleAssets)
	if err := os.Remove(filepath.Join(dir, "PKL_test.xml")); err != nil {
		t.Fatal(err)
	}
	writePKL(t, dir, "PKL_TEST.XML", sampleAssets)
	res := mustValidate(t, dir)
	if len(res.Results) != 2 {
		t.Fatalf("PKL.XML was not validated: %+v", res)
	}
	assertAllValid(t, res)
}

func TestMultiplePKLsAllValidated(t *testing.T) {
	dir := buildDCP(t, t.TempDir(), sampleAssets)
	if err := os.Remove(filepath.Join(dir, "PKL_test.xml")); err != nil {
		t.Fatal(err)
	}
	writePKL(t, dir, "PKL_a.xml", sampleAssets[:1])
	writePKL(t, dir, "PKL_b.xml", sampleAssets[1:])
	res := mustValidate(t, dir)
	if len(res.Results) != 2 {
		t.Fatalf("expected results from both PKLs, got %d", len(res.Results))
	}
	assertAllValid(t, res)
}

func TestNoPKLIsError(t *testing.T) {
	dir := buildDCP(t, t.TempDir(), sampleAssets)
	if err := os.Remove(filepath.Join(dir, "PKL_test.xml")); err != nil {
		t.Fatal(err)
	}
	if _, err := Validate(dir); err == nil {
		t.Fatal("expected error when no PKL is present")
	}
}

func TestDirectoryNameContainingXML(t *testing.T) {
	dir := buildDCP(t, filepath.Join(t.TempDir(), "xml-dcp"), sampleAssets)
	assertAllValid(t, mustValidate(t, dir))
}

func TestPathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	assets := []testAsset{{id: "urn:uuid:1111", rel: "../secret.mxf", data: "x"}}
	writeAssetMap(t, dir, "ASSETMAP.xml", assets)
	writePKL(t, dir, "PKL_test.xml", assets)
	res := mustValidate(t, dir)
	if res.Errors != 1 || !strings.Contains(res.Results[0].Error, "escapes") {
		t.Fatalf("expected traversal to be rejected, got %+v", res)
	}
}

func TestMalformedPKLIsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "PKL_bad.xml"), "<PackingList><AssetList>")
	if _, err := Validate(dir); err == nil {
		t.Fatal("expected parse error for malformed PKL")
	}
}

func TestJSONShape(t *testing.T) {
	empty := t.TempDir()
	writePKL(t, empty, "PKL_empty.xml", nil)
	b, err := json.Marshal(mustValidate(t, empty))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"results":[]`) {
		t.Errorf("empty results should marshal as [], got %s", b)
	}

	b, err = json.Marshal(mustValidate(t, buildDCP(t, t.TempDir(), sampleAssets)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"size_valid"`) {
		t.Errorf("expected size_valid key, got %s", b)
	}
}
