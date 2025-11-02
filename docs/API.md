# API Reference

## Configuration Package (`config/`)

### Types

#### `Config`
Configuration struct for client implementations.

```go
type Config struct {
    Host        string  // Server host (e.g., "localhost")
    Port        int     // Server port (e.g., 5000)
    Requests    int     // Total number of requests to send
    Concurrency int     // Number of concurrent workers (for worker pool patterns)
}
```

**Default Values:**
- Host: "localhost"
- Port: 5000
- Requests: 7500
- Concurrency: 15

### Functions

#### `NewConfig(host string, port int) *Config`

Creates a new configuration with specified host and port, using default values for requests and concurrency.

**Parameters:**
- `host` (string): Server hostname or IP address
- `port` (int): Server port number

**Returns:**
- `*Config`: Pointer to new Config struct with defaults applied

**Example:**
```go
cfg := config.NewConfig("api.example.com", 8080)
```

#### `GetDefaultConfig() *Config`

Returns a pre-configured Config with standard production defaults.

**Parameters:** None

**Returns:**
- `*Config`: Pointer to Config with defaults (localhost:5000, 7500 requests, 15 concurrency)

**Example:**
```go
cfg := config.GetDefaultConfig()
latencies := make([]time.Duration, 0, cfg.Requests)
```

---

## Shared Package (`shared/`)

### Client Module (`client.go`)

#### `ConsumeServer(cfg *Config) (latency time.Duration, status int)`

Makes a single HTTP GET request to the test server and measures response latency.

**Parameters:**
- `cfg` (*Config): Configuration containing host and port

**Returns:**
- `latency` (time.Duration): Round-trip time for the request
- `status` (int): HTTP status code (200 for success, 500 for errors)

**Behavior:**
- Makes GET request to `http://{cfg.Host}:{cfg.Port}/data`
- Measures round-trip time from start to finish
- Prints "." to stdout on success
- Prints "E" to stdout on error
- Returns status 500 if any error occurs

**Example:**
```go
cfg := config.GetDefaultConfig()
latency, status := shared.ConsumeServer(cfg)

if status == 200 {
    fmt.Printf("Request succeeded in %v\n", latency)
}
```

**Error Handling:**
- Network errors: Returns (0, 500)
- Timeout: Returns (0, 500)
- Server errors: Returns response latency and actual status code

### Reporting Module (`report.go`)

#### `Report(latencies []time.Duration, statuses []int, totalTime time.Duration, memProfile map[string]uint64)`

Aggregates performance metrics and prints formatted results to stdout.

**Parameters:**
- `latencies` ([]time.Duration): Slice of individual request latencies
- `statuses` ([]int): Slice of HTTP status codes for each request
- `totalTime` (time.Duration): Wall-clock time for entire execution
- `memProfile` (map[string]uint64): Memory statistics with keys:
  - "Alloc": Current memory allocation (bytes)
  - "TotalAlloc": Total memory ever allocated (bytes)
  - "Sys": System memory reserved (bytes)
  - "NumGC": Number of garbage collections

**Returns:** None (prints to stdout)

**Output Format:**
```
=== Performance Report ===
Average Latency: XXms
Total Time: XXs
Status Code Distribution:
  200: XXXX requests
  500: XX requests
Memory Profile:
  Current Allocation: XXmb
  Total Allocated: XXmb
  System Reserved: XXmb
  Garbage Collections: XX
```

**Example:**
```go
var m runtime.MemStats
runtime.ReadMemStats(&m)

memProfile := map[string]uint64{
    "Alloc":     m.Alloc,
    "TotalAlloc": m.TotalAlloc,
    "Sys":       m.Sys,
    "NumGC":     uint64(m.NumGC),
}

shared.Report(latencies, statuses, totalTime, memProfile)
```

**Metrics Calculated:**
- **Average Latency:** Sum of all latencies divided by count
- **Status Distribution:** Count of each unique status code
- **Memory Efficiency:** Ratio of allocations
- **GC Activity:** Number of garbage collection cycles

---

## Client Implementation Patterns

### Pattern 1: Sequential Client (`cmd/simple/main.go`)

Basic pattern for sending requests sequentially.

```go
func main() {
    cfg := config.GetDefaultConfig()
    latencies := make([]time.Duration, 0, cfg.Requests)
    statuses := make([]int, 0, cfg.Requests)

    start := time.Now()

    for i := 0; i < cfg.Requests; i++ {
        latency, status := shared.ConsumeServer(cfg)
        latencies = append(latencies, latency)
        statuses = append(statuses, status)
    }

    totalTime := time.Since(start)

    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    memProfile := map[string]uint64{
        "Alloc":      m.Alloc,
        "TotalAlloc": m.TotalAlloc,
        "Sys":        m.Sys,
        "NumGC":      uint64(m.NumGC),
    }

    shared.Report(latencies, statuses, totalTime, memProfile)
}
```

