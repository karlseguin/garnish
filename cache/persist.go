package cache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/karlseguin/garnish/gc"
	"log"
	"os"
)

type persist struct {
	count int
	path  string
	done  chan error
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
		serializer.WriteString(entry.Primary)
		serializer.WriteString(entry.Secondary)
		if _, err = file.Write(serializer.Bytes()); err != nil {
			return
		}
	}
	err = file.Sync()
}

func loadFromFile(path string) ([]*Entry, error) {
	return nil, nil
	// f, err := os.Open(path)
	// if err != nil {
	// 	return nil, err
	// }
	// defer f.Close()
	//
	// var count int
	// binary.Read(file, endianness, &count)
	// entries := make([]*Entry, count)
	// for i := 0; i < count; i++ {
	//
	// }
	// return entries, nil
}

type Serializer struct {
	scratch []byte
	*bytes.Buffer
}

func newSerializer() *Serializer {
	return &Serializer{
		scratch: make([]byte, 4),
		Buffer:  new(bytes.Buffer),
	}
}

func (s *Serializer) WriteByte(b byte) {
	s.Buffer.WriteByte(b)
}

func (s *Serializer) WriteInt(i int) {
	s.Buffer.Grow(4)
	binary.BigEndian.PutUint32(s.scratch, uint32(i))
	s.Buffer.Write(s.scratch)
}

func (s *Serializer) Write(b []byte) {
	s.WriteInt(len(b))
	s.Buffer.Write(b)
}

func (s *Serializer) WriteString(str string) {
	s.Write([]byte(str))
}
