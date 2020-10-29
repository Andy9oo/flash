package gui

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"github.com/rakyll/magicmime"
)

type result struct {
	*gtk.ListBoxRow
}

func newResult(path string) *result {
	row, _ := gtk.ListBoxRowNew()
	container, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)

	row.SetName(path)
	theme, _ := gtk.IconThemeGetDefault()

	icon := getIcon(path, theme)
	content := getContent(path)

	container.Add(icon)
	container.Add(content)

	row.Add(container)

	return &result{ListBoxRow: row}
}

func getContent(path string) *gtk.Box {
	container, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 1)
	name := filepath.Base(path)

	title, _ := gtk.LabelNew("")
	title.SetXAlign(0)
	markup := fmt.Sprintf("<span weight=\"500\" size=\"%d\">%s</span>", 12*pango.PANGO_SCALE, name)
	title.SetMarkup(markup)

	info, _ := gtk.LabelNew("")
	info.SetXAlign(0)
	markup = fmt.Sprintf("<span weight=\"300\" size=\"%d\" color=\"#505050\">%s</span>", 10*pango.PANGO_SCALE, path)
	info.SetMarkup(markup)

	container.Add(title)
	container.Add(info)

	return container
}

func getIcon(path string, theme *gtk.IconTheme) *gtk.Image {
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
		log.Fatal(err)
	}
	defer magicmime.Close()

	mime, err := magicmime.TypeByFile(path)
	fmt.Println(mime)
	if err != nil {
		log.Fatalf("error occured during type lookup: %v", err)
	}

	if strings.Contains(mime, "image") {
		mime = "image/x/generic"
	}

	path = strings.Replace(mime, "/", "-", -1)
	pixbuf, err := theme.LoadIcon(path, 32, gtk.ICON_LOOKUP_GENERIC_FALLBACK)
	if err != nil {
		icon, _ := gtk.ImageNewFromIconName("file", gtk.ICON_SIZE_DND)
		return icon
	}

	icon, _ := gtk.ImageNewFromPixbuf(pixbuf)
	return icon
}
