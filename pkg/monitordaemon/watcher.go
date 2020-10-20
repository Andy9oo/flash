package monitordaemon

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type watcher struct {
	*fsnotify.Watcher
}

func newWatcher() *watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	return &watcher{w}
}

func (w *watcher) addDir(dir string) error {
	addDir := func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsDir() {
			if len(fi.Name()) > 0 && fi.Name()[0:1] == "." {
				return filepath.SkipDir
			}
			return w.Add(path)
		}
		return nil
	}

	w.Add(dir)
	if err := filepath.Walk(dir, addDir); err != nil {
		return err
	}

	return nil
}

func (w *watcher) removeDir(dir string) error {
	removeDir := func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsDir() {
			return w.Remove(path)
		}
		return nil
	}

	w.Remove(dir)
	if err := filepath.Walk(dir, removeDir); err != nil {
		return err
	}

	return nil
}
