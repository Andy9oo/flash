package partition

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

const partitionLimit = 1 << 10

// Collector is used to abstract partitioning, automatically writing and merging partitions where needed
type Collector struct {
	dir               string
	extension         string
	memory            *partition
	disk              []*partition
	newImplementation func() Implementation
}

// NewCollector creates a new collector
func NewCollector(dir, extension string, newImplementation func() Implementation) *Collector {
	c := Collector{
		dir:               dir,
		extension:         extension,
		newImplementation: newImplementation,
	}
	c.addPartition()
	return &c
}

// Load will load the partitions into memory
func (c *Collector) Load() error {
	return c.loadInfo()
}

// Add insets a new key value pair into the index
func (c *Collector) Add(key string, val Entry) {
	if c.memory.full() {
		c.mergeParitions()
	}
	c.memory.add(key, val)
}

// GetBuffers returns all of the buffers which use the given key
func (c *Collector) GetBuffers(key string) []*bytes.Buffer {
	var buffers []*bytes.Buffer
	for _, p := range append(c.disk, c.memory) {
		if buf, ok := p.getBuffer(key); ok {
			buffers = append(buffers, buf)
		}
	}
	return buffers
}

// GetEntries returns all entries which match the given key
func (c *Collector) GetEntries(key string) []Entry {
	var entries []Entry
	for _, p := range append(c.disk, c.memory) {
		if e, ok := p.getEntry(key); ok {
			entries = append(entries, e)
		}
	}
	return entries
}

func (c *Collector) addPartition() {
	if c.memory != nil {
		c.disk = append(c.disk, c.memory)
	}
	c.memory = newPartition(c.dir, c.extension, 0, partitionLimit, c.newImplementation())
}

func (c *Collector) mergeParitions() {
	// Dump current partition into temp file
	mem := c.memory
	mem.dump()

	// Sort partitions in order of generation
	sort.Slice(c.disk, func(p1, p2 int) bool {
		return c.disk[p1].generation < c.disk[p2].generation
	})

	// Anticipate collisions
	g := 1
	var parts []*partition
	for p := 0; p < len(c.disk); p++ {
		if c.disk[p].generation == g {
			parts = append(parts, c.disk[p])
			g++
		} else {
			break
		}
	}

	if len(parts) == 0 {
		oldPath := mem.getPath()
		// Set current partition as final
		mem.generation = 1
		os.Rename(oldPath, mem.getPath())
		mem.loadDict()
	} else {
		// Remove old partitions from the index
		c.disk = c.disk[len(parts):]
		c.memory = nil
		// Merge partitions
		p := newPartition(c.dir, c.extension, g, partitionLimit, c.newImplementation())
		merge(append(parts, mem), p)
		c.disk = append(c.disk, p)
		p.loadDict()
	}

	c.addPartition()
}

func (c *Collector) dumpInfo() {
	path := fmt.Sprintf("%v/%v.info", c.dir, c.extension)
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := new(bytes.Buffer)
	for p := range c.disk {
		binary.Write(buf, binary.LittleEndian, uint32(c.disk[p].generation))
	}
	binary.Write(buf, binary.LittleEndian, uint32(c.memory.generation))
	buf.WriteTo(f)
}

func (c *Collector) loadInfo() error {
	path := fmt.Sprintf("%v/%v.info", c.dir, c.extension)
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	buf := make([]byte, 4)

	for {
		n, err := reader.Read(buf)
		if n == 0 || err != nil {
			break
		}

		gen := int(binary.LittleEndian.Uint32(buf))
		part := loadPartition(c.dir, c.extension, gen, partitionLimit, c.newImplementation())

		// Load in-memory index
		if gen == 0 {
			c.memory = part
		} else {
			c.disk = append(c.disk, part)
		}
	}
	return nil
}

// ClearMemory writes any remaining info to disk
func (c *Collector) ClearMemory() {
	c.dumpInfo()
	c.memory.dump()
}
