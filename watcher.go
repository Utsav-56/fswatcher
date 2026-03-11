package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceDuration = 80 * time.Millisecond

type FsWatcher struct {
	Path    string
	Options *Options

	OnCreate func(CreateEvent)
	OnDelete func(DeleteEvent)
	OnRename func(RenameEvent)
	OnModify func(ModifyEvent)
	OnChange func(Event)

	fsInfo FsInfo

	running bool
	mu      sync.Mutex

	watcher *fsnotify.Watcher

	cancelFunc context.CancelFunc

	eventTrigger chan struct{}
}

func (fs *FsWatcher) StartWithContext(ctx context.Context) error {

	fs.mu.Lock()
	if fs.running {
		fs.mu.Unlock()
		return nil
	}
	fs.running = true
	fs.mu.Unlock()

	var err error
	fs.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	fs.eventTrigger = make(chan struct{}, 1)

	// initial scan
	if fs.Options.Recursive {
		fs.fsInfo = readRecursive(fs.Path, fs.Options.RecursiveDepth)
		fs.addRecursiveWatches(fs.Path)
	} else {
		fs.fsInfo = readDir(fs.Path)
		fs.watcher.Add(fs.Path)
	}

	ctx, cancel := context.WithCancel(ctx)
	fs.cancelFunc = cancel

	go fs.eventLoop(ctx)
	go fs.scanWorker(ctx)

	return nil
}

func (fs *FsWatcher) eventLoop(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			fs.watcher.Close()
			return

		case ev, ok := <-fs.watcher.Events:
			if !ok {
				return
			}

			// if new directory created → watch it
			if fs.Options.Recursive && ev.Op&fsnotify.Create == fsnotify.Create {

				info, err := os.Stat(ev.Name)
				if err == nil && info.IsDir() {
					fs.addRecursiveWatches(ev.Name)
				}
			}

			// trigger scan (non blocking)
			select {
			case fs.eventTrigger <- struct{}{}:
			default:
			}

		case <-fs.watcher.Errors:
			// ignore errors for now
		}
	}
}

func (fs *FsWatcher) scanWorker(ctx context.Context) {

	timer := time.NewTimer(time.Hour)
	timer.Stop()

	for {
		select {

		case <-ctx.Done():
			return

		case <-fs.eventTrigger:

			timer.Reset(debounceDuration)

		case <-timer.C:

			fs.scanAndEmit()
		}
	}
}

func (fs *FsWatcher) scanAndEmit() {

	var newFs FsInfo

	if fs.Options.Recursive {
		newFs = readRecursive(fs.Path, fs.Options.RecursiveDepth)
	} else {
		newFs = readDir(fs.Path)
	}

	diff := diffFsInfo(fs.fsInfo, newFs)

	fs.fsInfo = newFs

	if fs.Options.DirsOnly {
		diff.addedFiles = nil
		diff.removedFiles = nil
	}

	if fs.Options.FilesOnly {
		diff.addedDirs = nil
		diff.removedDirs = nil
	}

	if len(diff.addedDirs) > 0 || len(diff.addedFiles) > 0 {

		if fs.OnCreate != nil {
			fs.OnCreate(CreateEvent{
				DirsCreated:  diff.addedDirs,
				FilesCreated: diff.addedFiles,
			})
		}
	}

	if len(diff.removedDirs) > 0 || len(diff.removedFiles) > 0 {

		if fs.OnDelete != nil {
			fs.OnDelete(DeleteEvent{
				DirsDeleted:  diff.removedDirs,
				FilesDeleted: diff.removedFiles,
			})
		}
	}

	if fs.OnChange != nil &&
		(len(diff.addedDirs)+len(diff.removedDirs)+len(diff.addedFiles)+len(diff.removedFiles)) > 0 {

		fs.OnChange(Event{
			Type: Modify,
			Path: fs.Path,
		})
	}
}

func (fs *FsWatcher) addRecursiveWatches(root string) {

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {

		if err != nil {
			return nil
		}

		if d.IsDir() {
			fs.watcher.Add(path)
		}

		return nil
	})
}

func (fs *FsWatcher) Start() error {
	return fs.StartWithContext(context.Background())
}

func (fs *FsWatcher) Stop() {
	if fs.cancelFunc != nil {
		fs.cancelFunc()
	}
}
