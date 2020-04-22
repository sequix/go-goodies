package muxer

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkMux10(b *testing.B) {
	const jobscount = 10
	benchMux(b, jobscount)
}

func BenchmarkMux100(b *testing.B) {
	const jobscount = 100
	benchMux(b, jobscount)
}

func BenchmarkMux1000(b *testing.B) {
	const jobscount = 1000
	benchMux(b, jobscount)
}

func BenchmarkMux2500(b *testing.B) {
	const jobscount = 2500
	benchMux(b, jobscount)
}

func BenchmarkMux5000(b *testing.B) {
	const jobscount = 5000
	benchMux(b, jobscount)
}

func BenchmarkMux10000(b *testing.B) {
	const jobscount = 10000
	benchMux(b, jobscount)
}

const minTimeMilli = 100
const maxTimeMilli = 1000

func benchMux(b *testing.B, jobscount int) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		muxJobs(b ,jobscount)
	}
}

func muxJobs(b *testing.B, jobscount int) {
	sources := []interface{}{}
	for i := 0; i < jobscount-1; i++ {
		sources = append(sources, newWorker())
	}
	sources = append(sources, newWorstCaseWorker())

	sink := make(chan time.Duration)
	if err := Do(sink, sources...); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for range sink {
		b.SetBytes(1)
	}
}

func newWorker() <-chan time.Duration {
	res := make(chan time.Duration)
	go func() {
		sleep := time.Duration(rand.Intn(maxTimeMilli-minTimeMilli) + minTimeMilli)
		time.Sleep(sleep * time.Millisecond)
		res <- sleep
		close(res)
	}()

	return res
}

func newWorstCaseWorker() <-chan time.Duration {
	res := make(chan time.Duration)
	go func() {
		sleep := maxTimeMilli * time.Millisecond
		time.Sleep(sleep)
		res <- sleep
		close(res)
	}()
	return res
}