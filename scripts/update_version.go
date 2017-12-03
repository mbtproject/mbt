// +build ignore

package main

import (
	"strings"
	"strconv"
	"fmt"
	"io/ioutil"
	"regexp"
	"flag"
)

var major, minor, patch bool

func main() {
	flag.BoolVar(&major, "major", false, "Update major version")
	flag.BoolVar(&minor, "minor", false, "Update minor version")
	flag.BoolVar(&patch, "patch", false, "Update patch")

	flag.Parse()

	contentBytes, err := ioutil.ReadFile("cmd/version.go")
	if err != nil {
		panic(err)
	}

	content := string(contentBytes)

	versionPattern := regexp.MustCompile(`"(?P<Major>\d*?)\.(?P<Minor>\d*?)\.(?P<Patch>\d*?)"`)

	match := versionPattern.FindStringSubmatch(content)
	if len(match) != 4 {
		panic("Unable to find the version string in source")
	}

	currentVersion := match[0]
	
	majorVersion, err := strconv.ParseInt(match[1], 10, 32)
	if err != nil {
		panic(err)
	}

	minorVersion, err := strconv.ParseInt(match[2], 10, 32)
	if err != nil {
		panic(err)
	}

	patchLevel, err := strconv.ParseInt(match[3], 10, 32)
	if err != nil {
		panic(err)
	}

	if major {
		majorVersion++
		minorVersion = 0
		patchLevel = 0
	} else if minor {
		minorVersion++
		patchLevel = 0
	} else if patch {
		patchLevel++
	}

	nextVersion := fmt.Sprintf(`"%v.%v.%v"`, majorVersion, minorVersion, patchLevel)
	newContent := strings.Replace(content, currentVersion, nextVersion, 1)

	err = ioutil.WriteFile("cmd/version.go", []byte(newContent), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v.%v.%v\n", majorVersion, minorVersion, patchLevel)
}
