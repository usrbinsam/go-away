package config

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
)

type FileConfig string

func (fc *FileConfig) String() string {
	return os.ExpandEnv(string(*fc))
}

func (fc *FileConfig) Load() *AppConfig {
	fn := fc.String()
	fh, err := os.OpenFile(fn, os.O_RDONLY, 0600)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &AppConfig{}
		}

		log.Fatalf("error opening app storage file: %s\n", err)
	}

	b, err := io.ReadAll(fh)
	if err != nil {
		log.Fatalf("error reading app storage file contents: %s\n", err)
	}

	var as AppConfig
	err = json.Unmarshal(b, &as)
	if err != nil {
		log.Fatalf("error parsing app storage: %s\n", err)
	}
	return &as
}

func (fc *FileConfig) Store(v *AppConfig) {
	b, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		log.Fatalf("error marshalling application storage: %s\n", err)
	}
	fn := fc.String()
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("error opening app storage file: %s\n", err)
	}

	_, err = fh.Write(b)
	if err != nil {
		log.Fatalf("error writing file: %s\n", err)
	}
}

func (fc *FileConfig) Exists() bool {
	_, err := os.Stat(fc.String())
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	} else {
		log.Fatalf("unexpected error checking for config file: %s\n", err)
	}

	panic("unreachable")
}
