package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-fingerprint/fingerprint"
	"github.com/go-fingerprint/gochroma"
)

var (
	file1, file2 string
)

const inputSize = 300
const sampleRate = 48000
const bytesPerSample = 2 //16 bit depth

const windowSize uint = 10                         //incremets for first loop
const maxWindowIndex = inputSize / int(windowSize) // upper constraint for first loop

const windowSizeInSamples = 2000                                                                                                 //increments for second loop
const maxSampleIndex = ((inputSize * sampleRate) / windowSizeInSamples) - ((int(windowSize) * sampleRate) / windowSizeInSamples) //uppper constraint for second loop

func samples(lenght uint) uint {
	return sampleRate * lenght * bytesPerSample
}

func trimHeader(wav []byte) []byte {
	dataIndex := bytes.Index(wav, []byte("data"))
	fmt.Printf("wav header ended at: %v \n", dataIndex)
	return wav[dataIndex:len(wav)]
}

// raw data, len in seconds, padding in samples
func trim(data []byte, lenght uint, padding uint) []byte {
	return data[0+padding*2 : samples(lenght)+padding*2]
}

func searchIntro2(audio1 []byte, audio2 []byte, fpcalc *gochroma.Printer) {
	// Search for Intro
	// Trim byte slices
	maxScore := 0.0
	sampeIndexOfMaxScore := 0

	for windowIndex := 0; windowIndex < maxWindowIndex; windowIndex++ {
		//windowSize * windowIndex = something in seconds
		audio1OffsetSeconds := (int(windowSize) * windowIndex)
		audio1OffsetSamples := audio1OffsetSeconds * sampleRate
		trimmedData1 := trim(audio1, windowSize, uint(audio1OffsetSamples))
		r1 := bytes.NewReader(trimmedData1)

		// Get fingerprints as a slices of 32-bit integers
		fprint1, err := fpcalc.RawFingerprint(
			fingerprint.RawInfo{
				Src:        r1,
				Channels:   1,
				Rate:       sampleRate,
				MaxSeconds: windowSize,
			})

		if err != nil {
			log.Fatal(err)
		}

		for sampleIndex := 0; sampleIndex < maxSampleIndex; sampleIndex++ {
			audio2OffsetSamples := sampleIndex * windowSizeInSamples
			trimmedData2 := trim(audio2, windowSize, uint(audio2OffsetSamples))
			r2 := bytes.NewReader(trimmedData2)

			// Get fingerprints as a slices of 32-bit integers
			fprint2, err := fpcalc.RawFingerprint(
				fingerprint.RawInfo{
					Src:        r2,
					Channels:   1,
					Rate:       sampleRate,
					MaxSeconds: windowSize,
				})
			if err != nil {
				log.Fatal(err)
			}

			// Compare fingerprints
			s, err := fingerprint.Compare(fprint1, fprint2)
			if err != nil {
				log.Fatal(err)
			}

			if maxScore < s {
				maxScore = s
				sampeIndexOfMaxScore = audio2OffsetSamples
			}
		}
		fmt.Printf("Working... iteration: %v out of: %v \n", windowIndex, maxWindowIndex)
	}

	if maxScore > 0.85 {
		fmt.Printf("Found match at: %d, score is: %v \n", sampeIndexOfMaxScore, maxScore)
	} else {
		fmt.Printf("Fingerprints differ a lot, with max score: %v \n", maxScore)
	}

	/*
		// Get graphical representation of distance between fingerprints
		i, err := fingerprint.ImageDistance(fprint1, fprint2)
		if err != nil {
			log.Fatal(err)
		}

		out, err := os.Create(filepath.Join(filepath.Dir(flag.Arg(0)), "out.png"))
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(out, i); err != nil {
			log.Fatal(nil)
		}
	*/
}

func main() {
	fmt.Printf("input len sec: %v, sampleRate: %v, outerLoopIncremets: %v, outerLoopLimit: %v, innerLoopIncrements: %v, innerLoopLimit: %v\n", inputSize, sampleRate, windowSize, maxWindowIndex, windowSizeInSamples, maxSampleIndex)
	// Read in the whole files into byte slice
	flag.Parse()

	if flag.NArg() < 2 {
		println("Usage: compare <file1> <file2>")
		os.Exit(0)
	}

	audio1, err := ioutil.ReadFile(flag.Arg(0)) //os.Open(flag.Arg(0))

	if err != nil {
		log.Fatal(err)
	}

	audio2, err := ioutil.ReadFile(flag.Arg(1)) //os.Open(flag.Arg(1))

	if err != nil {
		log.Fatal(err)
	}

	// Create new fingerprint calculator
	fpcalc := gochroma.New(gochroma.AlgorithmDefault)
	defer fpcalc.Close()

	fmt.Printf("Samples total %v \n", len(trimHeader(audio1))/2)

	searchIntro2(trimHeader(audio1), trimHeader(audio2), fpcalc)
}

/*
func searchIntro(audio1 []byte, audio2 []byte, fpcalc *gochroma.Printer) {
	// Search for Intro
	for i := 3; i < 30; i++ {
		// Trim byte slices
		trimmedData1 := trim(audio1, i)
		trimmedData2 := trim(audio2, i)

		r1 := bytes.NewReader(trimmedData1)
		r2 := bytes.NewReader(trimmedData2)

		// Get fingerprints as a slices of 32-bit integers
		fprint1, err := fpcalc.RawFingerprint(
			fingerprint.RawInfo{
				Src:        r1,
				Channels:   1,
				Rate:       44100,
				MaxSeconds: 120,
			})

		if err != nil {
			log.Fatal(err)
		}

		fprint2, err := fpcalc.RawFingerprint(
			fingerprint.RawInfo{
				Src:        r2,
				Channels:   1,
				Rate:       44100,
				MaxSeconds: 120,
			})

		if err != nil {
			log.Fatal(err)
		}

		// Compare fingerprints
		s, err := fingerprint.Compare(fprint1, fprint2)

		if err != nil {
			log.Fatal(err)
		}

		if s > 0.99 {
			fmt.Printf("Fingerprints do not differ a lot at: %d, score is: %v \n", i, s)
		} else {
			fmt.Printf("Fingerprints differ a lot, at: %d, with a score: %v \n", i, s)
			break
		}
	}
}
*/