**Key Points:**
- No goroutines or concurrency
- Direct sequential loop
- Simple error handling (errors already counted in ConsumeServer)
- Baseline for performance comparison

---

### Pattern 2: WaitGroup Client (`cmd/waitgroups/main.go`)

Pattern using goroutines with `sync.WaitGroup` synchronization.

```go
func main() {
    cfg := config.GetDefaultConfig()
    var latencies []time.Duration
    var statuses []int
    var mu sync.Mutex
    var wg sync.WaitGroup

    start := time.Now()

    for i := 0; i < cfg.Requests; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            latency, status := shared.ConsumeServer(cfg)

            mu.Lock()
            latencies = append(latencies, latency)
            statuses = append(statuses, status)
            mu.Unlock()
        }()
    }

    wg.Wait()
    totalTime := time.Since(start)

    // Memory profiling and reporting...
}
```

**Key Points:**
- One goroutine per request (7,500 total)
- `sync.Mutex` protects slice access
- High concurrency, unbounded
- Requires synchronization for shared data

---

### Pattern 3: Fan-Out/Fan-In Client (`cmd/fanoutin/main.go`)

Pattern using worker pool with channels.

```go
func main() {
    cfg := config.GetDefaultConfig()

    requestChan := make(chan int, 100)
    responseChan := make(chan Result, 100)

    // Start workers
    for i := 0; i < cfg.Concurrency; i++ {
        go func() {
            for range requestChan {
                latency, status := shared.ConsumeServer(cfg)
                responseChan <- Result{latency, status}
            }
        }()
    }

    start := time.Now()

    // Fan-out: distribute requests
    go func() {
        for i := 0; i < cfg.Requests; i++ {
            requestChan <- i
        }
        close(requestChan)
    }()

    // Fan-in: collect responses
    latencies := make([]time.Duration, cfg.Requests)
    statuses := make([]int, cfg.Requests)
    for i := 0; i < cfg.Requests; i++ {
        result := <-responseChan
        latencies[i] = result.Latency
        statuses[i] = result.Status
    }

    totalTime := time.Since(start)

    // Memory profiling and reporting...
}
```

**Key Points:**
- Fixed number of worker goroutines (default: 15)
- Work distribution via channel
- Results collection via channel
- Bounded concurrency model

---

### Pattern 4: Fan-Out/Fan-In with Backpressure (`cmd/fanoutinwbp/main.go`)

Advanced pattern with adaptive load balancing.

```go
func main() {
    cfg := config.GetDefaultConfig()

    requestChan := make(chan int, 100)
    responseChan := make(chan Result, 100)
    backpressureChan := make(chan struct{}, 1000)

    // Workers with backpressure signaling
    for i := 0; i < cfg.Concurrency; i++ {
        go func() {
            for range requestChan {
                select {
                case backpressureChan <- struct{}{}:
                default:
                }

                latency, status := shared.ConsumeServer(cfg)
                responseChan <- Result{latency, status}

                <-backpressureChan
            }
        }()
    }

    // Backpressure monitor
    timeout := time.Duration(0)
    go func() {
        ticker := time.NewTicker(100 * time.Millisecond)
        for range ticker.C {
            pressure := len(backpressureChan)
            if pressure > 10 {
                timeout = min(timeout+5*time.Millisecond, 50*time.Millisecond)
            } else if pressure < 3 {
                timeout = 0
            }
        }
    }()

    start := time.Now()

    // Fan-out with adaptive delay
    go func() {
        for i := 0; i < cfg.Requests; i++ {
            if timeout > 0 {
                time.Sleep(timeout)
            }
            requestChan <- i
        }
        close(requestChan)
    }()

    // Fan-in: collect responses
    latencies := make([]time.Duration, cfg.Requests)
    statuses := make([]int, cfg.Requests)
    for i := 0; i < cfg.Requests; i++ {
        result := <-responseChan
        latencies[i] = result.Latency
        statuses[i] = result.Status
    }

    totalTime := time.Since(start)

    // Memory profiling and reporting...
}
```

**Key Points:**
- Additional backpressure signaling channel
- Monitor goroutine adjusts sleep timeout
- Adaptive based on worker load
- Protects downstream systems

---

## Type Definitions

### Result Struct (Used in worker pool patterns)

```go
type Result struct {
    Latency time.Duration
    Status  int
}
```

Used internally to pass request results from workers to main goroutine.

---

## Memory Profiling

### Runtime MemStats

The `runtime.MemStats` struct provides detailed memory information:

