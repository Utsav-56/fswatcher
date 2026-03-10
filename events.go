// package watcher provides file system monitoring capabilities with detailed event tracking.
// This file defines all event types and structures used throughout the watcher.
package watcher

// EventType represents the type of file system event that occurred.
type EventType int

// Event type constants representing different file system operations.
const (
	// Create indicates that a new file or directory was created.
	Create EventType = iota
	// Delete indicates that a file or directory was removed.
	Delete
	// Rename indicates that a file or directory was renamed.
	Rename
	// Modify indicates that a file or directory was modified.
	Modify
)

// Event represents a base file system event with common fields shared by all event types.
// It can be used directly for generic change detection or embedded in more specific event types.
//
// Example:
//
//	e := Event{
//		Type:  Create,
//		Path:  "/path/to/file",
//		IsDir: false,
//	}
type Event struct {
	// Type specifies the kind of file system operation (Create, Delete, Rename, or Modify).
	Type EventType
	// Path is the file system path where the event occurred.
	Path string
	// IsDir indicates whether the event target is a directory (true) or a file (false).
	IsDir bool
}

// CreateEvent represents a creation event containing lists of newly created directories and files.
// It embeds the base Event struct and provides detailed information about what was created.
//
// Example usage in callback:
//
//	OnCreate: func(e CreateEvent) {
//		fmt.Printf("Created dirs: %v\n", e.DirsCreated)
//		fmt.Printf("Created files: %v\n", e.FilesCreated)
//	}
type CreateEvent struct {
	Event
	// DirsCreated is a list of full paths to directories that were created.
	DirsCreated []string
	// FilesCreated is a list of full paths to files that were created.
	FilesCreated []string
}

// DeleteEvent represents a deletion event containing lists of removed directories and files.
// It embeds the base Event struct and provides detailed information about what was deleted.
//
// Example usage in callback:
//
//	OnDelete: func(e DeleteEvent) {
//		fmt.Printf("Deleted dirs: %v\n", e.DirsDeleted)
//		fmt.Printf("Deleted files: %v\n", e.FilesDeleted)
//	}
type DeleteEvent struct {
	Event
	// DirsDeleted is a list of full paths to directories that were deleted.
	DirsDeleted []string
	// FilesDeleted is a list of full paths to files that were deleted.
	FilesDeleted []string
}

// RenameEvent represents a rename or move operation for a file or directory.
// It embeds the base Event struct and provides both the old and new paths.
//
// Example usage in callback:
//
//	OnRename: func(e RenameEvent) {
//		fmt.Printf("Renamed: %s -> %s\n", e.OldPath, e.NewPath)
//	}
type RenameEvent struct {
	Event
	// OldPath is the original path before the rename operation.
	OldPath string
	// NewPath is the new path after the rename operation.
	NewPath string
}

// ModifyEvent represents a modification event for files or directories.
// Note: If a directory is renamed, this event is not triggered even though renaming
// is technically a modification. Use RenameEvent for rename operations.
//
// Example usage in callback:
//
//	OnModify: func(e ModifyEvent) {
//		fmt.Printf("Modified dirs: %v\n", e.DirsModified)
//		fmt.Printf("Modified files: %v\n", e.FilesModified)
//	}
type ModifyEvent struct {
	Event
	// DirsModified is a list of full paths to directories that were modified.
	DirsModified []string
	// FilesModified is a list of full paths to files that were modified.
	FilesModified []string
}
