// Written by Paul Schou (paulschou.com) March 2022
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
)

type Package struct {
	XMLName xml.Name `xml:"package"`
	//Text     string   `xml:",chardata"`
	Type     string `xml:"type,attr"`
	Name     string `xml:"name"`
	Checksum struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Pkgid string `xml:"pkgid,attr"`
	} `xml:"checksum"`
	Size struct {
		Text      string `xml:",chardata"`
		Package   string `xml:"package,attr"`
		Installed string `xml:"installed,attr"`
		Archive   string `xml:"archive,attr"`
	} `xml:"size"`
	Location struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
	} `xml:"location"`
}

type Metadata struct {
	XMLName     xml.Name  `xml:"metadata"`
	Text        string    `xml:",chardata"`
	Xmlns       string    `xml:"xmlns,attr"`
	Rpm         string    `xml:"rpm,attr"`
	Packages    int       `xml:"packages,attr"`
	PackageList []Package `xml:"package"`
}

func readPackageFile(fileName string) []Package {
	if *debug {
		log.Println("Reading in file", fileName)
	}

	// Open our xmlFile
	rawFile, err := os.Open(fileName)
	check(err)

	// Make sure the file is closed at the end of the function
	defer rawFile.Close()

	// Declare file handle for the reading
	var file io.Reader
	file = rawFile

	// Detect magic number
	b1 := make([]byte, 4)
	n1, err := file.Read(b1)
	_, err = rawFile.Seek(0, 0)

	if n1 == 4 && string(b1) == "\x1f\x8b\x08\x00" {
		if *debug {
			fmt.Println("Using gz decoder")
		}
		check(err)

		gz, err := gzip.NewReader(rawFile)
		check(err)

		// Make sure gzip handle is closed at the end of the function
		defer gz.Close()

		// Substitute the gz reader in the place of file to handle the compressed file
		file = gz
	} else {
		bufReader := bufio.NewReaderSize(file, 100000)
		file = bufReader
	}

	decoder := xml.NewDecoder(file)
	var dat Metadata
	err = decoder.Decode(&dat)
	check(err)

	if len(dat.PackageList) == 0 {
		log.Fatal("No packages found")
	}
	if len(dat.PackageList) != dat.Packages {
		log.Fatal("XML Packages count does not match the number of Packages")
	}
	return dat.PackageList
}
