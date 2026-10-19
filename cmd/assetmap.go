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

// readAssetMap returns a map of asset ID to relative file path.
func readAssetMap(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var am assetMap
	if err := xml.Unmarshal(data, &am); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	paths := make(map[string]string, len(am.Assets))
	for _, a := range am.Assets {
		id := strings.TrimSpace(a.ID)
		switch len(a.Chunks) {
		case 0:
			continue
		case 1:
			paths[id] = strings.TrimSpace(a.Chunks[0].Path)
		default:
			return nil, fmt.Errorf("asset %s has %d chunks; multi-chunk assets are not supported", id, len(a.Chunks))
		}
	}
	return paths, nil
}
