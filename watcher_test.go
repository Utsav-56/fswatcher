package watcher

import "log"

func main() {
	fs := &FsWatcher{
		Path:    "./test",
		Options: NewOptions(),
		OnCreate: func(e CreateEvent) {
			log.Printf(`Created 
dirs: %v, files: %v
			`, e.DirsCreated, e.FilesCreated)
		},
		OnDelete: func(e DeleteEvent) {
			log.Printf(`Deleted 
dirs: %v, files: %v
			`, e.DirsDeleted, e.FilesDeleted)
		},
		OnChange: func(e Event) {
			log.Printf(`Change detected in path: %v
			`, e.Path)
		},
	}
	fs.Options.Recursive = true
	fs.Start()
	defer fs.Stop()

	select {} // block forever
}
