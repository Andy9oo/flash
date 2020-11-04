package tika

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-tika/tika"
	"github.com/spf13/viper"
)

// Server controls a tika server
type Server struct {
	*tika.Server
}

// GetServer returns a new sever
func GetServer() *Server {
	tikapath := viper.GetString("tikapath")
	tikaport := viper.GetString("tikaport")

	_, err := os.Stat(tikapath)
	if err != nil {
		fmt.Println("Tika not found, downloading")
		err := tika.DownloadServer(context.Background(), "1.21", tikapath)
		if err != nil {
			log.Fatal(err)
		}
	}

	s, err := tika.NewServer(tikapath, tikaport)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{s}
}

// StartServer starts the tika server
func (ts *Server) StartServer() {
	err := ts.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

// StopServer stops the tika server
func (ts *Server) StopServer() {
	ts.Stop()
}
