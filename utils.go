package main

import "os"

func fmap(vs []os.FileInfo, f func(os.FileInfo) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func clip(val int, min int, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func getTheBiggestIndex(arr []float64) int {
	index := 0
	biggest := 0.0

	for i := 0; i < len(arr); i++ {
		val := arr[i]
		if val > biggest {
			biggest = val
			index = i
		}
	}

	return index
}
