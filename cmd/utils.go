package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func findXmlFiles(dir string) ([]string, string, error) {
	var fileList []string
	var assetMapPath string

	_ = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if strings.Contains(path, "ASSETMAP") {
			assetMapPath = path
		}

		if strings.Contains(path, "xml") {
			fileList = append(fileList, path)
		}
		return err
	})

	return fileList, assetMapPath, nil
}

func verifyHash(s string, h string) string {

	fmt.Println("\nValidating:                " + s)

	f, e := os.Open(s)
	if e != nil {
		panic(e)
	}
	defer f.Close()

	hash := sha1.New()

	_, e = io.Copy(hash, f)
	if e != nil {
		panic(e)
	}

	hashBytes := hash.Sum(nil)
	encodedHash := base64.StdEncoding.EncodeToString(hashBytes)
	return encodedHash
}

func verifySize(f string, s string) string {
	file, e := os.Stat(f)
	if e != nil {
		panic(e)
	}

	fmt.Println("The reported filesize is:  " + s)
	fmt.Printf("The actual filesize is:    %d\n", file.Size())
	fileSizeString := strconv.FormatInt(file.Size(), 10)
	return fileSizeString

}
