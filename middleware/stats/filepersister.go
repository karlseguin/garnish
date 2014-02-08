package stats

import (
	"encoding/json"
	"os"
	"time"
)

type FilePersister struct {
	path string
}

func (p *FilePersister) Persist(data map[string]Snapshot) error {
	file, err := os.OpenFile(p.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
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
