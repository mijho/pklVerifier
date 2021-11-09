package main

import (
	"crypto/sha1"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
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

func verifyHash(f string) (string, error) {
	file, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	hashBytes := hash.Sum(nil)
	encodedHash := base64.StdEncoding.EncodeToString(hashBytes)

	return encodedHash, err
}

func verifySize(f string, s string) (int64, error) {
	file, err := os.Stat(f)

	if err != nil {
		return -1, err
	}

	return file.Size(), err
}
