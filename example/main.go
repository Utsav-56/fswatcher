package main

import (
	"log"

	watcher "github.com/utsav-56/fswatcher"
)

func main() {
	fs := &watcher.FsWatcher{
		Path: "./test",
		Options: &watcher.Options{
			Recursive: false,
			DirsOnly:  true,
		},
		OnCreate: func(e watcher.CreateEvent) {
			log.Printf(`Created 
dirs: %v, files: %v
			`, e.DirsCreated, e.FilesCreated)
		},
		OnDelete: func(e watcher.DeleteEvent) {
			log.Printf(`Deleted 
dirs: %v, files: %v
			`, e.DirsDeleted, e.FilesDeleted)
		},
		OnChange: func(e watcher.Event) {
			log.Printf(`Change detected in path: %v
			`, e.Path)
		},
	}
	// fs.Options.Recursive = true
	fs.Start()
	defer fs.Stop()

	select {} // block forever
}
