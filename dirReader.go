package watcher

import (
	"log"
	"os"
	"path/filepath"
)

type FsInfo struct {
	Dirs  map[string]os.FileInfo
	Files map[string]os.FileInfo
}

type pathDiff struct {
	addedDirs    []string
	removedDirs  []string
	addedFiles   []string
	removedFiles []string
}

func newFsInfo() FsInfo {
	return FsInfo{
		Dirs:  make(map[string]os.FileInfo, 128),
		Files: make(map[string]os.FileInfo, 256),
	}
}

func readDir(path string) FsInfo {
	info := newFsInfo()
	scanDir(path, false, -1, &info)
	return info
}

func readRecursive(path string, depth int) FsInfo {
	info := newFsInfo()
	scanDir(path, true, depth, &info)
	return info
}

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

			fi, err := e.Info()
			if err != nil {
				continue
			}

			info.Dirs[fullPath] = fi

			if recursive {

				nextDepth := depth
				if depth > 0 {
					nextDepth = depth - 1
				}

				scanDir(fullPath, recursive, nextDepth, info)
			}

		} else {

			fi, err := e.Info()
			if err != nil {
				continue
			}

			info.Files[fullPath] = fi
		}
	}
}

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
