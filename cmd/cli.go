package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	. "github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"
)

// Build time variables
var (
	Build   string
	Commit  string
	Name    string
	Version string
	Author  string
	Email   string
)

var inputValue string
var inputFlag = &cli.StringFlag{
	Name:        "input",
	Aliases:     []string{"i"},
	Usage:       "Specify the DCP to validate",
	Required:    false,
	Destination: &inputValue,
}

var outputValue string
var outputFlag = &cli.StringFlag{
	Name:        "output",
	Aliases:     []string{"o"},
	Usage:       "Specify a file to write out the results to",
	Required:    false,
	Destination: &outputValue,
}

// CLIApp constructs the cli applications
var CLIApp = &cli.App{
	Name:        "gobcrypt",
	Usage:       "A DCP Validator",
	ArgsUsage:   "",
	Version:     Version + "-" + Commit,
	Description: fmt.Sprintf("%s: Build: %s", Name, Build),
	Commands: []*cli.Command{
		{
			Name:  "validate",
			Usage: "validate provided DCP",
			Flags: []cli.Flag{
				inputFlag,
				outputFlag,
			},
			Action: ValidateHandler,
		},
	},
	Authors: []*cli.Author{
		{
			Name:  Author,
			Email: Email,
		},
	},
}

type Result struct {
	Name         string `json:"name"`
	ReportedSize string `json:"reported_size"`
	ActualSize   int64  `json:"actual_size"`
	ReportedHash string `json:"reported_hash"`
	ActualHash   string `json:"actual_hash"`
	HashValid    bool   `json:"hash_valid"`
	SizeValid    bool   `json:"size_vaild`
}

// ValidateHandler provides functionality to validate DCP
func ValidateHandler(c *cli.Context) error {
	ec := 0
	listOfPkls, assetMapPath, err := findXmlFiles(inputValue)
	if err != nil {
		log.Fatalf("There was an error finding XML files: %s\n", err)
	}

	var assetsArray []map[string]string

	for _, file := range listOfPkls {
		body, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("There was an error reading body of XML: %s, %s\n", file, body)
		}

		if strings.Contains(string(body), "<PackingList") {
			assetsArray, err = getAssetValues(file, assetMapPath)
			if err != nil {
				log.Fatalf("There was an error retieving asset values: %s\n", err)
			}
		}
	}

	for _, asset := range assetsArray {
		var hashValid, sizeValid bool
		fileToVerify := strings.Join([]string{inputValue, "/", asset["Name"]}, "")

		encodedHash, err := verifyHash(fileToVerify)
		if err != nil {
			log.Fatalf("There was an error hashing %s: %s \n", fileToVerify, err)
		}

		if asset["Hash"] != encodedHash {
			hashValid = false
			ec++
		} else {
			hashValid = true
		}

		fileSize, err := verifySize(fileToVerify, asset["Size"])
		if err != nil {
			log.Fatalf("There was an error getting file size %s: %s \n", fileToVerify, err)
		}

		if asset["Size"] != strconv.FormatInt(fileSize, 10) {
			sizeValid = false
			ec++
		} else {
			sizeValid = true
		}

		result := &Result{
			Name:         fileToVerify,
			ReportedSize: asset["Size"],
			ActualSize:   fileSize,
			ReportedHash: asset["Hash"],
			ActualHash:   encodedHash,
			HashValid:    hashValid,
			SizeValid:    sizeValid,
		}

		b, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("error marshalling json: %v", err)
		}
		fmt.Println(string(b))
	}

	// TODO: create struct to marshal individual results in and nest overall results
	if ec != 0 {
		fmt.Printf("\nThe hashcheck has completed with %d errors.\n", Red(ec))
	} else {
		fmt.Printf("\nThe hashcheck has completed with %d errors.\n", Green(ec))
	}

	return nil
}
