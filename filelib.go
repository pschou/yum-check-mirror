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
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

func readMirrors(mirrorFile string) []string {
	file, err := os.Open(mirrorFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var line string

	ret := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		line = strings.TrimSuffix(line, "/")
		ret = append(ret, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ret
}

func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		// check that the existing path is a directory
		info, err := os.Stat(dirName)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("path exists but is not a directory")
		}
		return nil
	}
	return err
}

func readFile(filePath string) string {
	// Declare file handle for the reading
	var file io.Reader

	if _, err := os.Stat(filePath); err == nil {
		log.Println("Reading in file", filePath)

		// Open our xmlFile
		rawFile, err := os.Open(filePath)
		if err != nil {
			log.Println("Error in HTTP get request", err)
			return ""
		}

		// Make sure the file is closed at the end of the function
		defer rawFile.Close()
		file = rawFile
	} else if strings.HasPrefix(filePath, "http") {
		resp, err := client.Get(filePath)
		if err != nil {
			log.Println("Error in HTTP get request", err)
			return ""
		}

		defer resp.Body.Close()
		file = resp.Body
	} else if _, err := os.Stat(filePath); err == nil {
		log.Println("Reading in file", filePath)

		// Open our xmlFile
		rawFile, err := os.Open(filePath)
		if err != nil {
			log.Println("Error opening file locally", err)
			return ""
		}

		// Make sure the file is closed at the end of the function
		defer rawFile.Close()
		file = rawFile
	} else {
		log.Println("Unable to open file", filePath)
		return ""
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	return buf.String()
}
