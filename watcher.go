// package watcher implements a polling-based file system watcher.
package watcher

import (
	"context"
	"os"
	"sync"
	"time"
)

// FsWatcher is the main file system watcher that monitors a directory for changes.
// It uses a polling mechanism (checking every 500ms by default) to detect file system
// changes and triggers appropriate callbacks.
//
// Example:
//
//	fs := &FsWatcher{
//		Path:    "./watch_dir",
//		Options: NewOptions(),
//		OnCreate: func(e CreateEvent) {
//			fmt.Printf("Created: %v\n", e.DirsCreated)
//		},
//		OnDelete: func(e DeleteEvent) {
//			fmt.Printf("Deleted: %v\n", e.DirsDeleted)
//		},
//	}
//	fs.Start()
//	defer fs.Stop()
type FsWatcher struct {
	// Path is the root directory to watch for changes.
	Path string
	// Options configures the watcher behavior (recursive, depth, filters, etc.).
	Options *Options
	// OnCreate is called when new files or directories are created.
	OnCreate func(e CreateEvent)
	// OnDelete is called when files or directories are deleted.
	OnDelete func(e DeleteEvent)
	// OnRename is called when files or directories are renamed or moved.
	OnRename func(e RenameEvent)
	// OnModify is called when files or directories are modified.
	OnModify func(e ModifyEvent)

	// OnChange is a universal callback triggered for all events. The event type
	// can be determined by checking the Type field in the Event struct.
	// This is useful when you want to handle all events in one place.
	OnChange func(e Event)

	// internal fields
	// fsInfo stores the internal state of the path to compare with new state
	// for determining what events to trigger.
	fsInfo FsInfo
	// running tracks whether the watcher is currently active.
	running bool
	// mu protects concurrent access to the running field.
	mu sync.Mutex
	// ticker provides the polling mechanism for checking file system changes.
	ticker *time.Ticker
	// cancelFunc allows stopping the watcher goroutine.
	cancelFunc context.CancelFunc
}

// FsInfo represents a snapshot of the file system state at a point in time.
// It stores information about all directories and files in separate maps for efficient lookups.
type FsInfo struct {
	// Dirs maps directory paths to their os.FileInfo metadata.
	Dirs map[string]os.FileInfo
	// Files maps file paths to their os.FileInfo metadata.
	Files map[string]os.FileInfo
}

// DirDiff represents the difference in directories between two file system snapshots.
type DirDiff struct {
	// Added contains paths of directories that were newly created.
	Added []string
	// Removed contains paths of directories that were deleted.
	Removed []string
}

// FileDiff represents the difference in files between two file system snapshots.
type FileDiff struct {
	// Added contains paths of files that were newly created.
	Added []string
	// Removed contains paths of files that were deleted.
	Removed []string
}

// StartWithContext starts the file system watcher with the provided context.
// It initializes the watcher state, takes an initial snapshot of the file system,
// and begins polling for changes every 500 milliseconds.
//
// The watcher runs in a background goroutine and will continue monitoring until
// either Stop() is called or the provided context is cancelled.
//
// Parameters:
//   - ctx: A context that can be used to stop the watcher externally
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	fs := &FsWatcher{
//		Path: "./watch_dir",
//		Options: NewOptions(),
//		OnChange: func(e Event) {
//			fmt.Println("Change detected!")
//		},
//	}
//	fs.StartWithContext(ctx)
//
// Note: If the watcher is already running, this method returns immediately
// without starting a new watcher instance.
func (fs *FsWatcher) StartWithContext(ctx context.Context) {
	fs.mu.Lock()
	if fs.running {
		fs.mu.Unlock()
		return
	}
	fs.running = true
	fs.mu.Unlock()

	// initialize fsInfo snapshot
	if fs.Options.Recursive {
		depth := fs.Options.RecursiveDepth
		if depth == -1 {
			depth = 10000 // effectively infinite
		}
		fs.fsInfo = readRecursive(fs.Path, depth)
	} else {
		fs.fsInfo = readDir(fs.Path)
	}

	fs.ticker = time.NewTicker(500 * time.Millisecond) // polling interval
	ctx, cancel := context.WithCancel(ctx)
	fs.cancelFunc = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				fs.ticker.Stop()
				fs.mu.Lock()
				fs.running = false
				fs.mu.Unlock()
				return
			case <-fs.ticker.C:
				fs.scanAndEmit()
			}
		}
	}()
}

func (fs *FsWatcher) Start() {
	fs.StartWithContext(context.Background())
}

// Stop gracefully stops the file system watcher by cancelling its context.
// It stops the polling ticker and sets the running state to false.
//
// This method is safe to call multiple times. If the watcher is not running,
// it does nothing.
//
// Example:
//
//	fs := &FsWatcher{...}
//	fs.Start()
//
//	// Do some work...
//
//	fs.Stop() // Gracefully stop monitoring
func (fs *FsWatcher) Stop() {
	if fs.cancelFunc != nil {
		fs.cancelFunc()
	}
}

// scanAndEmit is an internal method that performs a scan of the monitored directory,
// compares it with the previous state, and triggers appropriate callbacks for detected changes.
//
// This method is called periodically by the polling ticker (every 500ms by default).
// It respects the DirsOnly and FilesOnly options to filter events accordingly.
//
// The method updates the internal file system state after each scan to track changes
// over time.
func (fs *FsWatcher) scanAndEmit() {
	var newFs FsInfo
	if fs.Options.Recursive {
		depth := fs.Options.RecursiveDepth
		if depth == -1 {
			depth = 10000
		}
		newFs = readRecursive(fs.Path, depth)
	} else {
		newFs = readDir(fs.Path)
	}

	addedDirs, removedDirs, addedFiles, removedFiles := diffFsInfo(fs.fsInfo, newFs)

	// Update state
	fs.fsInfo = newFs

	if fs.Options.DirsOnly {
		addedFiles = nil
		removedFiles = nil
	}
	if fs.Options.FilesOnly {
		addedDirs = nil
		removedDirs = nil
	}

	// trigger events
	if len(addedDirs) > 0 || len(addedFiles) > 0 {
		if fs.OnCreate != nil {
			fs.OnCreate(CreateEvent{
				Event:        Event{Type: Create, Path: fs.Path, IsDir: false},
				DirsCreated:  addedDirs,
				FilesCreated: addedFiles,
			})
		}
	}

	if len(removedDirs) > 0 || len(removedFiles) > 0 {
		if fs.OnDelete != nil {
			fs.OnDelete(DeleteEvent{
				Event:        Event{Type: Delete, Path: fs.Path, IsDir: false},
				DirsDeleted:  removedDirs,
				FilesDeleted: removedFiles,
			})
		}
	}

	if len(addedDirs)+len(removedDirs)+len(addedFiles)+len(removedFiles) > 0 {
		if fs.OnChange != nil {
			fs.OnChange(Event{Type: Modify, Path: fs.Path, IsDir: false})
		}
	}
}
