package cache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/karlseguin/garnish/gc"
	"log"
	"os"
)

var (
	endianness = binary.LittleEndian
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

	buffer := new(bytes.Buffer)
	binary.Write(file, endianness, len(entries))
	for _, entry := range entries {
		buffer.Reset()
		switch entry.CachedResponse.(type) {
		case *gc.NormalResponse:
			buffer.WriteByte(1)
		case *gc.HydrateResponse:
			buffer.WriteByte(2)
		default:
			err = errors.New("unknown response type")
			return
		}
		if err = entry.Serialize(buffer); err != nil {
			return
		}
		binary.Write(file, endianness, buffer.Len())
		if _, err = buffer.WriteTo(file); err != nil {
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
