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

// BlacklistPatterns is a list of patterns
type BlacklistPatterns struct {
	Patterns []string
}

// DirList is a list of added directories
type DirList struct {
	Dirs []string
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
	defer h.dmn.lock.Unlock()

	path := h.dmn.index.GetPath()
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	h.dmn.index = index.Load(path)

	for _, d := range viper.GetStringSlice("dirs") {
		h.dmn.watcher.Remove(d)
	}

	h.dmn.index.ResetBlacklist()

	viper.Set("dirs", []string{})
	viper.Set("blacklist", []string{})
	return nil
}

// Add adds a directory to the watcher
func (h *Handler) Add(dir string, res *bool) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}

	h.dmn.lock.Lock()
	defer h.dmn.lock.Unlock()
	dirs := viper.GetStringSlice("dirs")

	for _, d := range dirs {
		if d == dir {
			return errors.New("Directory already in index")
		}
	}

	viper.Set("dirs", append(dirs, dir))

	h.dmn.watcher.addDir(dir)
	go h.dmn.index.Add(dir, h.dmn.lock)

	return nil
}

// Remove removes a directory from the watcher
func (h *Handler) Remove(dir string, res *bool) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}

	h.dmn.lock.Lock()
	defer h.dmn.lock.Unlock()

	dirs := viper.GetStringSlice("dirs")
	for i := range dirs {
		if dirs[i] == dir {
			dirs[i] = dirs[len(dirs)-1]
			dirs[len(dirs)-1] = ""
			dirs = dirs[:len(dirs)-1]
			break
		}
	}
	viper.Set("dirs", dirs)

	h.dmn.watcher.removeDir(dir)
	h.dmn.index.Delete(dir)
	return nil
}

// BlacklistAdd adds a pattern to the blacklist
func (h *Handler) BlacklistAdd(pattern string, res *bool) error {
	h.dmn.lock.Lock()
	defer h.dmn.lock.Unlock()

	err := h.dmn.index.Blacklist(pattern)
	if err != nil {
		return err
	}

	blacklist := viper.GetStringSlice("blacklist")
	viper.Set("blacklist", append(blacklist, pattern))
	return nil
}

// BlacklistRemove adds a pattern to the blacklist
func (h *Handler) BlacklistRemove(pattern string, res *bool) error {
	h.dmn.lock.Lock()
	defer h.dmn.lock.Unlock()

	blacklists := viper.GetStringSlice("blacklist")
	for i := range blacklists {
		if blacklists[i] == pattern {
			blacklists[i] = blacklists[len(blacklists)-1]
			blacklists[len(blacklists)-1] = ""
			blacklists = blacklists[:len(blacklists)-1]
			break
		}
	}
	viper.Set("blacklist", blacklists)
	h.dmn.index.RemoveBlacklist(pattern)
	return nil
}

// BlacklistGet returns a list of all blacklisted patterns
func (h *Handler) BlacklistGet(_ string, res *BlacklistPatterns) error {
	res.Patterns = h.dmn.index.GetBlacklist()
	return nil
}

// List returns all directories added to the index
func (h *Handler) List(_ string, res *DirList) error {
	res.Dirs = viper.GetStringSlice("dirs")
	return nil
}
