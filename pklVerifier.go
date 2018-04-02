package main

import (
	"flag"
	"fmt"
    . "github.com/logrusorgru/aurora"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func findXmlFiles(f string) ([]string, string, error) {
	searchDir := f

	fileList := make([]string, 0)
	var assetMapPath string

	e := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {

		if strings.Contains(path, "ASSETMAP") {
			assetMapPath = path
		}

		if strings.Contains(path, "xml") {
			fileList = append(fileList, path)

		}
		return err
	})
	
	if e != nil {
		panic(e)
	}

	for i, v := range fileList {
		if strings.Contains(v, "ASSETMAP") {
			assetMapPath = v
			fileList = append(fileList[:i], fileList[i+1:]...)
			break
		}
	}
	return fileList, assetMapPath, nil
}

func main() {

	var DcpDir, outFile string
	flag.StringVar(&DcpDir, "d", "", "Specify a DCP to verify")
	flag.StringVar(&outFile, "o", "", "Specify a file to write the results to")
	flag.Parse()

	ec := 0

	listOfPkls, assetMapPath, e := findXmlFiles(DcpDir)
	assetsArray := make([][]string, 0)

	for _, file := range listOfPkls {
		r, err := ioutil.ReadFile(file)
		if err != nil {
			panic(e)
		}

		s := string(r)
		if strings.Contains(s, "<PackingList") {
			assetsArray = GetAssetValues(file, assetMapPath)
		}
	}

	for _, asset := range assetsArray {
		fileToVerify := strings.Join([]string{DcpDir, "/", asset[1]}, "")
		encodedHash := VerifyHash(fileToVerify, asset[2])
		fileSizeString := VerifySize(fileToVerify, asset[3])

        fmt.Println("Hash from PKL:             " + asset[2] + "\nHash of file:              " + encodedHash)
        if asset[2] != encodedHash {
                fmt.Println("Hash result:              ", Red("NOT VALID"))
                ec++
        } else {
                fmt.Println("Hash result:              ", Green("VALID"))
        }

        if asset[3] != fileSizeString {
                fmt.Println("Size result:              ", Red("NOT VALID"))
                ec++
        } else {
                fmt.Println("Size result:              ", Green("VALID"))
        }
	}

	if ec != 0 {
		fmt.Printf("\nThe hashcheck has completed with %d errors.\n", Red(ec))
	} else {
		fmt.Printf("\nThe hashcheck has completed with %d errors.\n", Green(ec))
	}
}