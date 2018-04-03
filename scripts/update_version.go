// +build ignore

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var (
	major, minor, patch bool
	custom              string
)

func nextReleaseVersion(currentMajor, currentMinor, currentPatch string) (string, error) {
	majorVersion, err := strconv.ParseInt(currentMajor, 10, 32)
	if err != nil {
		return "", err
	}

	minorVersion, err := strconv.ParseInt(currentMinor, 10, 32)
	if err != nil {
		return "", err
	}

	patchLevel, err := strconv.ParseInt(currentPatch, 10, 32)
	if err != nil {
		return "", err
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

	return fmt.Sprintf("%v.%v.%v", majorVersion, minorVersion, patchLevel), nil
}

func main() {
	flag.BoolVar(&major, "major", false, "Update major version")
	flag.BoolVar(&minor, "minor", false, "Update minor version")
	flag.BoolVar(&patch, "patch", false, "Update patch")
	flag.StringVar(&custom, "custom", "", "Use a custom version")

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

	// Use the custom version if specified, otherwise calculate the next
	// release version.
	nextVersion := custom
	if custom == "" {
		nextVersion, err = nextReleaseVersion(match[1], match[2], match[3])
		if err != nil {
			panic(err)
		}
	}

	quotedNextVersion := fmt.Sprintf(`"%s"`, nextVersion)
	newContent := strings.Replace(content, currentVersion, quotedNextVersion, 1)

	err = ioutil.WriteFile("cmd/version.go", []byte(newContent), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", nextVersion)
}
