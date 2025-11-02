package main

import (
	"runtime"
	"time"

	"github.com/aawadall/go-concurrency-patterns/config"
	"github.com/aawadall/go-concurrency-patterns/shared"
)

func main() {
	// this is a fan out fan in based client
	cfg := config.GetDefaultConfig()

	// Initial memory stats
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	var latencies []time.Duration
	var statuses []int
	startTime := time.Now()

	// define request channel
	requests := make(chan struct{}, cfg.Requests)

	// define response channel
	type response struct {
		latency time.Duration
		status  int
	}
	responses := make(chan response, cfg.Requests)

	// fan out
	for i := 0; i < cfg.Concurrency; i++ {
		go func() {
			for range requests {
				latency, status := shared.ConsumeServer(cfg)
				responses <- response{latency: latency, status: status}
			}
		}()
	}

	// send requests
	go func() {
		for i := 0; i < cfg.Requests; i++ {
			requests <- struct{}{}
		}
		close(requests)
	}()

	// collect responses
	for i := 0; i < cfg.Requests; i++ {
		resp := <-responses
		latencies = append(latencies, resp.latency)
		statuses = append(statuses, resp.status)
	}
	close(responses)

	// fan in complete

	totalTime := time.Since(startTime)

	// Final memory stats
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate memory usage
	memUsed := m2.Alloc - m1.Alloc
	peakMem := m2.Sys

	memProfile := map[string]uint64{
		"InitialAlloc": m1.Alloc,
		"FinalAlloc":   m2.Alloc,
		"MemUsed":      memUsed,
		"TotalAlloc":   m2.TotalAlloc,
		"Sys":          m2.Sys,
		"NumGC":        uint64(m2.NumGC),
		"PeakMem":      peakMem,
	}

	shared.Report(latencies, statuses, totalTime, memProfile)
	
}
