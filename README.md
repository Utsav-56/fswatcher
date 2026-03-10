# Folder Watcher

A lightweight, file system watcher library for Go that monitors directories for changes and triggers callbacks for various file system events.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Components](#core-components)
- [Detailed Usage](#detailed-usage)
- [API Reference](#api-reference)
- [Examples](#examples)
- [How It Works](#how-it-works)

## Features

- **Directory Monitoring**: Watch directories for file and folder changes
- **Recursive Scanning**: Optionally monitor subdirectories with configurable depth limits
- **Event Types**: Track Create, Delete, Rename, and Modify events
- **Flexible Filtering**: Watch only files, only directories, or both
- **Callback-Based**: Register handlers for specific event types
- **Context Support**: Graceful shutdown using Go contexts
- **Polling-Based**: Simple polling mechanism (500ms intervals) for reliable cross-platform behavior

## Installation

```bash
go get github.com/fsnotify/fsnotify
```

Add to your project:

```go
import "github.com/yourproject/folder_watcher"
```

## Quick Start

```go
package watcher

import (
    "log"
)

func main() {
    // Create a new watcher
    fs := &FsWatcher{
        Path:    "./watch_directory",
        Options: NewOptions(),
        OnCreate: func(e CreateEvent) {
            log.Printf("Created dirs: %v, files: %v\n",
                e.DirsCreated, e.FilesCreated)
        },
        OnDelete: func(e DeleteEvent) {
            log.Printf("Deleted dirs: %v, files: %v\n",
                e.DirsDeleted, e.FilesDeleted)
        },
    }

    // Start monitoring
    fs.Start()
    defer fs.Stop()

    // Block forever (or until interrupted)
    select {}
}
```

## Core Components

### 1. FsWatcher

The main struct that manages file system monitoring.

**Fields:**

- `Path`: Directory path to watch
- `Options`: Configuration options
- `OnCreate`: Callback for creation events
- `OnDelete`: Callback for deletion events
- `OnRename`: Callback for rename events
- `OnModify`: Callback for modification events
- `OnChange`: Universal callback for all events

### 2. Options

Configures watcher behavior.

**Fields:**

- `Recursive`: Enable recursive directory scanning
- `RecursiveDepth`: Maximum depth for recursive scanning (-1 = unlimited)
- `DirsOnly`: Monitor only directories
- `FilesOnly`: Monitor only files
- `Verbose`: Enable detailed logging

### 3. Event Types

- `Event`: Base event structure
- `CreateEvent`: Files/directories created
- `DeleteEvent`: Files/directories deleted
- `RenameEvent`: Files/directories renamed
- `ModifyEvent`: Files/directories modified

## Detailed Usage

### Basic Watching (Non-Recursive)

Watch a single directory without recursion:

```go
fs := &FsWatcher{
    Path:    "./documents",
    Options: NewOptions(), // Recursive defaults to false
    OnChange: func(e Event) {
        log.Printf("Change detected at: %s\n", e.Path)
    },
}
fs.Start()
defer fs.Stop()
```

### Recursive Watching

Monitor a directory and all its subdirectories:

```go
opts := NewOptions()
opts.Recursive = true
opts.RecursiveDepth = -1 // Unlimited depth

fs := &FsWatcher{
    Path:    "./project",
    Options: opts,
    OnCreate: func(e CreateEvent) {
        log.Printf("New items created:\n")
        for _, dir := range e.DirsCreated {
            log.Printf("  DIR: %s\n", dir)
        }
        for _, file := range e.FilesCreated {
            log.Printf("  FILE: %s\n", file)
        }
    },
}
fs.Start()
defer fs.Stop()
```

### Recursive Watching with Depth Limit

Limit recursion to specific depth:

```go
opts := NewOptions()
opts.Recursive = true
opts.RecursiveDepth = 2 // Only 2 levels deep

fs := &FsWatcher{
    Path:    "./src",
    Options: opts,
    OnChange: func(e Event) {
        log.Println("Change detected!")
    },
}
fs.Start()
```

### Watch Only Directories

Monitor only directory changes, ignore files:

```go
opts := NewOptions()
opts.DirsOnly = true

fs := &FsWatcher{
    Path:    "./folders",
    Options: opts,
    OnCreate: func(e CreateEvent) {
        log.Printf("New directories: %v\n", e.DirsCreated)
        // e.FilesCreated will be nil
    },
    OnDelete: func(e DeleteEvent) {
        log.Printf("Deleted directories: %v\n", e.DirsDeleted)
        // e.FilesDeleted will be nil
    },
}
fs.Start()
```

### Watch Only Files

Monitor only file changes, ignore directories:

```go
opts := NewOptions()
opts.FilesOnly = true

fs := &FsWatcher{
    Path:    "./logs",
    Options: opts,
    OnCreate: func(e CreateEvent) {
        log.Printf("New files: %v\n", e.FilesCreated)
        // e.DirsCreated will be nil
    },
}
fs.Start()
```

### Using Context for Controlled Shutdown

Use context for timeout or cancellation:

```go
import (
    "context"
    "time"
)

// Timeout after 5 minutes
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

fs := &FsWatcher{
    Path:    "./temp",
    Options: NewOptions(),
    OnChange: func(e Event) {
        log.Println("Change detected")
    },
}

// Start with context - will auto-stop after timeout
fs.StartWithContext(ctx)
```

### Multiple Event Handlers

Handle different event types separately:

```go
fs := &FsWatcher{
    Path:    "./data",
    Options: NewOptions(),

    OnCreate: func(e CreateEvent) {
        for _, file := range e.FilesCreated {
            log.Printf("NEW FILE: %s\n", file)
            // Process new file...
        }
    },

    OnDelete: func(e DeleteEvent) {
        for _, file := range e.FilesDeleted {
            log.Printf("FILE DELETED: %s\n", file)
            // Cleanup or log deletion...
        }
    },

    OnChange: func(e Event) {
        // Universal handler - called for ALL events
        log.Printf("Event type: %v, Path: %s\n", e.Type, e.Path)
    },
}
fs.Start()
```

## API Reference

### Functions

#### `NewOptions() *Options`

Creates a new Options struct with default values.

**Returns:** Pointer to Options with defaults:

- `Recursive`: false
- `DirsOnly`: false
- `FilesOnly`: false
- `Verbose`: false
- `RecursiveDepth`: -1

**Example:**

```go
opts := NewOptions()
opts.Recursive = true
```

#### `readDir(path string) FsInfo`

Scans a single directory (non-recursively) and returns file system information.

**Parameters:**

- `path`: Directory path to scan

**Returns:** `FsInfo` containing maps of directories and files

**Example:**

```go
info := readDir("/home/user/documents")
fmt.Printf("Found %d directories\n", len(info.Dirs))
fmt.Printf("Found %d files\n", len(info.Files))
```

#### `readRecursive(path string, depth int) FsInfo`

Recursively scans a directory up to the specified depth.

**Parameters:**

- `path`: Root directory to scan
- `depth`: Maximum depth (0 = none, -1 or large number = unlimited)

**Returns:** `FsInfo` containing all directories and files found

**Example:**

```go
// Scan 3 levels deep
info := readRecursive("/home/user/projects", 3)

// Effectively unlimited
info := readRecursive("/home/user/projects", 10000)
```

#### `diffFsInfo(oldFs, newFs FsInfo) (addedDirs, removedDirs []string, addedFiles, removedFiles []string)`

Compares two file system snapshots and returns the differences.

**Parameters:**

- `oldFs`: Previous snapshot
- `newFs`: Current snapshot

**Returns:**

- `addedDirs`: Newly created directories
- `removedDirs`: Deleted directories
- `addedFiles`: Newly created files
- `removedFiles`: Deleted files

**Example:**

```go
oldSnapshot := readDir("/path/to/dir")
time.Sleep(2 * time.Second)
newSnapshot := readDir("/path/to/dir")

addedDirs, removedDirs, addedFiles, removedFiles := diffFsInfo(oldSnapshot, newSnapshot)
fmt.Printf("Changes: +%d/-%d dirs, +%d/-%d files\n",
    len(addedDirs), len(removedDirs), len(addedFiles), len(removedFiles))
```

### Methods

#### `(fs *FsWatcher) Start()`

Starts the watcher with a background context.

**Example:**

```go
fs := &FsWatcher{Path: "./watch", Options: NewOptions()}
fs.Start()
defer fs.Stop()
```

#### `(fs *FsWatcher) StartWithContext(ctx context.Context)`

Starts the watcher with a custom context for controlled lifecycle.

**Parameters:**

- `ctx`: Context for cancellation/timeout control

**Example:**

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

fs := &FsWatcher{Path: "./watch", Options: NewOptions()}
fs.StartWithContext(ctx)

// Cancel from another goroutine
go func() {
    time.Sleep(10 * time.Second)
    cancel() // Stops the watcher
}()
```

#### `(fs *FsWatcher) Stop()`

Stops the watcher gracefully.

**Example:**

```go
fs := &FsWatcher{Path: "./watch", Options: NewOptions()}
fs.Start()

// Later...
fs.Stop()
```

## Examples

### Example 1: Log File Monitor

Monitor a log directory and process new log files:

```go
opts := NewOptions()
opts.FilesOnly = true

fs := &FsWatcher{
    Path:    "./logs",
    Options: opts,
    OnCreate: func(e CreateEvent) {
        for _, logFile := range e.FilesCreated {
            if strings.HasSuffix(logFile, ".log") {
                log.Printf("Processing new log: %s\n", logFile)
                // Process the log file...
            }
        }
    },
}
fs.Start()
defer fs.Stop()

select {} // Keep running
```

### Example 2: Backup Trigger

Trigger backups when files are modified:

```go
fs := &FsWatcher{
    Path:    "./important_data",
    Options: NewOptions(),
    OnCreate: func(e CreateEvent) {
        if len(e.FilesCreated) > 0 {
            log.Println("New files detected, triggering backup...")
            // triggerBackup()
        }
    },
    OnDelete: func(e DeleteEvent) {
        if len(e.FilesDeleted) > 0 {
            log.Println("Files deleted, updating backup...")
            // updateBackup()
        }
    },
}
fs.Start()
defer fs.Stop()
```

### Example 3: Project Structure Monitor

Monitor a source code project recursively:

```go
opts := NewOptions()
opts.Recursive = true
opts.RecursiveDepth = 5

fs := &FsWatcher{
    Path:    "./my_project",
    Options: opts,
    OnCreate: func(e CreateEvent) {
        for _, file := range e.FilesCreated {
            if strings.HasSuffix(file, ".go") {
                log.Printf("New Go file: %s\n", file)
                // Run linter, formatter, etc.
            }
        }
    },
    OnChange: func(e Event) {
        log.Printf("Project structure changed at: %s\n", e.Path)
    },
}
fs.Start()
defer fs.Stop()
```

### Example 4: Temporary Directory Cleaner

Watch a temp directory and clean up old files:

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
defer cancel()

fs := &FsWatcher{
    Path:    "./temp",
    Options: NewOptions(),
    OnCreate: func(e CreateEvent) {
        // Log file creations
        log.Printf("Temp files created: %d\n", len(e.FilesCreated))
    },
}

fs.StartWithContext(ctx)

// Watcher will automatically stop after 1 hour
```

### Example 5: Conditional Watching

Watch different paths based on conditions:

```go
watchPath := "./development"
if isProduction {
    watchPath = "./production"
}

opts := NewOptions()
opts.Recursive = true
opts.Verbose = true

fs := &FsWatcher{
    Path:    watchPath,
    Options: opts,
    OnChange: func(e Event) {
        switch e.Type {
        case Create:
            log.Println("Something was created")
        case Delete:
            log.Println("Something was deleted")
        case Modify:
            log.Println("Something was modified")
        case Rename:
            log.Println("Something was renamed")
        }
    },
}
fs.Start()
```

## How It Works

1. **Initialization**: When `Start()` or `StartWithContext()` is called, the watcher takes an initial snapshot of the file system using `readDir()` or `readRecursive()`.

2. **Polling Loop**: A goroutine runs in the background with a ticker that fires every 500 milliseconds.

3. **Snapshot Comparison**: On each tick, the watcher:
   - Takes a new snapshot of the file system
   - Compares it with the previous snapshot using `diffFsInfo()`
   - Identifies added and removed files/directories

4. **Event Dispatch**: Based on the differences found:
   - Filters results according to `DirsOnly` or `FilesOnly` options
   - Calls the appropriate callbacks (`OnCreate`, `OnDelete`, etc.)
   - Calls the universal `OnChange` callback if registered

5. **State Update**: Updates the internal state with the new snapshot for the next comparison cycle.

6. **Shutdown**: When `Stop()` is called or the context is cancelled:
   - The ticker stops
   - The goroutine exits cleanly
   - The `running` flag is set to false

## Limitations

- **Polling-Based**: Uses 500ms polling intervals, so very rapid changes might be missed between polls
- **No Rename Detection**: Rename operations are detected as delete + create
- **No Modify Detection**: Currently doesn't detect file content modifications, only structural changes
- **Memory Usage**: Keeps file system state in memory, which can grow with large directory trees

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is open source and available under the MIT License.
