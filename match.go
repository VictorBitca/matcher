package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sort"

	"github.com/go-fingerprint/fingerprint"
	"github.com/go-fingerprint/gochroma"
	"github.com/steakknife/hamming"
)

type SearchResult struct {
	name       string
	start, end float64
}

type WorkPair struct {
	first, second string
}

// Analyses the input then writes the results in the global result variable
func analyse(pair WorkPair) {
	audio1, err1 := ioutil.ReadFile(pair.first)
	audio2, err2 := ioutil.ReadFile(pair.second)

	if err1 != nil || err2 != nil {
		log.Fatal(err1, err2)
	}

	fpcalc := gochroma.New(gochroma.AlgorithmDefault)
	defer fpcalc.Close()

	p1start, p1end, p2start, p2end, e := searchIntro(trimHeader(audio1), trimHeader(audio2), fpcalc)

	if e == nil {
		r1 := SearchResult{name: pair.first, start: p1start, end: p1end}
		r2 := SearchResult{name: pair.second, start: p2start, end: p2end}

		saveResult(r1)
		saveResult(r2)
	}
}

func searchIntro(audio1 []byte, audio2 []byte, fpcalc *gochroma.Printer) (float64, float64, float64, float64, error) {
	// Trim byte slices
	trimmedData1 := trim(audio1, inputSize, 0)
	trimmedData2 := trim(audio2, inputSize, 0)

	r1 := bytes.NewReader(trimmedData1)
	r2 := bytes.NewReader(trimmedData2)

	// Get fingerprints as a slices of 32-bit integers
	fprint1, err1 := getFingerprint(r1, fpcalc)

	fprint2, err2 := getFingerprint(r2, fpcalc)
	if err1 != nil || err2 != nil {
		log.Fatal(err1, err2)
	}

	// If lenghts/2 are not whole numbers trim them down a bit
	if len(fprint1)%2 != 0 {
		fprint1 = fprint1[0 : len(fprint1)-1]
		fprint2 = fprint2[0 : len(fprint2)-1]
	}

	// Get the offset with best score
	offset := getBestOffset(fprint1, fprint2)
	f1, f2 := getAllingedFingerprints(offset, fprint1, fprint2)
	hammed := hammItUp(f1, f2)

	// Find the contigious region
	start, end := findContiguousRegion(hammed, minBitDistance)
	if start < 0 || end < 0 {
		return 0.0, 0.0, 0.0, 0.0, errors.New("No common regions found")
	}

	//Convert everything to seconds
	secondsPerSample := float64(inputSize) / float64(len(fprint1))
	offsetInSeconds := float64(offset) * secondsPerSample
	commonRegionStart := float64(start) * secondsPerSample
	commonRegionEnd := float64(end) * secondsPerSample

	firstFileRegionStart := 0.0
	firstFileRegionEnd := 0.0

	secondFileRegionStart := 0.0
	secondFileRegionEnd := 0.0

	if offset >= 0 {
		firstFileRegionStart = commonRegionStart + offsetInSeconds
		firstFileRegionEnd = commonRegionEnd + offsetInSeconds

		secondFileRegionStart = commonRegionStart
		secondFileRegionEnd = commonRegionEnd
	} else {
		firstFileRegionStart = commonRegionStart
		firstFileRegionEnd = commonRegionEnd

		secondFileRegionStart = commonRegionStart - offsetInSeconds
		secondFileRegionEnd = commonRegionEnd - offsetInSeconds
	}

	// Check if the found region is bigger than min length
	if firstFileRegionEnd-firstFileRegionStart < minLength {
		return 0.0, 0.0, 0.0, 0.0, errors.New("No significant common regions found")
	}

	return firstFileRegionStart, firstFileRegionEnd, secondFileRegionStart, secondFileRegionEnd, nil
}

// Get fingerprints as a slices of 32-bit integers
func getFingerprint(r *bytes.Reader, fpcalc *gochroma.Printer) ([]int32, error) {
	return fpcalc.RawFingerprint(
		fingerprint.RawInfo{
			Src:        r,
			Channels:   1,
			Rate:       sampleRate,
			MaxSeconds: inputSize,
		})
}

