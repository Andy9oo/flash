package monitordaemon

import (
	"flash/pkg/index"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/takama/daemon"
)

const port = ":9977"

// MonitorDaemon has embedded daemon
type MonitorDaemon struct {
	daemon  daemon.Daemon
	watcher *watcher
	index   *index.Index
	dirs    []string
}

// Init initializes and returns the monitor
func Init() *MonitorDaemon {
	d, err := daemon.New("flashmonitor", "flashmonitor watches for file changes", daemon.SystemDaemon)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	return &MonitorDaemon{daemon: d}
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
	dirs := viper.GetStringSlice("dirs")
	for _, dir := range dirs {
		d.index.Add(dir)
		d.watcher.addDir(dir)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	go d.watch()
	go d.handleRequests()
	<-interrupt
	d.index.ClearMemory()
	viper.WriteConfig()
}

// watch watches for file changes in the added files
func (d *MonitorDaemon) watch() {
	for {
		select {
		case event, ok := <-d.watcher.Events:
			if !ok {
				return
			}
			switch event.Op {
			case fsnotify.Create:
				d.index.Add(event.Name)
			case fsnotify.Write, fsnotify.Chmod:
				d.index.Delete(event.Name)
				d.index.Add(event.Name)
			case fsnotify.Rename, fsnotify.Remove:
				d.index.Delete(event.Name)
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
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:12345")
	if err != nil {
		log.Fatal(err)
	}
	inbound, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	h := &Handler{d}
	rpc.Register(h)
	for {
		rpc.Accept(inbound)
	}
}
