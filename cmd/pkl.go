package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

// packingList models the parts of a SMPTE PKL that validation needs.
type packingList struct {
	XMLName xml.Name   `xml:"PackingList"`
	Assets  []pklAsset `xml:"AssetList>Asset"`
}

type pklAsset struct {
	ID   string `xml:"Id"`
	Name string `xml:"OriginalFileName"`
	Hash string `xml:"Hash"`
	Size string `xml:"Size"`
}

func readPackingList(path string) ([]pklAsset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pl packingList
	if err := xml.Unmarshal(data, &pl); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return pl.Assets, nil
}
