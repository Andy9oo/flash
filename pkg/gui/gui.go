package gui

import (
	"flash/pkg/monitordaemon"
	"fmt"
	"log"
	"net/rpc"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const escape uint = 65307

func Display() {
	// Initialize GTK
	gtk.Init(nil)
	// Create Window
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle("Flash 🔍")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	win.SetDecorated(false)
	win.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)

	title, _ := gtk.LabelNew("Spotlight 🔍")

	col, _ := gtk.BoxNew(gtk.Orientation(gtk.ORIENTATION_VERTICAL), 2)
	row, _ := gtk.BoxNew(gtk.Orientation(gtk.ORIENTATION_HORIZONTAL), 2)
	results, _ := gtk.BoxNew(gtk.Orientation(gtk.ORIENTATION_VERTICAL), 2)

	b, _ := gtk.ButtonNewWithLabel("Search")
	b.SetSizeRequest(100, 50)

	entry, _ := gtk.EntryNew()
	entry.SetSizeRequest(500, 50)

	// Add events
	b.Connect("clicked", func() {
		handleSearch(entry, results)
	})

	entry.Connect("key_release_event", func() {
		win.Resize(600, 50)
		handleSearch(entry, results)
	})

	win.Connect("key_press_event", func(_ *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{Event: ev}
		if keyEvent.KeyVal() == escape {
			win.Destroy()
		}
	})

	col.Add(title)
	col.Add(row)
	col.Add(results)

	// Add elements
	row.Add(entry)
	row.Add(b)

	win.Add(col)

	// Setup window
	win.SetDefaultSize(600, 50)
	win.SetResizable(false)
	win.ShowAll()
	gtk.Main()
}

func handleSearch(entry *gtk.Entry, resultsCol *gtk.Box) {
	text, err := entry.GetText()
	if err != nil {
		log.Fatal(err)
	}

	children := resultsCol.GetChildren()
	children.Foreach(func(child interface{}) {
		switch child.(type) {
		case gtk.IWidget:
			resultsCol.Remove(child.(gtk.IWidget))
		default:
			fmt.Println("Not widget")
		}
	})

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	var results monitordaemon.Results
	err = client.Call("Handler.Search", monitordaemon.Query{Str: text, N: 10}, &results)
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range results.Paths {
		l, _ := gtk.LabelNew(path)
		resultsCol.Add(l)
		resultsCol.ShowAll()
	}
}
