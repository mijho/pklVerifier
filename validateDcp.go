package main

import (
	"encoding/base64"
	"crypto/sha1"
        "fmt"
        "io"
        "os"
        "strconv"
)

func VerifyHash(s string, h string) string {

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

func VerifySize(f string, s string) string {
        
        file, e := os.Stat(f)
        if e != nil {
                panic(e)
        }

        fmt.Println("The reported filesize is:  " + s)
        fmt.Printf("The actual filesize is:    %d\n", file.Size())
        fileSizeString := strconv.FormatInt(file.Size(),10)
        return fileSizeString

}
