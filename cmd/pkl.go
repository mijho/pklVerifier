package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type PackingList struct {
	XMLName   xml.Name `xml:"PackingList"`
	AssetList Assets   `xml:"AssetList"`
}

// our struct which contains the complete
// array of all Assets in the file
type Assets struct {
	XMLName xml.Name `xml:"AssetList"`
	Assets  []Asset  `xml:"Asset"`
}

// the Asset struct
type Asset struct {
	XMLName xml.Name `xml:"Asset"`
	Id      string   `xml:"Id"`
	Name    string   `xml:"OriginalFileName"`
	Hash    string   `xml:"Hash"`
	Size    string   `xml:"Size"`
	Type    string   `xml:"Type"`
}

func getAssetValues(s string, a string) ([]map[string]string, error) {

	xmlFile, err := os.Open(s)
	if err != nil {
		return nil, err
	}

	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)

	var assets PackingList
	var assetsArray []map[string]string
	xml.Unmarshal(byteValue, &assets)

	for i := 0; i < len(assets.AssetList.Assets); i++ {
		assetMap := make(map[string]string)
		assetMap["Id"] = assets.AssetList.Assets[i].Id
		assetMap["Name"] = assets.AssetList.Assets[i].Name
		assetMap["Hash"] = assets.AssetList.Assets[i].Hash
		assetMap["Size"] = assets.AssetList.Assets[i].Size
		assetMap["Type"] = assets.AssetList.Assets[i].Type
		assetsArray = append(assetsArray, assetMap)

	}
	return assetsArray, err
}
