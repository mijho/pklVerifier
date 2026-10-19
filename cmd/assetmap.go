package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// assetMap models the parts of a SMPTE ASSETMAP that map asset IDs to paths.
type assetMap struct {
	Assets []struct {
		ID     string `xml:"Id"`
		Chunks []struct {
			Path string `xml:"Path"`
		} `xml:"ChunkList>Chunk"`
	} `xml:"AssetList>Asset"`
}

// readAssetMap returns each asset ID mapped to its ordered chunk paths.
// An asset's bytes are the concatenation of its chunks, in order.
func readAssetMap(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var am assetMap
	if err := xml.Unmarshal(data, &am); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	paths := make(map[string][]string, len(am.Assets))
	for _, a := range am.Assets {
		if len(a.Chunks) == 0 {
			continue
		}
		chunks := make([]string, 0, len(a.Chunks))
		for _, c := range a.Chunks {
			chunks = append(chunks, strings.TrimSpace(c.Path))
		}
		paths[strings.TrimSpace(a.ID)] = chunks
	}
	return paths, nil
}
