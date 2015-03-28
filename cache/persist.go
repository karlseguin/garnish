package cache

import (
	"bytes"
	"errors"
	"gopkg.in/karlseguin/garnish.v1/gc"
	"io"
	"log"
	"os"
	"time"
)

type persist struct {
	count  int
	path   string
	cutoff time.Duration
	done   chan error
}

func (p persist) persist(entries []*Entry) {
	var err error
	defer func() { p.done <- err }()
	if len(entries) == 0 {
		return
	}

	file, err := os.Create(p.path)
	if err != nil {
		return
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Println("cache file close", err)
		}
	}()

	serializer := newSerializer()
	serializer.WriteInt(len(entries))
	if _, err = file.Write(serializer.Bytes()); err != nil {
		return
	}
	for _, entry := range entries {
		serializer.Reset()
		serializer.WriteString(entry.Primary)
		serializer.WriteString(entry.Secondary)

		switch entry.CachedResponse.(type) {
		case *gc.NormalResponse:
			serializer.WriteByte(1)
		case *gc.HydrateResponse:
			serializer.WriteByte(2)
		default:
			err = errors.New("unknown response type")
			return
		}
		if err = entry.Serialize(serializer); err != nil {
			return
		}
		if _, err = file.Write(serializer.Bytes()); err != nil {
			return
		}
	}
	err = file.Sync()
}

func loadFromFile(path string) ([]*Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	deserializer := newDeserializer(f)
	count := deserializer.ReadInt()
	entries := make([]*Entry, count)
	for i := 0; i < count; i++ {
		primary, secondary := deserializer.ReadString(), deserializer.ReadString()
		var response gc.CachedResponse
		switch deserializer.ReadByte() {
		case 1:
			response = new(gc.NormalResponse)
		case 2:
			response = new(gc.HydrateResponse)
		default:
			return nil, errors.New("unknown response type")
		}
		response.Deserialize(deserializer)
		entries[i] = &Entry{
			Primary:        primary,
			Secondary:      secondary,
			CachedResponse: response,
		}
	}
	return entries, nil
}

type Serializer struct {
	*bytes.Buffer
}

func newSerializer() *Serializer {
	return &Serializer{
		Buffer: new(bytes.Buffer),
	}
}

func (s *Serializer) WriteByte(b byte) {
	s.Buffer.WriteByte(b)
}

func (s *Serializer) WriteInt(value int) {
	s.WriteByte(byte(value & 0xFF))
	s.WriteByte(byte(value >> 8))
	s.WriteByte(byte(value >> 16))
	s.WriteByte(byte(value >> 24))
}

func (s *Serializer) Write(b []byte) {
	s.WriteInt(len(b))
	s.Buffer.Write(b)
}

func (s *Serializer) WriteString(str string) {
	s.Write([]byte(str))
}

type Deserializer struct {
	scratch []byte
	reader  io.Reader
}

func newDeserializer(reader io.Reader) *Deserializer {
	return &Deserializer{
		reader:  reader,
		scratch: make([]byte, 36864),
	}
}

func (d *Deserializer) ReadInt() int {
	d.reader.Read(d.scratch[:4])
	value := int(d.scratch[0])
	value += int(d.scratch[1]) << 8
	value += int(d.scratch[2]) << 16
	value += int(d.scratch[3]) << 24
	return value
}

func (d *Deserializer) ReadString() string {
	return string(d.ReadBytes())
}

func (d *Deserializer) ReadBytes() []byte {
	return d.ReadN(d.ReadInt())
}

func (d *Deserializer) CloneBytes() []byte {
	b := d.ReadN(d.ReadInt())
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (d *Deserializer) ReadByte() byte {
	return d.ReadN(1)[0]
}

func (d *Deserializer) ReadN(n int) []byte {
	if len(d.scratch) < n {
		d.scratch = make([]byte, n)
	}
	scratch := d.scratch[:n]
	d.reader.Read(scratch)
	return scratch
}
