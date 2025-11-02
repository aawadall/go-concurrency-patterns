package main

import (
	"runtime"
	"sync"
	"time"

	"github.com/aawadall/go-concurrency-patterns/config"
	"github.com/aawadall/go-concurrency-patterns/shared"
)

func main() {
	cfg := config.GetDefaultConfig()

	// Initial memory stats
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	var latencies []time.Duration
	var statuses []int

	startTime := time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < cfg.Requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			latency, status := shared.ConsumeServer(cfg)
			latencies = append(latencies, latency)
			statuses = append(statuses, status)
		}()
	}
	wg.Wait()
	totalTime := time.Since(startTime)

	// Final memory stats
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	memProfile := map[string]uint64{
		"Alloc":      m2.Alloc - m1.Alloc,
		"TotalAlloc": m2.TotalAlloc - m1.TotalAlloc,
		"Sys":        m2.Sys - m1.Sys,
		"NumGC":      uint64(m2.NumGC - m1.NumGC),
	}
	shared.Report(latencies, statuses, totalTime, memProfile)
}
