package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type AssetMap struct {
	XMLName 	xml.Name `xml:"AssetMap"`
	AssetList   AssetMapAssets   `xml:"AssetList"`
}

type AssetMapAssets struct {
	XMLName 		 xml.Name `xml:"AssetList"`
	AssetMapAssets   []AssetMapAsset   `xml:"Asset"`
}

type AssetMapAsset struct {
	XMLName 			xml.Name `xml:"Asset"`
	AssetMapId			string   `xml:"Id"`
	AssetMapChunkList   []AssetMapChunkList   `xml:"ChunkList"`
}

type AssetMapChunkList struct {
	XMLName 		xml.Name `xml:"ChunkList"`
	AssetMapChunk 	[]AssetMapChunk `xml:"Chunk"`
}

type AssetMapChunk struct {
	AssetMapPath	string `xml:"Path"`
}

func GetNameFromAssetMap(s string, a string) string {
	
	xmlFile, e := os.Open(a)
	if e != nil {
		panic(e)
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)
	var assets AssetMap
	xml.Unmarshal(byteValue, &assets)

	for i := 0; i < len(assets.AssetList.AssetMapAssets); i++ {
		if s == assets.AssetList.AssetMapAssets[i].AssetMapId {
			return assets.AssetList.AssetMapAssets[i].AssetMapChunkList[0].AssetMapChunk[0].AssetMapPath
		}
	}
	return "Not found"
}