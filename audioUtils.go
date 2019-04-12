package main

import "bytes"

func samples(lenght uint) uint {
	bytesPerSample := uint(2)
	return sampleRate * lenght * bytesPerSample
}

func trimHeader(wav []byte) []byte {
	dataIndex := bytes.Index(wav, []byte("data"))
	return wav[dataIndex:len(wav)]
}

// raw data, len in seconds, padding in samples
func trim(data []byte, lenght uint, padding uint) []byte {
	start := 0 + padding*2
	end := samples(lenght) + padding*2

	if start > uint(len(data)) {
		start = uint(len(data)) - 1
	}

	if end > uint(len(data)) {
		end = uint(len(data)) - 1
	}
	return data[start:end]
}
