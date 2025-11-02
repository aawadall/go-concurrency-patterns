# Getting Started Guide

Welcome to the Go Concurrency Patterns project! This guide will help you get up and running quickly.

## Prerequisites

### System Requirements
- Go 1.22.2 or later
- Python 3.8 or later
- pip (Python package manager)

### Installation

#### Go
```bash
# Verify Go installation
go version
# Should output: go version go1.22.2 (or later)
```

#### Python
```bash
# Verify Python installation
python3 --version
# Should output: Python 3.8.x (or later)
```

---

## Quick Start (5 minutes)

### 1. Clone or Navigate to Repository

```bash
cd /path/to/go-concurrency-patterns
```

### 2. Run All Patterns (Automated)

The simplest way to get started is to run the automated test suite:

```bash
./simulate.sh
```

This script will:
1. Set up a Python virtual environment
2. Install server dependencies
3. Start the Flask test server
4. Run all client implementations sequentially
5. Display performance reports for each pattern

**Expected Output:**
```
Starting Flask server...
Server running on localhost:5000

=== Running Simple (Sequential) Client ===
.......................................
Average Latency: XXms
Total Time: XXs
...

=== Running WaitGroups Client ===
.......................................
Average Latency: XXms
Total Time: XXs
...

[Results for FanOutIn and FanOutInWBP follow]
```

---

## Manual Setup (For Development)

If you want more control over the process, follow these manual steps:

### Step 1: Start the Test Server

Open a terminal and navigate to the server directory:

```bash
cd server
```

Create and activate a Python virtual environment:

```bash
# Create virtual environment
python3 -m venv venv

# Activate it
# On Linux/macOS:
source venv/bin/activate
# On Windows:
venv\Scripts\activate
```

Install dependencies:

```bash
pip install -r requirements.txt
```

Start the server:

```bash
python3 server.py
```

**Expected Output:**
```
 * Serving Flask app 'server'
 * Running on http://127.0.0.1:5000
 * Press CTRL+C to quit
```

Keep this terminal open; the server should remain running.

### Step 2: Run Clients in Another Terminal

Open a new terminal in the project root:

```bash
cd /path/to/go-concurrency-patterns
```

Run individual clients:

```bash
# Sequential baseline
go run cmd/simple/main.go

# WaitGroup-based
go run cmd/waitgroups/main.go

# Fan-Out/Fan-In
go run cmd/fanoutin/main.go

# Fan-Out/Fan-In with Backpressure
go run cmd/fanoutinwbp/main.go
```

---

## Understanding the Output

### Sample Report Output

```
=== Performance Report ===
Average Latency: 150ms
Total Time: 10s

Status Code Distribution:
  200: 7450 requests
  500: 50 requests

Memory Profile:
  Current Allocation: 15mb
  Total Allocated: 45mb
  System Reserved: 60mb
  Garbage Collections: 12
```

### Interpreting Metrics

| Metric | Meaning | What to Look For |
|--------|---------|------------------|
| Average Latency | Mean response time per request | Lower is better |
| Total Time | Wall-clock time to complete all requests | Lower is better |
| Status 200 | Successful requests | Should be majority |
| Status 500 | Failed requests (errors) | Should be minimal |
| Current Allocation | Memory in use right now | Lower memory = more efficient |
| Total Allocated | Total memory ever allocated | Indicator of GC pressure |
| Garbage Collections | How many times GC ran | More = more allocations |

---

## Comparing Patterns

### Quick Comparison

Run all patterns and compare outputs to understand:

1. **Speed:** Which pattern completes fastest?
   - Simple is slowest (sequential baseline)
   - WaitGroup is usually fastest (unbounded concurrency)
   - FanOutIn is balanced
   - FanOutInWBP is adaptive

2. **Memory:** Which pattern uses least memory?
   - Simple uses least
   - WaitGroup uses most (7,500 goroutines)
   - FanOutIn uses medium (15 workers)
   - FanOutInWBP uses medium (15 workers + monitor)

