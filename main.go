package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const targetFiles = ".wav"
const inputSize = 180    // equal or smaller that input files
const sampleRate = 48000 // sample rate of input files

const minBitDistance = 8 // hamming distance
const minLength = 5      // min common region to be validated (seconds)

const maxWorkers = 5 // max paralel workers

var results map[string]SearchResult
var start time.Time

var jobs chan WorkPair
var output chan SearchResult

func init() {
	start = time.Now()
}

func main() {
	jobs = make(chan WorkPair, 30)
	output = make(chan SearchResult, 30)
	results = make(map[string]SearchResult)

	fmt.Printf("Search range: 0..%v seconds  \n", inputSize)

	allFiles := listAllFiles()
	pairs := pairUpFiles(allFiles)

	for index := 0; index < maxWorkers; index++ {
		go analyse(jobs, output)
	}

	for _, pair := range pairs {
		fmt.Print(".")
		jobs <- pair
	}
	close(jobs)

	for r := 0; r < len(allFiles)*2-2; r++ {
		saveResult(<-output)
	}

	printSuccessfulResults()
	printFailedResults(allFiles)
	fmt.Println("Finished in: ", time.Since(start))
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
