package main

import (
	"flag"
	"fmt"
    . "github.com/logrusorgru/aurora"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func validateFiles(s []string, DcpDir string) {
	ec := 0
	asset := s
	fileToVerify := strings.Join([]string{DcpDir, "/", asset[1]}, "")
	encodedHash := VerifyHash(fileToVerify, asset[2])
	fileSizeString := VerifySize(fileToVerify, asset[3])

    fmt.Println("\nValidating:                " + asset[1])
    fmt.Println("The reported filesize is:  " + asset[3] + "\nThe actual filesize is:    " + fileSizeString)
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

	if ec != 0 {
		fmt.Printf("\nThe hashcheck has completed with %d errors.\n", Red(ec))
	} else {
		fmt.Printf("\nThe hashcheck has completed with %d errors.\n", Green(ec))
	}
}

func main() {

	var DcpDir, outFile string
	flag.StringVar(&DcpDir, "d", "", "Specify a DCP to verify")
	flag.StringVar(&outFile, "o", "", "Specify a file to write the results to")
	flag.Parse()

	// ec := 0

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
	var wg sync.WaitGroup
	// var i int = -1
	// var asset []string
	for _, asset := range assetsArray {
		wg.Add(1)
		go func (a []string) {
			validateFiles(a, DcpDir)
			wg.Done()
		} (asset)
		// fileToVerify := strings.Join([]string{DcpDir, "/", asset[1]}, "")
		// encodedHash := VerifyHash(fileToVerify, asset[2])
		// fileSizeString := VerifySize(fileToVerify, asset[3])

  //       fmt.Println("Hash from PKL:             " + asset[2] + "\nHash of file:              " + encodedHash)
  //       if asset[2] != encodedHash {
  //               fmt.Println("Hash result:              ", Red("NOT VALID"))
  //               ec++
  //       } else {
  //               fmt.Println("Hash result:              ", Green("VALID"))
  //       }

  //       if asset[3] != fileSizeString {
  //               fmt.Println("Size result:              ", Red("NOT VALID"))
  //               ec++
  //       } else {
  //               fmt.Println("Size result:              ", Green("VALID"))
  //       }
	}
	wg.Wait()

	// if ec != 0 {
	// 	fmt.Printf("\nThe hashcheck has completed with %d errors.\n", Red(ec))
	// } else {
	// 	fmt.Printf("\nThe hashcheck has completed with %d errors.\n", Green(ec))
	// }
}