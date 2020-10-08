package monitordaemon

import (
	"context"
	"log"
	"os"

	"github.com/google/go-tika/tika"
	"github.com/spf13/viper"
)

type tikaServer struct {
	*tika.Server
}

func getTikaServer() *tikaServer {
	tikapath := viper.GetString("tikapath")
	tikaport := viper.GetString("tikaport")

	_, err := os.Stat(tikapath)
	if err != nil {
		err := tika.DownloadServer(context.Background(), "1.21", tikapath)
		if err != nil {
			log.Fatal(err)
		}
	}

	s, err := tika.NewServer(tikapath, tikaport)
	if err != nil {
		log.Fatal(err)
	}

	return &tikaServer{s}
}

func (ts *tikaServer) start() {
	err := ts.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func (ts *tikaServer) stop() {
	ts.Stop()
}
