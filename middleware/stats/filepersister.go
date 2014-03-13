package stats

import (
	"encoding/json"
	"os"
	"time"
)

type FilePersister struct {
	Path   string
	Append bool
}

func (p *FilePersister) LogEmpty() bool {
	return p.Append == false
}

func (p *FilePersister) Persist(data map[string]Snapshot) error {
	flags := os.O_CREATE | os.O_WRONLY
	if p.Append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(p.Path, flags, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	m := map[string]interface{}{
		"time":   time.Now(),
		"routes": data,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := file.Write(bytes); err != nil {
		return err
	}
	file.Write([]byte("\n"))
	return nil
}
