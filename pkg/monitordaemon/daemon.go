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

// Init initializes and returns the monitor
func Init(dirs []string) *MonitorDaemon {
	d, err := daemon.New("flashmonitor", "flashmonitor watches for file changes", daemon.SystemDaemon)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range dirs {
		watcher.Add(d)
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

				switch event.Op {
				case fsnotify.Create:
					fmt.Println(event.Name, "Created")
				case fsnotify.Write:
					fmt.Println(event.Name, "Modified")
				case fsnotify.Remove:
					fallthrough
				case fsnotify.Rename:
					fmt.Println(event.Name, "DELETE")
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
	}()
	killSignal := <-interrupt
	fmt.Println("Got signal:", killSignal)
}

// Add adds the given directory to the watcher
func (d *MonitorDaemon) Add(dir string) {
	err := d.watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
}
