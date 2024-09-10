# Ratelimiter

Ratelimiter is a flexible and efficient rate limiting library for Go applications. It provides multiple rate limiting algorithms and a simple interface for easy integration into your projects.

## Features

- Multiple rate limiting algorithms:
    - Token Bucket
    - Fixed Window
    - Sliding Window
    - Nested Window
- Configurable rate and burst limits
- Blocking and non-blocking operations
- Concurrent-safe implementation
- Built-in metrics collection
- Easy to use and extend

## Installation

To install the ratelimiter package, use the following command:

```bash
go get github.com/popeskul/ratelimiter
```

## Quick Start

Here's a simple example of how to use the ratelimiter package:

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/popeskul/ratelimiter"
)

func main() {
    limiter, err := ratelimiter.New(
        ratelimiter.WithAlgorithm("token_bucket"),
        ratelimiter.WithRate(10),
        ratelimiter.WithCapacity(5),
        ratelimiter.WithMetrics(true),
    )
    if err != nil {
        panic(err)
    }

    for i := 0; i < 20; i++ {
        if limiter.Allow() {
            fmt.Println("Request allowed")
        } else {
            fmt.Println("Request denied")
        }
    }

    // Wait for a request to be allowed
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := limiter.Wait(ctx); err != nil {
        fmt.Printf("Wait error: %v\n", err)
    } else {
        fmt.Println("Request allowed after waiting")
    }

    // Print metrics
    metrics := limiter.GetMetrics()
    fmt.Printf("Metrics: %+v\n", metrics)
}
```

## Configuration

The ratelimiter package provides several configuration options:

- `WithAlgorithm(algo string)`: Set the rate limiting algorithm ("token_bucket", "fixed_window", "sliding_window", or "nested_window")
- `WithRate(rate int)`: Set the maximum number of requests per time window
- `WithBurst(burst int)`: Set the maximum burst size (for algorithms that support it)
- `WithCapacity(capacity int)`: Set the token bucket capacity (for Token Bucket algorithm)
- `WithWindow(window time.Duration)`: Set the time window for window-based algorithms
- `WithMetrics(enabled bool)`: Enable or disable metrics collection

## Algorithms

### Token Bucket

The Token Bucket algorithm uses a bucket that continuously fills with tokens at a fixed rate. Each request consumes one token, and if there are no tokens available, the request is denied.

### Fixed Window

The Fixed Window algorithm tracks the number of requests in fixed time windows. Once the limit is reached for a window, subsequent requests are denied until a new window begins.

### Sliding Window

The Sliding Window algorithm provides a smoother rate limiting experience by considering a sliding time window, which helps prevent sudden bursts at the edges of fixed windows.

### Nested Window

The Nested Window algorithm combines two windows: a larger outer window for the overall rate limit and a smaller inner window for short-term burst control.

## Metrics

When metrics are enabled, the ratelimiter provides the following information:

- Total requests
- Allowed requests
- Denied requests
- Current rate
- Last reset time
- Total wait time
- Maximum wait time
- Window duration
- Inner rate (for Nested Window)
- Inner window duration (for Nested Window)

## Contributing

Contributions to the ratelimiter package are welcome! Please feel free to submit issues, fork the repository and send pull requests!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
