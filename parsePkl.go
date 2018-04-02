package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type PackingList struct {
	XMLName xml.Name `xml:"PackingList"`
	AssetList   Assets   `xml:"AssetList"`
}

// our struct which contains the complete
// array of all Assets in the file
type Assets struct {
	XMLName xml.Name `xml:"AssetList"`
	Assets   []Asset   `xml:"Asset"`
}

// the Asset struct, this contains our
// Type attribute, our user's name and
// a social struct which will contain all
// our social links
type Asset struct {
	XMLName xml.Name `xml:"Asset"`
	Id		string   `xml:"Id"`
	Name    string   `xml:"OriginalFileName"`
	Hash	string	 `xml:"Hash"`
	Size	string	 `xml:"Size"`
	Type    string   `xml:"Type"`
}

func GetAssetValues(s string, a string) [][]string {

	xmlFile, e := os.Open(s)
	if e != nil {
		panic(e)
	}

	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)

	var assets PackingList
	xml.Unmarshal(byteValue, &assets)
	assetsArray := make([][]string, 0)

	for i := 0; i < len(assets.AssetList.Assets); i++ {
		assetArray := make([]string, 0)
		assetArray = append(assetArray, assets.AssetList.Assets[i].Id)
		if assets.AssetList.Assets[i].Name != "" {
			assetArray = append(assetArray, assets.AssetList.Assets[i].Name)
		} else {
			assetArray = append(assetArray, GetNameFromAssetMap(assets.AssetList.Assets[i].Id, a))
		}
		assetArray = append(assetArray, assets.AssetList.Assets[i].Hash)
		assetArray = append(assetArray, assets.AssetList.Assets[i].Size)
		assetArray = append(assetArray, assets.AssetList.Assets[i].Type)

		assetsArray = append(assetsArray, assetArray)
	}
	return assetsArray
}