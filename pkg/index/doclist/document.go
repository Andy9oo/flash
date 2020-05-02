package doclist

// Document datastructure
type Document struct {
	id     uint32
	path   string
	length uint32
}

// ID returns the documents id
func (d *Document) ID() uint32 {
	return d.id
}

// Path returns the documents path
func (d *Document) Path() string {
	return d.path
}

// Length returns the documents length
func (d *Document) Length() uint32 {
	return d.length
}
