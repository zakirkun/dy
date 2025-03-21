# dy

A simple, flexible, and efficient logging package for Go applications with color support.

## Features

- Multiple log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Customizable output format with timestamps and prefixes
- Function call tracing with automatic indentation
- Support for JSON output format
- Caller information (file, line, function)
- Colored output for terminal environments
- Thread-safe operation
- Low overhead with level-based filtering
- Global default logger and instance-based loggers

## Installation

```bash
go get github.com/zakirkun/dy
```

## Basic Usage

```go
package main

import (
    "github.com/zakirkun/dy"
)

func main() {
    // Use the default logger
    dy.Info("This is an informational message")
    dy.Warn("This is a warning message")
    dy.Error("This is an error message: %s", "connection timeout")
    
    // Create a custom logger
    logger := dy.New(
        dy.WithPrefix("APP"),
        dy.WithLevel(dy.DebugLevel),
        dy.WithColor(true), // Enable colored output
    )
    
    logger.Debug("Debug message with color")
    logger.Info("Custom logger info message")
}
```

## Colored Output

The logger supports colored output in terminal environments:

- DEBUG: Blue
- INFO: Green
- WARN: Yellow
- ERROR: Red
- FATAL: Bold Red

Colors are automatically enabled by default when writing to a terminal (os.Stdout or os.Stderr) and can be toggled with the `WithColor` option.

## Function Call Tracing

```go
func processData() {
    // Log entry and exit of function with proper indentation
    defer dy.TraceFunction()()
    
    dy.Info("Processing data...")
    // Do work here
}
```

## JSON Output Format

```go
logger := dy.New(
    dy.WithJSONFormat(true),
    dy.WithCallerInfo(true),
)

logger.Info("This will be output in JSON format")
```

## Configuration Options

- `WithOutput(io.Writer)`: Set custom output destination
- `WithLevel(Level)`: Set minimum log level 
- `WithPrefix(string)`: Add a prefix to all log messages
- `WithTimestamp(bool)`: Enable/disable timestamps
- `WithTrace(bool)`: Enable/disable function call tracing
- `WithIndentString(string)`: Customize indentation for nested function calls
- `WithJSONFormat(bool)`: Enable/disable JSON format
- `WithCallerInfo(bool)`: Include caller file/line information
- `WithColor(bool)`: Enable/disable colored output

## Performance

The package is designed to be efficient:

- Minimal heap allocations
- Level-based message filtering before string formatting
- Concurrency-safe through mutex locking
- Benchmark suite included

## Example

See the `example` folder for complete examples of usage.