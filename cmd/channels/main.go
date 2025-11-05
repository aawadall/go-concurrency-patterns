package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// play with channels
const (
	bufferSize  = 5
	workerCount = 2
	jobCount    = 3
)

func main() {
	ch := make(chan int, bufferSize)
	resultChan := make(chan int, bufferSize)

	defer close(ch)
	defer close(resultChan)

	var wg sync.WaitGroup

	for i := range jobCount {
		wg.Add(1)
		log.Println("Main loop iteration:", i)
		go func(param int) {
			defer wg.Done()
			channelWriter(ch, param)
		}(i)
	}

	// wait for all writers to complete
	wg.Wait()

	// drain the channel ch with reader goroutines
	readerWg := sync.WaitGroup{}
drainLoop:
	for {
		select {
		case val := <-ch:
			readerWg.Add(1)
			go func(v int) {
				defer readerWg.Done()
				channelReader(v, resultChan)
			}(val)
			log.Println("Drained value from channel:", val)
		default:
			log.Println("No more values to drain from channel")
			break drainLoop
		}
	}

	readerWg.Wait()
	log.Println("All workers completed")

	// dump results from resultChan until empty
resultLoop:
	for {
		select {
		case result := <-resultChan:
			log.Println("Result from resultChan:", result)
		default:
			log.Println("No more results in resultChan")
			break resultLoop
		}
	}
	log.Println("Program completed")
}

func channelWriter(ch chan<- int, param int) {
	log.Println("Writing to channel:", param+1)
	// sleep for random time
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	ch <- param + 1
}

func channelReader(val int, resultChan chan<- int) {
	log.Println("Reading value:", val)
	// sleep for random time
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	resultChan <- val * 2
}
