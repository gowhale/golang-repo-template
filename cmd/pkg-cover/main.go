// Package main contains code to do with ensuring coverage is over 80%
package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

const (
	minPercentCov = 80.0

	coverageStringNotFound = -1
	firstItemIndex         = 1
	floatByteSize          = 64
	emptySliceLen          = 0
	lenOfPercentChar       = 1
	indexOfEmptyLine       = 1
)

var excludedPkgs = map[string]bool{
	"go-shopping-list":                  true,
	"go-shopping-list/cmd/pkg-cover":    true,
	"go-shopping-list/cmd/authenticate": true,
	"go-shopping-list/pkg/common":       true,
	"go-shopping-list/pkg/fruit":        true,
}

func main() {
	if err := execute(); err != nil {
		log.Fatalln(err)
	}
	log.Println("Tests=PASS Coverage=PASS")
}

func execute() error {
	output, err := runGoTest()
	if err != nil {
		log.Println(output)
		return err
	}

	tl, err := covertOutputToCoverage(output)
	if err != nil {
		return err
	}

	return validateTestOutput(tl, output)
}

var execCommand = exec.Command

func runGoTest() (string, error) {
	cmd := execCommand("go", "test", "./...", "--cover")
	output, err := cmd.CombinedOutput()
	termOutput := string(output)
	return termOutput, err
}

func getCoverage(pkgName, line string) (bool, float64, error) {
	if _, ok := excludedPkgs[pkgName]; !ok {
		coverageIndex := strings.Index(line, "coverage: ")
		if coverageIndex != coverageStringNotFound {
			lineFields := strings.Fields(line[coverageIndex:])
			pkgPercentStr := lineFields[firstItemIndex][:len(lineFields[firstItemIndex])-lenOfPercentChar]
			pkgPercentFloat, err := strconv.ParseFloat(pkgPercentStr, floatByteSize)
			if err != nil {
				return false, coverageStringNotFound, err
			}
			return true, pkgPercentFloat, nil
		}
		return true, coverageStringNotFound, nil
	}
	return false, coverageStringNotFound, nil
}

type testLine struct {
	pkgName  string
	coverage float64
}

func covertOutputToCoverage(termOutput string) ([]testLine, error) {
	testStruct := []testLine{}
	lines := strings.Split(termOutput, "\n")
	for _, line := range lines[:len(lines)-indexOfEmptyLine] {
		if !strings.Contains(line, "go: downloading") {
			pkgName := strings.Fields(line)[firstItemIndex]
			covLine, covVal, err := getCoverage(pkgName, line)
			if err != nil {
				return nil, err
			}
			if covLine {
				testStruct = append(testStruct, testLine{pkgName: pkgName, coverage: covVal})
			}
		}
	}
	return testStruct, nil
}

func validateTestOutput(tl []testLine, o string) error {
	invalidOutputs := []string{}
	for _, line := range tl {
		switch {
		case line.coverage == coverageStringNotFound:
			invalidOutputs = append(invalidOutputs, fmt.Sprintf("pkg=%s is missing tests", line.pkgName))
		case line.coverage < minPercentCov:
			invalidOutputs = append(invalidOutputs, fmt.Sprintf("pkg=%s cov=%f under the %f%% minimum line coverage", line.pkgName, line.coverage, minPercentCov))
		}
	}
	if len(invalidOutputs) == emptySliceLen {
		return nil
	}
	log.Println(o)
	log.Println("###############################")
	log.Println("###############################")
	log.Println("invalid pkg's:")
	for i, invalid := range invalidOutputs {
		log.Printf("id=%d problem=%s", i, invalid)
	}
	return fmt.Errorf("the following pkgs are not valid: %+v", invalidOutputs)
}