3. **Stability:** Which produces most consistent latencies?
   - Simple is most consistent but slowest
   - WaitGroup is fast but variable
   - FanOutIn is balanced
   - FanOutInWBP is smooth (adapts to load)

### Expected Results on Typical Systems

```
Pattern            Time       Memory    Throughput
Sequential         100%       100%      Baseline
WaitGroups         ~20-30%    ~400%     Very High
FanOutIn           ~25-35%    ~150%     High
FanOutInWBP        ~30-40%    ~150%     Adaptive
```

**Note:** Actual percentages vary based on:
- Server capacity
- Network latency
- System resources
- Server load during test

---

## Configuration

### Modifying Default Settings

Edit the configuration in `config/config.go`:

```go
func GetDefaultConfig() *Config {
    return &Config{
        Host:        "localhost",      // Server host
        Port:        5000,             // Server port
        Requests:    7500,             // Total requests
        Concurrency: 15,               // Worker pool size
    }
}
```

### Common Configurations

#### Quick Test (1,000 requests)
```go
return &Config{
    Host:        "localhost",
    Port:        5000,
    Requests:    1000,      // Reduced from 7500
    Concurrency: 15,
}
```

#### Remote Server Test
```go
return &Config{
    Host:        "api.example.com",
    Port:        8080,
    Requests:    7500,
    Concurrency: 20,       // Increase for remote servers
}
```

#### High Concurrency Test
```go
return &Config{
    Host:        "localhost",
    Port:        5000,
    Requests:    10000,
    Concurrency: 50,       // More workers
}
```

### Overriding in Code

```go
cfg := config.NewConfig("api.example.com", 8080)
cfg.Requests = 5000
cfg.Concurrency = 25
```

---

## Troubleshooting

### Issue: "Connection refused" Error

**Problem:** Clients can't connect to the server.

**Solution:**
1. Ensure Flask server is running (check that terminal)
2. Verify server is on correct host/port:
   ```bash
   lsof -i :5000  # Check if port 5000 is in use
   ```
3. Check config host/port match server settings
4. If on different machine, update Host in config to server IP

### Issue: "Address already in use"

**Problem:** Port 5000 is already in use.

**Solution:**
1. Find what's using port 5000:
   ```bash
   lsof -i :5000
   ```
2. Kill the process:
   ```bash
   kill -9 <PID>
   ```
3. Or change server port in `server/server.py` and update config

### Issue: "Permission denied" on simulate.sh

**Problem:** Can't execute the script.

**Solution:**
```bash
chmod +x simulate.sh
./simulate.sh
```

### Issue: Python Virtual Environment Issues

**Problem:** `source venv/bin/activate` doesn't work.

**Solution:**
```bash
# Try this for different shells
# Bash/Zsh:
source venv/bin/activate

# Fish:
source venv/bin/activate.fish

# PowerShell:
venv/Scripts/Activate.ps1
```

### Issue: Module Import Errors in Go

**Problem:** `go run` says module not found.

**Solution:**
```bash
# From project root:
go mod tidy
go run cmd/simple/main.go
```

---

## Next Steps

Once you have the basic setup working:

### 1. Explore the Code

- Read `/docs/ARCHITECTURE.md` for system design
- Read `/docs/PATTERNS.md` to understand each pattern
- Read `/docs/API.md` for detailed API reference

### 2. Compare Pattern Performance

Run each pattern and compare the metrics:

```bash
# Terminal 1: Start server
cd server && python3 server.py

# Terminal 2: Run all patterns and collect results
go run cmd/simple/main.go > results_simple.txt
go run cmd/waitgroups/main.go > results_wg.txt
go run cmd/fanoutin/main.go > results_foi.txt
go run cmd/fanoutinwbp/main.go > results_foibp.txt

# Compare results
cat results_*.txt
```

### 3. Modify and Experiment

Try changing the configuration:

```bash
# Edit config/config.go
# Change Requests to 1000 for faster tests
# Change Concurrency to different values
# Rebuild and run

go run cmd/fanoutin/main.go
```

