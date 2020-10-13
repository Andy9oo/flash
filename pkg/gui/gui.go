package gui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

func Display() {
	// Initialize GTK
	gtk.Init(nil)
	// Create Window
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle("Spotlight 🔍")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	col, _ := gtk.BoxNew(gtk.Orientation(gtk.ORIENTATION_VERTICAL), 2)
	// Create elements
	row, _ := gtk.BoxNew(gtk.Orientation(gtk.ORIENTATION_HORIZONTAL), 0)

	b, _ := gtk.ButtonNewWithLabel("Search")
	b.SetSizeRequest(100, 50)

	entry, _ := gtk.EntryNew()
	entry.SetSizeRequest(500, 50)

	// Add events
	b.Connect("clicked", func() {
		handleSearch(entry, col)
	})

	entry.Connect("activate", func() {
		handleSearch(entry, col)
	})

	col.Add(row)

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

func handleSearch(entry *gtk.Entry, col *gtk.Box) {
	text, err := entry.GetText()
	if err != nil {
		log.Fatal(err)
	}

	l, _ := gtk.LabelNew(text)

	col.Add(l)
	col.ShowAll()
	entry.SetText("")
}
