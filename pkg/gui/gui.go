package gui

import (
	"flash/pkg/monitordaemon"
	"fmt"
	"log"
	"net/rpc"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/viper"
)

const escape uint = 65307

// Show displays the gui
func Show() {
	// Init window
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle("Flash 🔍")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	win.SetDecorated(false)
	win.SetPosition(gtk.WIN_POS_CENTER)

	win.Connect("focus-out-event", func() {
		win.Destroy()
	})

	// Create UI Elements
	col, _ := gtk.BoxNew(gtk.Orientation(gtk.ORIENTATION_VERTICAL), 0)
	entry, _ := gtk.SearchEntryNew()
	results, _ := gtk.ListBoxNew()

	// Add events
	entry.Connect("search-changed", func() {
		win.Resize(600, entry.GetAllocatedHeight())
		handleSearch(entry, results)
	})

	win.Connect("key-press-event", func(_ *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{Event: ev}
		fmt.Println(keyEvent)
		if keyEvent.KeyVal() == escape {
			win.Destroy()
		}
	})

	results.SetActivateOnSingleClick(false)
	results.Connect("row-activated", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		file, err := row.GetName()
		if err != nil {
			fmt.Println(err)
			return
		}
		open.Run(file)
		win.Destroy()
	})

	col.Add(entry)
	col.Add(results)
	win.Add(col)

	win.SetDefaultSize(600, -1)
	win.ShowAll()
	gtk.Main()
}

func handleSearch(entry *gtk.SearchEntry, resultsCol *gtk.ListBox) {
	text, err := entry.GetText()
	if err != nil {
		log.Fatal(err)
	}

	children := resultsCol.GetChildren()
	children.Foreach(func(child interface{}) {
		switch child.(type) {
		case gtk.IWidget:
			widget := child.(gtk.IWidget)
			resultsCol.Remove(widget)
		default:
			fmt.Println("Not widget")
		}
	})

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	var results monitordaemon.Results
	err = client.Call("Handler.Search", monitordaemon.Query{Str: text, N: viper.GetInt("gui_results")}, &results)
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range results.Paths {
		row := newResult(path)
		resultsCol.Add(row)
		resultsCol.ShowAll()
	}
}