```go
var m runtime.MemStats
runtime.ReadMemStats(&m)

// Key fields used:
// m.Alloc        - Current memory allocation (bytes)
// m.TotalAlloc   - Total memory allocated (bytes, never decreases)
// m.Sys          - System memory reserved (bytes)
// m.NumGC        - Number of GC cycles completed
```

**Usage in Reporting:**
```go
memProfile := map[string]uint64{
    "Alloc":      m.Alloc,
    "TotalAlloc": m.TotalAlloc,
    "Sys":        m.Sys,
    "NumGC":      uint64(m.NumGC),
}
shared.Report(latencies, statuses, totalTime, memProfile)
```

---

## Common Patterns

### Measuring Execution Time

```go
start := time.Now()
// ... work ...
duration := time.Since(start)
```

### Collecting Metrics

```go
var latencies []time.Duration
var statuses []int

for i := 0; i < count; i++ {
    latency, status := shared.ConsumeServer(cfg)
    latencies = append(latencies, latency)
    statuses = append(statuses, status)
}
```

### Error Handling

ConsumeServer handles errors internally and returns status 500 for any errors. Callers should treat status 500 as an error indicator.

```go
latency, status := shared.ConsumeServer(cfg)
if status == 200 {
    // Success
} else {
    // Error occurred
}
```

---

---

## Pipelines Package (`cmd/pipelines/pipeline/`)

### Types

#### `Message[T]`

Generic message wrapper for data flowing through the pipeline.

```go
type Message[T any] struct {
    ID      int64  // Unique message identifier for tracking
    Payload T      // Generic payload of any type
}
```

**Generic Parameter `T`:**
- Can be any Go type: `int`, `string`, `struct`, etc.
- Type-safe at compile time
- Example: `Message[int]`, `Message[User]`, `Message[[]float64]`

**Fields:**
- **ID:** Unique identifier for message tracking and correlation across stages
- **Payload:** The actual data being processed (generic type)

**Example:**
```go
// Integer message
msg := Message[int]{ID: 1, Payload: 42}

// String message
msg := Message[string]{ID: 2, Payload: "hello"}

// Struct message
type User struct {
    ID   int
    Name string
}
msg := Message[User]{ID: 3, Payload: User{1, "Alice"}}
```

#### `Stage[I, O]`

Configurable processing stage with type-safe input/output transformation.

```go
type Stage[I any, O any] struct {
    Name     string                                  // Stage identifier
    Workers  int                                     // Number of concurrent workers
    Buffer   int                                     // Channel buffer size
    Function func(Message[I]) (Message[O], error)   // Transformation function
}
```

**Generic Parameters:**
- `I` - Input message type (what this stage consumes)
- `O` - Output message type (what this stage produces)
- Enables type-safe transformations (e.g., `Stage[int, string]`)

**Fields:**
- **Name:** Human-readable identifier for logging and debugging
- **Workers:** Number of concurrent goroutines (worker pool size)
- **Buffer:** Buffered channel capacity (prevents blocking and deadlocks)
- **Function:** User-defined transformation function

**Default Values:**
- Workers: Configurable (typically 2-8)
- Buffer: Configurable (typically 10-100)

**Example:**
```go
// Simple stage: int → int
squareStage := Stage[int, int]{
    Name:    "Square",
    Workers: 3,
    Buffer:  4,
    Function: func(msg Message[int]) (Message[int], error) {
        return Message[int]{
            ID:      msg.ID,
            Payload: msg.Payload * msg.Payload,
        }, nil
    },
}

// Type-transforming stage: int → string
toStringStage := Stage[int, string]{
    Name:    "ToString",
    Workers: 2,
    Buffer:  4,
    Function: func(msg Message[int]) (Message[string], error) {
        return Message[string]{
            ID:      msg.ID,
            Payload: fmt.Sprintf("%d", msg.Payload),
        }, nil
    },
}
```

### Functions

#### `Stage[I, O].Run(ctx context.Context, input <-chan Message[I]) (<-chan Message[O], *errgroup.Group)`

Executes the stage with concurrent workers.

**Parameters:**
- `ctx` (context.Context) - For cancellation and deadline support
- `input` (<-chan Message[I]) - Read-only input channel

**Returns:**
- `<-chan Message[O]` - Output channel for downstream stages
- `*errgroup.Group` - For managing worker goroutines and collecting errors

**Behavior:**
1. Creates output channel with configured buffer size
2. Spawns N worker goroutines (where N = `Workers`)
3. Each worker reads from input channel and applies transformation
4. Workers respect context cancellation
5. Results sent to output channel
6. Output channel automatically closes when all workers complete
7. Error group manages synchronization and error collection

