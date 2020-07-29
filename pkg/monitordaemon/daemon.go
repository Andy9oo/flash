package monitordaemon

import (
	"flash/pkg/index"
	"flash/pkg/search"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/takama/daemon"
)

const port = ":9977"

// MonitorDaemon has embedded daemon
type MonitorDaemon struct {
	daemon  daemon.Daemon
	watcher *fsnotify.Watcher
	index   *index.Index
}

// Init initializes and returns the monitor
func Init(dirs []string, index *index.Index) *MonitorDaemon {
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

	return &MonitorDaemon{daemon: d, watcher: watcher, index: index}
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
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	go d.watch()
	go d.handleRequests()
	<-interrupt
	d.index.ClearMemory()
}

// Add adds the given directory to the watcher
func (d *MonitorDaemon) Add(dir string) {
	err := d.watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
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
				fmt.Println("Adding", event.Name)
				d.index.Add(event.Name)
			case fsnotify.Write:
				fmt.Println("Deleting & readding", event.Name)
				d.index.Delete(event.Name)
				d.index.Add(event.Name)
			case fsnotify.Remove:
				fallthrough
			case fsnotify.Rename:
				fmt.Println("Deleting", event.Name)
				fmt.Println(event)
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

	handleSearch := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			if err := req.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}

			query := req.FormValue("query")
			n, err := strconv.Atoi(req.FormValue("num_results"))
			if err != nil {
				fmt.Fprintf(w, "err: %v", err)
				return
			}

			engine := search.NewEngine(d.index)
			results := engine.Search(query, n)

			var b strings.Builder
			for i, result := range results {
				path, _, _ := d.index.GetDocInfo(result.ID)
				fmt.Fprintf(&b, "%v. %v (%v)\n", i+1, path, result.Score)
			}

			fmt.Fprint(w, b.String())
		} else {
			fmt.Fprintf(w, "Sorry, POST methods are supported.")
		}
	}

	http.HandleFunc("/search", handleSearch)
	http.ListenAndServe(port, nil)
}
