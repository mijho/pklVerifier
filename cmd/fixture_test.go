package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureArchive is a SMPTE 2K DCP made by DCP-o-matic 2.18.51 (see testdata/generate.sh).
const fixtureArchive = "../testdata/dcpomatic-2.18.51-smpte-2k-ov.tar.gz"

// extractFixture unpacks the real DCP into a fresh directory and returns the DCP path.
func extractFixture(t *testing.T) string {
	t.Helper()
	f, err := os.Open(fixtureArchive)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	root := t.TempDir()
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(hdr.Name, "..") || !(hdr.FileInfo().Mode().IsRegular() || hdr.FileInfo().IsDir()) {
			t.Fatalf("unsafe entry in fixture: %s", hdr.Name)
		}
		dst := filepath.Join(root, hdr.Name)
		if hdr.FileInfo().IsDir() {
			if err := os.MkdirAll(dst, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		out, err := os.Create(dst)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(out, tr); err != nil {
			t.Fatal(err)
		}
		out.Close()
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one DCP directory in fixture, got %v (%v)", entries, err)
	}
	return filepath.Join(root, entries[0].Name())
}

func TestRealDCPFixtureValid(t *testing.T) {
	res := mustValidate(t, extractFixture(t))
	if len(res.Results) != 3 {
		t.Fatalf("expected CPL, picture and sound, got %d results", len(res.Results))
	}
	assertAllValid(t, res)
	if len(res.Unlisted) != 0 {
		t.Errorf("unexpected unlisted files: %v", res.Unlisted)
	}
}

func TestRealDCPFixtureFilmFolder(t *testing.T) {
	film := t.TempDir()
	dcp := extractFixture(t)
	if err := os.Rename(dcp, filepath.Join(film, filepath.Base(dcp))); err != nil {
		t.Fatal(err)
	}
	res := mustValidate(t, film)
	assertAllValid(t, res)
	for _, r := range res.Results {
		if !strings.HasPrefix(r.Name, filepath.Base(dcp)+"/") {
			t.Errorf("expected nested name, got %q", r.Name)
		}
	}
}

func TestRealDCPFixtureTamperedPicture(t *testing.T) {
	dcp := extractFixture(t)
	mxf, err := filepath.Glob(filepath.Join(dcp, "j2c_*.mxf"))
	if err != nil || len(mxf) != 1 {
		t.Fatalf("picture track not found: %v", err)
	}
	b, err := os.ReadFile(mxf[0])
	if err != nil {
		t.Fatal(err)
	}
	b[len(b)/2] ^= 0xFF // same size, different content
	if err := os.WriteFile(mxf[0], b, 0o644); err != nil {
		t.Fatal(err)
	}
	res := mustValidate(t, dcp)
	if res.Errors != 1 {
		t.Fatalf("expected 1 error, got %d", res.Errors)
	}
	for _, r := range res.Results {
		if strings.HasPrefix(r.Name, "j2c_") && (r.HashValid || !r.SizeValid) {
			t.Errorf("expected hash invalid and size valid, got %+v", r)
		}
	}
}
