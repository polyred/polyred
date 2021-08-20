package main

import "testing"

type S [12]int64

var sX = make([]S, 1000)
var sY = make([]S, 1000)
var sZ = make([]S, 1000)
var sumX, sumY, sumZ int64

func Benchmark_Loop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sumX = 0
		for j := 0; j < len(sX); j++ {
			sumX += sX[j][0]
		}
	}
}

func Benchmark_Range_OneIterVar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sumY = 0
		for j := range sY {
			sumY += sY[j][0]
		}
	}
}

func Benchmark_Range_TwoIterVar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sumZ = 0
		for _, v := range sZ {
			sumZ += v[0]
		}
	}
}
