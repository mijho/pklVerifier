### pklVerifier

#### A DCP validation tool written in Golang

This is currently in development and shouldn't be relied upon to validate accurately at this point.

##### Installation Instructions

With go get:
```
$ go get github.com/mijho/pklVerifier
```
Or head to the releases page to download the prebuild binary for your system.

NB gobcrypt has been tested fully on OSX and Linux amd64 only, please let me know if there are any performance issues.

A windows release may be on the cards but would require some small changes first.

```
Usage of ./pklVerifier:
  -d string
    	Specify the path to the DCP directory
```
Example usage:

Verify a DCP:
```
$ pklVerifier -d /path/to/DCP 
Validating:                DCP/example_aud.mxf
The reported filesize is:  27031694
The actual filesize is:    27031694
Hash from PKL:             XXXXXXXXXXXXXXXXXXXXXXXXX+s=
Hash of file:              XXXXXXXXXXXXXXXXXXXXXXXXX+s=
Hash result:               VALID
Size result:               VALID

Validating:                DCP/CPL.xml
The reported filesize is:  230325
The actual filesize is:    230325
Hash from PKL:             XXXXXXXXXXXXXXXXXXXXXXXXX+s=
Hash of file:              XXXXXXXXXXXXXXXXXXXXXXXXX+s=
Hash result:               VALID
Size result:               VALID

The hashcheck has completed with 0 errors.
```