### 4. Add Your Own Pattern

Follow the template in `/docs/API.md` to add a new pattern:

```bash
cp cmd/fanoutin/main.go cmd/mypattern/main.go
# Edit mypattern to implement your approach
go run cmd/mypattern/main.go
```

### 5. Read the Source Code

Key files to understand:
- `cmd/simple/main.go` - Simplest pattern (start here)
- `cmd/fanoutin/main.go` - Most useful pattern for production
- `shared/client.go` - HTTP abstraction
- `shared/report.go` - Metrics calculation
- `server/server.py` - Server simulation

---

## Learning Resources

### Built-in Documentation

1. **ARCHITECTURE.md** - System design and components
2. **PATTERNS.md** - Detailed pattern explanations
3. **API.md** - Function signatures and usage examples
4. **This file** - Getting started guide

### Key Concepts to Understand

- **Goroutines:** Lightweight concurrency units in Go
- **Channels:** Inter-goroutine communication mechanism
- **sync.WaitGroup:** Synchronization primitive for goroutines
- **Worker Pool:** Pattern for bounded concurrency
- **Backpressure:** Mechanism to prevent system overload

### Go Documentation

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Context Package](https://golang.org/pkg/context/)
- [Sync Package](https://golang.org/pkg/sync/)

---

## Performance Tips

### For Accurate Benchmarks

1. **Isolate Tests:** Run on a quiet system with no other workloads
2. **Warm Up:** Run server for a bit before testing (let it stabilize)
3. **Repeated Runs:** Run each pattern 3+ times and average results
4. **Same Configuration:** Keep same request count and worker count across patterns
5. **Monitor Resources:** Watch CPU and memory during tests

### Optimizing for Your Use Case

- **Maximum Throughput:** Use WaitGroup or increase FanOutIn workers
- **Minimum Latency:** Use FanOutIn with tuned concurrency
- **Memory Efficiency:** Use FanOutIn or FanOutInWBP
- **Server Protection:** Use FanOutInWBP with backpressure
- **Reliability:** Use FanOutIn with error handling

---

## Common Questions

### Q: Which pattern should I use for my application?

**A:** It depends on your needs:
- Production API client → FanOutIn or FanOutInWBP
- Learning Go concurrency → Sequential then WaitGroup then FanOutIn
- Maximum speed → WaitGroup (if memory permits)
- Server protection → FanOutInWBP

### Q: Can I use these patterns with a real API?

**A:** Yes! Simply change the Config host and port to point to your API server, and adjust the endpoint in `shared/client.go` if needed.

### Q: How do I test with HTTPS?

**A:** Modify `shared/client.go` to use HTTPS URLs:
```go
url := fmt.Sprintf("https://%s:%d/data", cfg.Host, cfg.Port)
```

### Q: Can I add custom logic to requests?

**A:** Yes! Modify `shared/client.go` to add headers, body content, or other customizations.

### Q: Why is WaitGroup so much faster but uses more memory?

**A:** More goroutines = more parallelism = faster, but each goroutine uses memory. It's a trade-off.

### Q: How accurate are the latency measurements?

**A:** Very accurate for request round-trip time, but total time includes measurement overhead and scheduling delays.

---

## Getting Help

### Check the Documentation

1. Start with README.md in the root
2. Read ARCHITECTURE.md for system overview
3. Read PATTERNS.md for pattern details
4. Refer to API.md for specific functions

### Troubleshoot Systematically

1. Verify prerequisites are installed
2. Check server is running
3. Verify network connectivity to server
4. Review error messages in logs
5. Try manual setup instead of script

### Inspect the Code

All code is in `cmd/` and `shared/` directories. Start with `cmd/simple/main.go` as it's the simplest implementation.

---

## Summary

You now have:
- ✓ A running test server
- ✓ Four concurrent pattern implementations
- ✓ Performance reporting and metrics
- ✓ Detailed documentation
- ✓ Tools to compare patterns

Next: Run the patterns, compare results, and explore the code to deepen your understanding of Go concurrency!
