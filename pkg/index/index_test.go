package index

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/spf13/viper"
)

func setup() {
	tmp := os.TempDir()
	flashhome := fmt.Sprintf("%v/flash/", tmp)
	viper.Set("flashhome", flashhome)
	viper.Set("indexpath", flashhome+"index")
	viper.Set("dirs", []string{})
	viper.Set("tikapath", "../../tools/tika/tika.jar")
	viper.Set("tikaport", "9999")
	viper.Set("blacklist", []string{})
}

func TestNewIndex(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	if _, err := os.Stat(indexpath); err != nil || index == nil {
		t.Fail()
	}
}
func TestInvalidDir(t *testing.T) {
	setup()
	index := NewIndex("///")
	if _, err := os.Stat(index.GetPath()); err == nil {
		t.Fail()
	}
}

func TestAdd(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Add("./testdata/hello_world.txt", &sync.RWMutex{})
	if index.docs.NumDocs() != 1 {
		t.Fail()
	}
}

func TestSkipMissing(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Add("missing", &sync.RWMutex{})
	if index.docs.NumDocs() != 0 {
		t.Fail()
	}
}

func TestSkipHidden(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Add("./testdata/.hidden.txt", &sync.RWMutex{})
	if index.docs.NumDocs() != 0 {
		t.Fail()
	}
}

func TestSkipDuplicate(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Add("./testdata/hello_world.txt", &sync.RWMutex{})
	index.Add("./testdata/hello_world.txt", &sync.RWMutex{})
	if index.docs.NumDocs() != 1 {
		t.Fail()
	}
}

func TestRecursive(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Add("./testdata/directory", &sync.RWMutex{})
	if index.docs.NumDocs() != 2 {
		t.Fail()
	}
}

func TestSkipHiddenRecursive(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Add("./testdata/directory_with_hidden", &sync.RWMutex{})
	if index.docs.NumDocs() != 1 {
		t.Fail()
	}
}

func TestSkipBlacklisted(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Blacklist(".*\\.txt")
	index.Add("./testdata/hello_world.txt", &sync.RWMutex{})
	if index.docs.NumDocs() != 0 {
		t.Fail()
	}
}

func TestGetInfo(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Add("./testdata/directory", &sync.RWMutex{})
	info := index.GetInfo()
	if info.NumDocs != 2 || info.AvgLength != 4 {
		t.Fail()
	}
}

func TestGetDocInfoSuccess(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	file := "./testdata/hello_world.txt"

	stat, _ := os.Stat(file)
	var id uint64
	if sys, ok := stat.Sys().(*syscall.Stat_t); ok {
		id = sys.Ino
	}

	index.Add(file, &sync.RWMutex{})
	path, length, _ := index.GetDocInfo(id)
	if path != file || length != 5 {
		t.Fail()
	}
}

func TestGetDocInfoFail(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	file := "./testdata/hello_world.txt"

	stat, _ := os.Stat(file)
	var id uint64
	if sys, ok := stat.Sys().(*syscall.Stat_t); ok {
		id = sys.Ino
	}

	_, _, ok := index.GetDocInfo(id)
	if ok {
		t.Fail()
	}
}

func TestGetPath(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	if index.GetPath() != indexpath {
		t.Fail()
	}
}

func TestBlacklistAdd(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Blacklist("test")
	blacklist := index.GetBlacklist()
	if len(blacklist) != 1 || blacklist[0] != "test" {
		t.Fail()
	}
}

func TestBlacklistRemove(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Blacklist("test")
	index.RemoveBlacklist("test")
	blacklist := index.GetBlacklist()
	if len(blacklist) != 0 {
		t.Fail()
	}
}

func TestBlacklistReset(t *testing.T) {
	setup()
	indexpath := viper.GetString("indexpath")
	index := NewIndex(indexpath)
	index.Blacklist("test")
	index.ResetBlacklist()
	blacklist := index.GetBlacklist()
	if len(blacklist) != 0 {
		t.Fail()
	}
}
