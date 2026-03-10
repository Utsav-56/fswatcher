// package watcher provides configuration options for the file system watcher.
package watcher

// Options configures the behavior of the FsWatcher, including recursion settings
// and filtering options for directories and files.
//
// Example:
//
//	opts := NewOptions()
//	opts.Recursive = true
//	opts.RecursiveDepth = 3
//	opts.DirsOnly = true
type Options struct {
	// Recursive enables recursive monitoring of subdirectories when set to true.
	Recursive bool
	// DirsOnly, when true, monitors only directories and ignores file changes.
	DirsOnly bool
	// Verbose enables detailed logging output when set to true.
	Verbose bool
	// FilesOnly, when true, monitors only files and ignores directory changes.
	// Note: If both DirsOnly and FilesOnly are false, the watcher monitors both
	// files and directories (default behavior).
	FilesOnly bool
	// RecursiveDepth limits the depth of recursive scanning. A value of -1 means
	// infinite depth (scans all subdirectories). Only effective when Recursive is true.
	RecursiveDepth int
}

// NewOptions creates and returns a new Options struct with default values.
// Default configuration:
//   - Recursive: false (only watches the specified directory, not subdirectories)
//   - DirsOnly: false (watches both files and directories)
//   - FilesOnly: false (watches both files and directories)
//   - Verbose: false (minimal logging)
//   - RecursiveDepth: -1 (infinite depth when Recursive is enabled)
//
// Example:
//
//	opts := NewOptions()
//	// opts is ready to use with default settings
//	// or customize:
//	opts.Recursive = true
//	opts.RecursiveDepth = 5
func NewOptions() *Options {
	return &Options{
		Recursive:      false,
		DirsOnly:       false,
		Verbose:        false,
		FilesOnly:      false,
		RecursiveDepth: -1, // -1 means infinite depth
	}
}
