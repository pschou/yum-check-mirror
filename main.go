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
	"flag"
	"fmt"
	"hash"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

var version = "test"
var debug *bool
var exitCode = 0

// Main is a function to fetch the HTTP repodata from a URL to get the latest
// package list for a repo
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Yum Check Mirror,  Version: %s\n\nUsage: %s [options...]\n\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	var basePath = flag.String("path", ".", "Path to the mirror base")
	var inRepoPath = flag.String("repo", "/", "Repo to check")
	var outputFile = flag.String("output", "-", "Output file to put the results of the check")
	var insecure = flag.Bool("insecure", false, "Skip signature checks")
	var keyringFile = flag.String("keyring", "keys/", "Use keyring for verifying, keyring.gpg or keys/ directory")
	debug = flag.Bool("debug", false, "Turn on debug, more verbose")
	flag.Parse()

	repoPath := strings.TrimSuffix(strings.TrimPrefix(*inRepoPath, "/"), "/")

	out := os.Stdout
	if *outputFile != "-" {
		f, err := os.Create(*outputFile)
		check(err)
		defer f.Close()
		out = f
	}

	var latestRepomd Repomd
	var keyring openpgp.EntityList
	if !*insecure {
		var err error
		if _, ok := isDirectory(*keyringFile); ok {
			//keyring = openpgp.EntityList{}
			for _, file := range getFiles(*keyringFile, ".gpg") {
				//fmt.Println("loading key", file)
				gpgFile := readFile(file)
				fileKeys, err := loadKeys(gpgFile)
				if err != nil {
					log.Fatal("Error loading keyring file", err)
				}
				//fmt.Println("  found", len(fileKeys), "keys")
				keyring = append(keyring, fileKeys...)
			}
		} else {
			gpgFile := readFile(*keyringFile)
			keyring, err = loadKeys(gpgFile)
			if err != nil {
				log.Fatal("Error loading keyring file", err)
			}
		}
		if len(keyring) == 0 {
			log.Fatal("no keys loaded")
		}
	}

	// Load in the repomd file.xml and parse it
	{
		repoPathSlash := path.Join(*basePath, repoPath) + "/"
		repomdPath := repoPathSlash + "repodata/repomd.xml"
		repomdPathGPG := repoPathSlash + "repodata/repomd.xml.asc"
		//log.Println("Loading", repomdPath)

		dat := readRepomdFile(repomdPath)
		if dat == nil {
			log.Fatal("Unable to open or invalid file")
		}

		if !*insecure {
			// Verify gpg signature file
			log.Println("Loading signature file:", repomdPathGPG)
			gpgFile := readFile(repomdPathGPG)
			signature_block, err := armor.Decode(strings.NewReader(gpgFile))
			if err != nil {
				log.Fatal("Unable decode signature")
			}
			p, err := packet.Read(signature_block.Body)
			if err != nil {
				log.Fatal("Unable parse signature")
			}
			var signed_at time.Time
			var issuerKeyId uint64
			var hash hash.Hash

			switch sig := p.(type) {
			case *packet.Signature:
				issuerKeyId = *sig.IssuerKeyId
				signed_at = sig.CreationTime
				if hash == nil {
					hash = sig.Hash.New()
				}
			case *packet.SignatureV3:
				issuerKeyId = sig.IssuerKeyId
				signed_at = sig.CreationTime
				if hash == nil {
					hash = sig.Hash.New()
				}
			default:
				log.Fatal("Signature block is invalid")
			}

			if issuerKeyId == 0 {
				log.Fatal("Signature doesn't have an issuer")
			}

			if keyring == nil {
				fmt.Printf("  %s - Signed by 0x%02X at %v\n", repomdPathGPG, issuerKeyId, signed_at)
				os.Exit(1)
			} else {
				fmt.Printf("Verifying %s has been signed by 0x%02X at %v...\n", repomdPathGPG, issuerKeyId, signed_at)
			}
			keys := keyring.KeysByIdUsage(issuerKeyId, packet.KeyFlagSign)

			if len(keys) == 0 {
				log.Fatal("error: No matching public key found to verify")
			}
			if len(keys) > 1 {
				fmt.Println("warning: More than one public key found matching KeyID")
			}

			dat.ascFileContents = gpgFile
			fmt.Println("GPG Verified!")
		}
		dat.path = repoPathSlash
		latestRepomd = *dat
	}

	var pkgFile string
	for _, filePath := range latestRepomd.Data {
		fileURL := path.Join(*basePath, repoPath, strings.TrimPrefix(filePath.Location.Href, "/"))
		if *debug {
			fmt.Println("checking", fileURL)
		}
		fileData := checkWithChecksum(fileURL, filePath.Checksum.Text, filePath.Checksum.Type)
		if !fileData {
			fmt.Fprintf(out, "{%s}%s %d %s\n", filePath.Checksum.Type, filePath.Checksum.Text, filePath.Size,
				path.Join(repoPath, filePath.Location.Href))
			exitCode = 1
			continue
		}
		if filePath.Type == "primary" {
			pkgFile = fileURL
		}
	}

	if pkgFile == "" {
		log.Fatal("Could not find primary file")
	}
	packages := readPackageFile(pkgFile)

	for _, pkg := range packages {
		// pkg: {XMLName:{Space:http://linux.duke.edu/metadata/common Local:package} Type:rpm Name:ImageMagick-c++ Checksum:{Text:f1599688dc9666846ae40b8c303967bac615f9d2d54c2538b3ae8c1555e169b2 Type:sha256 Pkgid:YES} Size:{Text: Package:171852 Installed:636081 Archive:637668} Location:{Text: Href:Packages/ImageMagick-c++-6.9.10.68-3.el7.x86_64.rpm}}
		fileData := checkWithChecksum(path.Join(*basePath, repoPath, pkg.Location.Href), pkg.Checksum.Text, pkg.Checksum.Type)
		if !fileData {
			fmt.Fprintf(out, "{%s}%s %s %s\n", pkg.Checksum.Type, pkg.Checksum.Text, pkg.Size.Package,
				path.Join(repoPath, pkg.Location.Href))
			exitCode = 1
			//fmt.Printf("pkg: %+v\n", pkg)
		} else if *debug {
			fmt.Println("passed", path.Join(repoPath, pkg.Location.Href))
		}
	}
	os.Exit(exitCode)
}

func check(e error) {
	if e != nil {
		//panic(e)
		log.Fatal(e)
	}
}

// isDirectory determines if a file represented
// by `path` is a directory or not
func isDirectory(path string) (exist bool, isdir bool) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, false
	}
	return true, fileInfo.IsDir()
}

func getFiles(walkdir, suffix string) []string {
	ret := []string{}
	err := filepath.Walk(walkdir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, suffix) {
				ret = append(ret, path)
			}
			return nil
		})
	if err != nil {
		log.Fatal(err)
	}
	return ret
}
