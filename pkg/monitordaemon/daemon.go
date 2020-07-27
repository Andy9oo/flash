package monitordaemon

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/takama/daemon"
)

// MonitorDaemon has embedded daemon
type MonitorDaemon struct {
	daemon  daemon.Daemon
	watcher *fsnotify.Watcher
}

// Get returns the monitor
func Get() *MonitorDaemon {
	d, err := daemon.New("flaskmonitor", "flaskmonitor watches for file changes")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	return &MonitorDaemon{daemon: d, watcher: watcher}
}

// Install installs the daemon
func (d *MonitorDaemon) Install() (string, error) {
	return d.daemon.Install("daemon run")
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

// Watch watches for file changes in the added files
func (d *MonitorDaemon) Watch() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		for {
			select {
			case event, ok := <-d.watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-d.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	killSignal := <-interrupt
	fmt.Println("Got signal:", killSignal)
}

// Add adds the given file watcher
func (d *MonitorDaemon) Add(file string) {
	err := d.watcher.Add(file)
	if err != nil {
		log.Fatal(err)
	}
}
