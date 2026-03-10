// package watcher provides directory scanning and comparison utilities.
package watcher

import (
	"log"
	"os"
	"path/filepath"
)

// readDir scans a single directory (non-recursively) and returns information about
// all files and subdirectories it contains. It does not traverse subdirectories.
//
// Parameters:
//   - path: The directory path to scan
//
// Returns:
//   - FsInfo: A struct containing maps of directories and files with their FileInfo
//
// Example:
//
//	info := readDir("/path/to/dir")
//	fmt.Printf("Found %d directories and %d files\n",
//		len(info.Dirs), len(info.Files))
//
// If an error occurs while reading the directory, it logs the error and returns
// an empty FsInfo struct.
func readDir(path string) FsInfo {
	info := FsInfo{
		Dirs:  make(map[string]os.FileInfo),
		Files: make(map[string]os.FileInfo),
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Error reading dir %s: %v", path, err)
		return info
	}

	for _, e := range entries {
		fullPath := filepath.Join(path, e.Name())
		fi, err := e.Info()
		if err != nil {
			continue
		}
		if fi.IsDir() {
			info.Dirs[fullPath] = fi
		} else {
			info.Files[fullPath] = fi
		}
	}
	return info
}

// readRecursive scans a directory recursively up to a specified depth limit.
// It collects information about all files and directories in the tree.
//
// Parameters:
//   - path: The root directory path to start scanning from
//   - depth: Maximum depth to scan. Use 0 to scan nothing, positive numbers for
//     limited depth, or large numbers (e.g., 10000) for effectively unlimited depth
//
// Returns:
//   - FsInfo: A struct containing maps of all directories and files found
//
// Example:
//
//	// Scan up to 3 levels deep
//	info := readRecursive("/path/to/dir", 3)
//
//	// Effectively unlimited depth
//	info := readRecursive("/path/to/dir", 10000)
//
// The function stops at the specified depth and logs errors for directories
// it cannot read.
func readRecursive(path string, depth int) FsInfo {
	info := FsInfo{
		Dirs:  make(map[string]os.FileInfo),
		Files: make(map[string]os.FileInfo),
	}

	if depth == 0 {
		return info
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Error reading dir %s: %v", path, err)
		return info
	}

	for _, e := range entries {
		fullPath := filepath.Join(path, e.Name())
		fi, err := e.Info()
		if err != nil {
			continue
		}
		if fi.IsDir() {
			info.Dirs[fullPath] = fi
			subInfo := readRecursive(fullPath, depth-1)
			for k, v := range subInfo.Dirs {
				info.Dirs[k] = v
			}
			for k, v := range subInfo.Files {
				info.Files[k] = v
			}
		} else {
			info.Files[fullPath] = fi
		}
	}

	return info
}

// diffFsInfo compares two FsInfo snapshots and computes the differences between them.
// It identifies which directories and files were added or removed.
//
// Parameters:
//   - oldFs: The previous file system state snapshot
//   - newFs: The current file system state snapshot
//
// Returns:
//   - addedDirs: Slice of directory paths that exist in newFs but not in oldFs
//   - removedDirs: Slice of directory paths that exist in oldFs but not in newFs
//   - addedFiles: Slice of file paths that exist in newFs but not in oldFs
//   - removedFiles: Slice of file paths that exist in oldFs but not in newFs
//
// Example:
//
//	oldSnapshot := readDir("/path")
//	time.Sleep(time.Second)
//	newSnapshot := readDir("/path")
//	addedDirs, removedDirs, addedFiles, removedFiles := diffFsInfo(oldSnapshot, newSnapshot)
//	fmt.Printf("Added: %d dirs, %d files\n", len(addedDirs), len(addedFiles))
//	fmt.Printf("Removed: %d dirs, %d files\n", len(removedDirs), len(removedFiles))
func diffFsInfo(oldFs, newFs FsInfo) (addedDirs, removedDirs []string, addedFiles, removedFiles []string) {

	// dirs
	for path := range newFs.Dirs {
		if _, ok := oldFs.Dirs[path]; !ok {
			addedDirs = append(addedDirs, path)
		}
	}
	for path := range oldFs.Dirs {
		if _, ok := newFs.Dirs[path]; !ok {
			removedDirs = append(removedDirs, path)
		}
	}

	// files
	for path := range newFs.Files {
		if _, ok := oldFs.Files[path]; !ok {
			addedFiles = append(addedFiles, path)
		}
	}
	for path := range oldFs.Files {
		if _, ok := newFs.Files[path]; !ok {
			removedFiles = append(removedFiles, path)
		}
	}

	return
}
