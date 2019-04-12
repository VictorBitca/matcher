package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const targetFiles = ".wav"
const inputSize = 180    // equal or smaller that input files
const sampleRate = 48000 // sample rate of input files

const minBitDistance = 8 // hamming distance
const minLength = 5      // min common region to be validated (seconds)

var results map[string]SearchResult

func main() {
	results = make(map[string]SearchResult)

	fmt.Printf("Search range: 0..%v seconds  \n", inputSize)

	allFiles := listAllFiles()
	pairs := pairUpFiles(allFiles)

	for _, value := range pairs {
		fmt.Print(".")
		analyse(value)
	}

	printSuccessfulResults()
	printFailedResults(allFiles)
}

func listAllFiles() []string {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	fileInfotoString := func(in os.FileInfo) string {
		return fmt.Sprintf("%s", in.Name())
	}

	mapped := fmap(files, fileInfotoString)

	filtered := filter(mapped, func(v string) bool {
		return strings.Contains(v, targetFiles)
	})

	return filtered
}

func pairUpFiles(singles []string) []WorkPair {
	results := make([]WorkPair, len(singles)-1)

	for i := 0; i < len(singles)-1; i++ {
		results[i] = WorkPair{first: singles[i], second: singles[i+1]}
	}

	return results
}
