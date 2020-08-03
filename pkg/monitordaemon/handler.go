package monitordaemon

import (
	"errors"
	"flash/pkg/index"
	"flash/pkg/search"
	"os"

	"github.com/spf13/viper"
)

// Handler is used to handle rpc requests
type Handler struct {
	dmn *MonitorDaemon
}

// Query is used to communicate search queries
type Query struct {
	Str string
	N   int
}

// Results is returned from a search
type Results struct {
	Paths  []string
	Scores []float64
}

// Search searches the index for a query
func (h *Handler) Search(q *Query, res *Results) error {
	engine := search.NewEngine(h.dmn.index)

	h.dmn.lock.RLock()
	results := engine.Search(q.Str, q.N)
	for _, val := range results {
		path, _, _ := h.dmn.index.GetDocInfo(val.ID)
		res.Paths = append(res.Paths, path)
		res.Scores = append(res.Scores, val.Score)
	}
	h.dmn.lock.RUnlock()
	return nil
}

// Reset resets the index, and removes all directories from the watcher
func (h *Handler) Reset(confirmation string, res *bool) error {
	h.dmn.lock.Lock()
	path := h.dmn.index.GetPath()
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	h.dmn.index = index.Load(path)

	for _, d := range viper.GetStringSlice("dirs") {
		h.dmn.watcher.Remove(d)
	}

	viper.Set("dirs", []string{})
	h.dmn.lock.Unlock()
	return nil
}

// Add adds a directory to the watcher
func (h *Handler) Add(dir string, res *bool) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}

	h.dmn.lock.Lock()
	dirs := viper.GetStringSlice("dirs")

	for _, d := range dirs {
		if d == dir {
			return errors.New("Directory already in index")
		}
	}

	viper.Set("dirs", append(dirs, dir))

	h.dmn.index.Add(dir)
	err = h.dmn.watcher.addDir(dir)
	if err != nil {
		return err
	}
	h.dmn.lock.Unlock()
	return nil
}

// Remove removes a directory from the watcher
func (h *Handler) Remove(dir string, res *bool) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}

	h.dmn.lock.Lock()
	dirs := viper.GetStringSlice("dirs")
	for i := range dirs {
		if dirs[i] == dir {
			dirs[i] = dirs[len(dirs)-1]
			dirs[len(dirs)-1] = ""
			dirs = dirs[:len(dirs)-1]
			break
		}
	}
	viper.Set("dirs", append(dirs, dir))

	h.dmn.index.Delete(dir)
	err = h.dmn.watcher.removeDir(dir)
	if err != nil {
		return err
	}
	h.dmn.lock.Unlock()
	return nil
}
