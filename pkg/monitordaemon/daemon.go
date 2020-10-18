package monitordaemon

import (
	"flash/pkg/index"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/takama/daemon"
)

const port = ":9977"

// MonitorDaemon has embedded daemon
type MonitorDaemon struct {
	daemon     daemon.Daemon
	watcher    *watcher
	index      *index.Index
	lock       *sync.RWMutex
	tikaServer *tikaServer
	dirs       []string
}

// Init initializes and returns the monitor
func Init() *MonitorDaemon {
	d, err := daemon.New("flashmonitor", "flashmonitor watches for file changes", daemon.SystemDaemon)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	return &MonitorDaemon{daemon: d, lock: &sync.RWMutex{}}
}

// Install installs the daemon
func (d *MonitorDaemon) Install() (string, error) {
	return d.daemon.Install("daemon run --config " + viper.ConfigFileUsed())
}

// Remove removes the daemon
func (d *MonitorDaemon) Remove() (string, error) {
	return d.daemon.Remove()
}

// Status returns the status of the daemon
func (d *MonitorDaemon) Status() (string, error) {
	return d.daemon.Status()
}

// Start starts the daemon
func (d *MonitorDaemon) Start() (string, error) {
	return d.daemon.Start()
}

// Stop stops the daemon
func (d *MonitorDaemon) Stop() (string, error) {
	return d.daemon.Stop()
}

// Run starts the services which the daemon controls
func (d *MonitorDaemon) Run() {
	d.index = index.Load(viper.GetString("indexpath"))

	d.watcher = newWatcher()

	d.tikaServer = getTikaServer()
	d.tikaServer.start()

	dirs := viper.GetStringSlice("dirs")

	for _, dir := range dirs {
		d.watcher.addDir(dir)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	go d.watch()
	go d.handleRequests()
	<-interrupt
	d.lock.Lock()
	d.index.ClearMemory()
	d.tikaServer.stop()
	viper.WriteConfig()
	d.lock.Unlock()
}

// watch watches for file changes in the added files
func (d *MonitorDaemon) watch() {
	for {
		select {
		case event, ok := <-d.watcher.Events:
			if !ok {
				return
			}
			log.Println(event)
			switch event.Op {
			case fsnotify.Create:
				stat, err := os.Stat(event.Name)
				if err == nil && stat.IsDir() {
					d.lock.Lock()
					d.watcher.addDir(event.Name)
					d.lock.Unlock()
				}
				fallthrough
			case fsnotify.Write, fsnotify.Chmod:
				d.index.Add(event.Name, d.lock)
			case fsnotify.Rename, fsnotify.Remove:
				if _, err := os.Stat(event.Name); err != nil {
					d.lock.Lock()
					d.index.Delete(event.Name)
					d.lock.Unlock()
				}
			default:
				fmt.Println(event)
			}
		case err, ok := <-d.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (d *MonitorDaemon) handleRequests() {
	fmt.Println("Handling")
	h := &Handler{d}
	err := rpc.Register(h)
	if err != nil {
		fmt.Println(err)
	}
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	log.Printf("Serving RPC server on port %d", 1234)
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal("Error serving: ", err)
	}
}
