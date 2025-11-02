package shared

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aawadall/go-concurrency-patterns/config"
)

const RED = "\033[0;31m"
const RESET = "\033[0m"

func ConsumeServer(cfg *config.Config) (latency time.Duration, status int) {
	startTime := time.Now()
	defer func() {
		latency = time.Since(startTime)
	}()
	// URL + port
	path := "/data"
	serverURL := fmt.Sprintf("http://%s:%d%s", cfg.Host, cfg.Port, path)

	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		fmt.Printf("%s Error parsing URL: %v %s\n", RED, err, RESET)
		return -1, 500
	}

	// Create HTTP client
	client := &http.Client{}

	// Create HTTP request
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		fmt.Printf("%s Error creating request: %v %s\n", RED, err, RESET)
		return -1, 500
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s Error performing request: %v %s\n", RED, err, RESET)
		return -1, 500
	}
	defer resp.Body.Close()

	// Read the response body
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		fmt.Printf("%s Error reading response body: %v %s\n", RED, err, RESET)
		return -1, resp.StatusCode
	}

	fmt.Printf(".")

	return latency, resp.StatusCode
}
