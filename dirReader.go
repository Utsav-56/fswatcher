// Package watcher provides file system monitoring capabilities.
// This file contains internal functions for reading directory structures,
// taking filesystem snapshots, and computing differences between snapshots.
package watcher

import (
	"log"
	"os"
	"path/filepath"
)

// null is an empty struct used as a placeholder value in maps to represent the presence of a key.
// Using struct{} as a value type is a common Go idiom for sets, as it occupies zero bytes of storage.
//
// this is performant for our use case because we only care about the presence of a path in the map, not any associated metadata.
type null struct{}

// FsInfo represents a snapshot of the filesystem state at a specific point in time.
// It stores directory and file information in separate maps for efficient lookups
// and comparison.
//
// The maps use full file paths as keys and os.FileInfo as values, which includes
// metadata like modification time, size, and permissions.
//
// This structure is used to compare filesystem states before and after events to
// determine exactly what changed.
type FsInfo struct {
	// Dirs maps directory paths to their file information metadata.
	Dirs map[string]null
	// Files maps file paths to their file information metadata.
	Files map[string]null
}

// pathDiff represents the differences between two filesystem snapshots.
// It contains lists of paths that were added or removed for both directories and files.
//
// This is an internal type used during diff computation. The results are converted
// to event types (CreateEvent, DeleteEvent) before being sent to user callbacks.
type pathDiff struct {
	// addedDirs contains paths of directories that exist in the new snapshot but not the old.
	addedDirs []string
	// removedDirs contains paths of directories that exist in the old snapshot but not the new.
	removedDirs []string
	// addedFiles contains paths of files that exist in the new snapshot but not the old.
	addedFiles []string
	// removedFiles contains paths of files that exist in the old snapshot but not the new.
	removedFiles []string
}

// newFsInfo creates a new FsInfo instance with pre-allocated maps.
// It allocates capacity of 128 for directories and 256 for files to reduce
// reallocations during scanning.
//
// These capacity values are chosen based on typical directory structures.
// For larger directories, the maps will automatically grow as needed.
//
// Returns:
//   - FsInfo: A new filesystem snapshot structure with empty but pre-allocated maps
func newFsInfo() FsInfo {
	return FsInfo{
		Dirs:  make(map[string]null, 128),
		Files: make(map[string]null, 256),
	}
}

// readDir reads a single directory (non-recursive) and returns a snapshot of its contents.
// It scans only the immediate children of the specified directory without descending
// into subdirectories.
//
// This function is used when the watcher is configured with Recursive=false.
//
// Parameters:
//   - path: The directory path to read
//
// Returns:
//   - FsInfo: A snapshot containing all directories and files in the immediate directory
func readDir(path string) FsInfo {
	info := newFsInfo()
	scanDir(path, false, -1, &info)
	return info
}

// readRecursive reads a directory and all its subdirectories up to the specified depth.
// It returns a snapshot containing all directories and files in the entire tree.
//
// This function is used when the watcher is configured with Recursive=true.
//
// Parameters:
//   - path: The root directory path to start scanning from
//   - depth: Maximum depth to scan. Use -1 for unlimited depth, 0 to scan only the root,
//     1 to include immediate children, 2 for grandchildren, etc.
//
// Returns:
//   - FsInfo: A snapshot containing all directories and files within the depth limit
func readRecursive(path string, depth int) FsInfo {
	info := newFsInfo()
	scanDir(path, true, depth, &info)
	return info
}

// scanDir is the internal workhorse that recursively scans directories and populates
// the FsInfo structure. It's called by readDir and readRecursive.
//
// The function handles:
//   - Depth limiting for recursive scans
//   - Symlink detection and skipping (prevents infinite loops)
//   - Error handling (logs errors but continues scanning)
//   - Efficient metadata collection using os.DirEntry
//
// Symbolic links are intentionally skipped to prevent infinite loops when a symlink
// points to a parent directory. This is a common security and stability practice.
//
// Parameters:
//   - path: The directory path to scan
//   - recursive: Whether to descend into subdirectories
//   - depth: Remaining depth budget. Decrements with each level. -1 means unlimited.
//   - info: Pointer to the FsInfo structure being populated
func scanDir(path string, recursive bool, depth int, info *FsInfo) {

	if depth == 0 {
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("watcher: read error %s: %v", path, err)
		return
	}

	for _, e := range entries {

		// avoid symlink loops
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}

		fullPath := filepath.Join(path, e.Name())

		if e.IsDir() {
			info.Dirs[fullPath] = null{}

			if recursive {

				nextDepth := depth
				if depth > 0 {
					nextDepth = depth - 1
				}

				scanDir(fullPath, recursive, nextDepth, info)
			}

		} else {
			info.Files[fullPath] = null{}
		}
	}
}

// diffFsInfo compares two filesystem snapshots and returns the differences.
// It computes four sets: added directories, removed directories, added files, and removed files.
//
// The algorithm is straightforward:
//   - Added: Items in newFs but not in oldFs
//   - Removed: Items in oldFs but not in newFs
//
// Note: This function does not detect modifications to existing files (size or time changes).
// It only detects additions and deletions. Modification detection could be added in future versions
// by comparing file metadata.
//
// Parameters:
//   - oldFs: The previous filesystem snapshot
//   - newFs: The current filesystem snapshot
//
// Returns:
//   - pathDiff: A structure containing lists of added and removed directories and files
func diffFsInfo(oldFs, newFs FsInfo) pathDiff {

	diff := pathDiff{
		addedDirs:    make([]string, 0, 16),
		removedDirs:  make([]string, 0, 16),
		addedFiles:   make([]string, 0, 32),
		removedFiles: make([]string, 0, 32),
	}

	// detect new directories
	for p := range newFs.Dirs {
		if _, ok := oldFs.Dirs[p]; !ok {
			diff.addedDirs = append(diff.addedDirs, p)
		}
	}

	// detect removed directories
	for p := range oldFs.Dirs {
		if _, ok := newFs.Dirs[p]; !ok {
			diff.removedDirs = append(diff.removedDirs, p)
		}
	}

	// detect new files
	for p := range newFs.Files {
		if _, ok := oldFs.Files[p]; !ok {
			diff.addedFiles = append(diff.addedFiles, p)
		}
	}

	// detect removed files
	for p := range oldFs.Files {
		if _, ok := newFs.Files[p]; !ok {
			diff.removedFiles = append(diff.removedFiles, p)
		}
	}

	return diff
}
