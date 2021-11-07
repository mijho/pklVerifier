package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

// ValidateHandler provides functionality to validate DCP
func ValidateHandler(c *cli.Context) error {
	ec := 0
	listOfPkls, assetMapPath, err := findXmlFiles(inputValue)
	if err != nil {
		log.Fatalf("There was an error finding XML files: %s", err)
	}

	assetsArray := make([][]string, 0)

	for _, file := range listOfPkls {
		body, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("There was an error reading body of XML: %s, %s", file, body)
		}

		if strings.Contains(string(body), "<PackingList") {
			assetsArray = getAssetValues(file, assetMapPath)
		}
	}

	for _, asset := range assetsArray {
		fileToVerify := strings.Join([]string{inputValue, "/", asset[1]}, "")
		encodedHash := verifyHash(fileToVerify, asset[2])
		fileSizeString := verifySize(fileToVerify, asset[3])

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

	return nil
}