**Example:**
```go
ctx := context.Background()
input := make(chan Message[int])

stage := Stage[int, int]{
    Name:    "Process",
    Workers: 3,
    Buffer:  4,
    Function: func(msg Message[int]) (Message[int], error) {
        return Message[int]{ID: msg.ID, Payload: msg.Payload * 2}, nil
    },
}

// Run stage
output, eg := stage.Run(ctx, input)

// Send data
go func() {
    for i := 1; i <= 10; i++ {
        input <- Message[int]{ID: int64(i), Payload: i}
    }
    close(input)
}()

// Consume results
for result := range output {
    fmt.Printf("[%d]: %d\n", result.ID, result.Payload)
}

// Wait for completion
if err := eg.Wait(); err != nil {
    log.Printf("Error: %v", err)
}
```

**Error Handling:**
- Transformation function can return errors
- Errors are collected in error group
- Error group's Wait() returns first error encountered (if any)

**Context Support:**
- Respects context cancellation (context.Done())
- Respects context deadlines
- Workers exit gracefully on cancellation

**Example with timeout:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

output, eg := stage.Run(ctx, input)
// If processing exceeds 5 seconds, context cancels remaining work
```

### Complete Pipeline Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/aawadall/go-concurrency-patterns/cmd/pipelines/pipeline"
)

func main() {
    ctx := context.Background()

    // Create input data
    input := make(chan pipeline.Message[int])
    go func() {
        for i := 1; i <= 10; i++ {
            input <- pipeline.Message[int]{ID: int64(i), Payload: i}
        }
        close(input)
    }()

    // Stage 1: Square numbers
    square := pipeline.Stage[int, int]{
        Name:    "Square",
        Workers: 3,
        Buffer:  4,
        Function: func(msg pipeline.Message[int]) (pipeline.Message[int], error) {
            return pipeline.Message[int]{
                ID:      msg.ID,
                Payload: msg.Payload * msg.Payload,
            }, nil
        },
    }

    // Stage 2: Double result
    double := pipeline.Stage[int, int]{
        Name:    "Double",
        Workers: 2,
        Buffer:  4,
        Function: func(msg pipeline.Message[int]) (pipeline.Message[int], error) {
            return pipeline.Message[int]{
                ID:      msg.ID,
                Payload: msg.Payload * 2,
            }, nil
        },
    }

    // Wire stages
    out1, g1 := square.Run(ctx, input)
    out2, g2 := double.Run(ctx, out1)

    // Wait for completion
    go func() {
        _ = g1.Wait()
        _ = g2.Wait()
    }()

    // Consume results
    for result := range out2 {
        fmt.Printf("[%d]: %d\n", result.ID, result.Payload)
    }
}
```

**Output:**
```
[1]: 2
[2]: 8
[3]: 18
[4]: 32
[5]: 50
[6]: 72
[7]: 98
[8]: 128
[9]: 162
[10]: 200
```

---

## Testing

Each client implementation can be run independently:

```bash
go run cmd/simple/main.go
go run cmd/waitgroups/main.go
go run cmd/fanoutin/main.go
go run cmd/fanoutinwbp/main.go
go run cmd/pipelines/main.go
```

Or run all via the automated script:

```bash
./simulate.sh
```

Ensure the test server is running before executing clients:
```bash
cd server
python3 server.py
```

---

## Extending the Framework

### Adding a New Pattern

1. Create `cmd/newpattern/main.go`
2. Import required packages
3. Implement request sending logic
4. Use `shared.Report()` to display results
5. Update `simulate.sh` to include new pattern

### Example Template

```go
package main

import (
    "time"
    "runtime"
    "github.com/aawadall/go-concurrency-patterns/config"
    "github.com/aawadall/go-concurrency-patterns/shared"
)

func main() {
    cfg := config.GetDefaultConfig()
    latencies := make([]time.Duration, cfg.Requests)
    statuses := make([]int, cfg.Requests)

    start := time.Now()

    // Your implementation here

    totalTime := time.Since(start)

    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    memProfile := map[string]uint64{
        "Alloc":      m.Alloc,
        "TotalAlloc": m.TotalAlloc,
        "Sys":        m.Sys,
        "NumGC":      uint64(m.NumGC),
    }

    shared.Report(latencies, statuses, totalTime, memProfile)
}
```

### Adding Custom Metrics

Extend `shared/report.go` to calculate additional metrics:

```go
func Report(latencies []time.Duration, statuses []int, totalTime time.Duration, memProfile map[string]uint64, customMetrics map[string]interface{}) {
    // Existing reporting code...

    // Custom metrics
    for key, value := range customMetrics {
        fmt.Printf("%s: %v\n", key, value)
    }
}
```