// Returns the offset at wich audio aligns the best
// Value is positive if first audio is late
func getBestOffset(f1 []int32, f2 []int32) int {
	len := len(f1)
	iterations := len + 1 //one for the middle ground, 0 index

	diff := (len / 2) - 1

	a := (len / 2)
	b := (len) - 1
	x := 0
	y := (len / 2) - 1

	output := make([]float64, iterations)

	for i := 0; i < iterations; i++ {
		upper := abs(a - b)
		output[i] = getMatchScore(f1[a:a+upper], f2[x:x+upper])

		a = clip(a-1, 0, len-1)

		bVal := func() int {
			if diff < 0 {
				return b - 1
			}
			return b
		}
		b = clip(bVal(), 0, len-1)

		xVal := func() int {
			if diff < 0 {
				return x + 1
			}
			return x
		}
		x = clip(xVal(), 0, len-1)

		yVal := func() int {
			if diff >= 0 {
				return y + 1
			}
			return y
		}
		y = clip(yVal(), 0, len-1)

		diff--
	}

	index := getTheBiggestIndex(output)
	return (iterations-1)/2 - index
}

// Compares two fingerprints
func getMatchScore(f1 []int32, f2 []int32) float64 {
	s, err := fingerprint.Compare(f1, f2)

	if err != nil {
		return 0.0
	}

	return s
}

//Returns the trimmed arrays so the fingerprints data lines up
func getAllingedFingerprints(offset int, f1 []int32, f2 []int32) ([]int32, []int32) {
	if offset >= 0 {
		offsetCorrectedF1 := f1[offset:len(f1)]
		offsetCorrectedF2 := f2[0 : len(f2)-offset]
		return offsetCorrectedF1, offsetCorrectedF2
	}

	offsetCorrectedF1 := f1[0 : len(f1)-abs(offset)]
	offsetCorrectedF2 := f2[abs(offset):len(f2)]
	return offsetCorrectedF1, offsetCorrectedF2
}

func hammItUp(f1 []int32, f2 []int32) []int {
	result := make([]int, len(f1))
	for i := 0; i < len(f1); i++ {
		result[i] = hamming.Int32(f1[i], f2[i])
	}

	return result
}

func findContiguousRegion(arr []int, upperLimit int) (int, int) {
	start := -1
	end := -1

	for i := 0; i < len(arr); i++ {
		if arr[i] < upperLimit && nextOnesAreAlsoSmall(arr, i, upperLimit) {
			if start == -1 {
				start = i
			}
			end = i
		}
	}

	return start, end
}

func nextOnesAreAlsoSmall(arr []int, index int, upperLimit int) bool {
	if index+3 < len(arr) {
		v1 := arr[index+1]
		v2 := arr[index+2]
		v3 := arr[index+3]
		average := (v1 + v2 + v3) / 3

		if average < upperLimit {
			return true
		}

		return false
	}

	return false
}

// Writes the result in the global resuls variable
func saveResult(result SearchResult) {
	if results[result.name].name != "" {
		if results[result.name].start > result.start {
			results[result.name] = SearchResult{name: result.name, start: result.start, end: results[result.name].end}
		}
		if results[result.name].end < result.end {
			results[result.name] = SearchResult{name: result.name, start: results[result.name].start, end: result.end}
		}
	} else {
		results[result.name] = result
	}
}

func printSuccessfulResults() {
	fmt.Println("\n \n Found common regions in:")

	sortedKeys := make([]string, len(results))

	ii := 0
	for key := range results {
		sortedKeys[ii] = key
		ii++
	}
	sort.Strings(sortedKeys)

	for i := range sortedKeys {
		value := results[sortedKeys[i]]
		fmt.Printf("%s intro starts at: %.1f ends at: %.1f\n", value.name, value.start, value.end)
	}
}

func printFailedResults(allFiles []string) {
	fmt.Printf("\nNo significant results found in range: 0..%v in: \n", inputSize)

	var allFailedResults []string

	for i := range allFiles {
		key := allFiles[i]
		if (results[key] == SearchResult{name: "", start: 0, end: 0}) {
			allFailedResults = append(allFailedResults, key)
		}
	}

	for i := range allFailedResults {
		fmt.Println(allFailedResults[i])
	}
}
