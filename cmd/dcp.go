package main

import (
	"fmt"
	"path/filepath"
)

// dcpRoot is one DCP: a directory holding an ASSETMAP, plus the PKLs and
// files beneath it that do not belong to a nested DCP.
type dcpRoot struct {
	dir      string
	assetMap string // "" when the input has no ASSETMAP at all
	pkls     []string
	files    []string
}

// groupDCPs assigns every PKL and file to the deepest DCP directory that
// contains it. A film folder holding one or more DCP subdirectories therefore
// yields one dcpRoot per DCP, each validated against its own ASSETMAP.
func groupDCPs(input string, f dcpFiles) ([]*dcpRoot, error) {
	var roots []*dcpRoot
	seen := make(map[string]bool)
	for _, am := range f.assetMaps {
		d := filepath.Dir(am)
		if seen[d] {
			return nil, fmt.Errorf("multiple ASSETMAP files found in %s", d)
		}
		seen[d] = true
		roots = append(roots, &dcpRoot{dir: d, assetMap: am})
	}
	if len(roots) == 0 {
		roots = append(roots, &dcpRoot{dir: input})
	}

	owner := func(p string) *dcpRoot {
		var best *dcpRoot
		for _, r := range roots {
			if within(r.dir, p) && (best == nil || len(r.dir) > len(best.dir)) {
				best = r
			}
		}
		return best
	}

	for _, p := range f.packingLists {
		r := owner(p)
		if r == nil {
			return nil, fmt.Errorf("PKL %s is not inside any DCP", p)
		}
		r.pkls = append(r.pkls, p)
	}
	for _, p := range f.regular {
		if r := owner(p); r != nil {
			r.files = append(r.files, p)
		}
	}
	for _, r := range roots {
		if len(r.pkls) == 0 {
			return nil, fmt.Errorf("no PackingList (PKL) found in DCP %s", r.dir)
		}
	}
	return roots, nil
}

// displayName prefixes rel with the DCP's location relative to the input, so
// results from nested DCPs stay distinguishable. DCPs at the input root are unprefixed.
func displayName(input, dcpDir, rel string) string {
	p := relSlash(input, dcpDir)
	if p == "." {
		return rel
	}
	return p + "/" + rel
}
