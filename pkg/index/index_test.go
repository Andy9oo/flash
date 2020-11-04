package index

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func setup() {
	tmp := os.TempDir()
	flashhome := tmp + "flash/"

	viper.Set("flashhome", flashhome)
	viper.Set("indexpath", flashhome+"index")
	viper.Set("dirs", []string{})
	viper.Set("tikapath", flashhome+"tika.jar")
	viper.Set("tikaport", "9999")
	viper.Set("blacklist", []string{})
}

func TestCreate(t *testing.T) {
	setup()
}
