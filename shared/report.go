package shared

import (
	"fmt"
	"time"
)

func Report(latencies []time.Duration, statuses []int, totalTime time.Duration, memProfile map[string]uint64) {
	var totalLatency time.Duration
	statusCount := make(map[int]int)

	for i, latency := range latencies {
		totalLatency += latency
		status := statuses[i]
		statusCount[status]++
	}

	avgLatency := totalLatency / time.Duration(len(latencies))
	// Yellow color for report
	fmt.Printf("\033[0;33m")
	fmt.Printf("\n\nAverage Latency: %v\n", avgLatency)
	fmt.Printf("Total Time: %v\n", totalTime)
	fmt.Println("Status Code Counts:")
	for status, count := range statusCount {
		fmt.Printf("  %d: %d\n", status, count)
	}
	fmt.Printf("\033[0m")

	if memProfile != nil {
		fmt.Printf("\nMemory Profile:\n")
		for key, value := range memProfile {
			fmt.Printf("  %s: %d\n", key, value)
		}
	}
}
