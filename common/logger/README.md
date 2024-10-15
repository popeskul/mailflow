# Logger Package

A structured logging package for Go applications with support for multiple output formats, log levels, and context-based logging.

## Features

- Multiple log levels (Debug, Info, Warn, Error, Fatal)
- Structured logging with fields
- Context-based logging with trace ID support
- JSON format support
- File rotation support
- Multiple output writers support
- Zap logger implementation

## Installation

```bash
go get -u git@ssh.antgit.com:alp-pay/alp-pay-libs
```

## Quick Start

```go
package main

import (
    "github.com/popeskul/mailflow/common/logger"
)

func main() {
    // Create a new logger
    log := logger.NewZapLogger(
        logger.WithLogLevel(logger.InfoLevel),
        logger.WithJSONFormat(),
    )
    defer log.Sync()

    // Basic logging
    log.Info("Application started", 
        logger.Field{Key: "version", Value: "1.0.0"},
    )

    // With fields
    log.WithFields(logger.Fields{
        "user_id": "123",
        "action": "login",
    }).Info("User logged in")
}
```

## Log Levels

The package supports the following log levels:
- Debug
- Info
- Warn
- Error
- Fatal

## Configuration Options

### Setting Log Level

```go
logger.WithLogLevel(logger.InfoLevel)
```

### JSON Format

```go
logger.WithJSONFormat()
```

### File Rotation

```go
logger.WithFileRotation(
    "/var/log/app.log", // file path
    10,                 // max size in MB
    5,                  // max backups
    30,                 // max age in days
)
```

### Multiple Outputs

```go
logger.WithOutputs(os.Stdout, file)
```

## Testing

Package includes comprehensive test suite. To run tests:

```bash
go test -v ./...
```
```
