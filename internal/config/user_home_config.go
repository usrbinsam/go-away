package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type UserHomeConfig struct {
	BasePath string
}

func (uhc *UserHomeConfig) getFilename() string {
	return os.ExpandEnv(fmt.Sprintf("%s/.go-away.json", uhc.BasePath))
}

func (uhc *UserHomeConfig) Load() *AppConfig {
	fh, err := os.OpenFile(uhc.getFilename(), os.O_RDONLY, 0600)
	if err != nil {
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

func (uhc *UserHomeConfig) Store(v *AppConfig) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("error marshalling application storage: %s\n", err)
	}

	fh, err := os.OpenFile(uhc.getFilename(), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("error opening app storage file: %s\n", err)
	}

	_, err = fh.Write(b)
	if err != nil {
		log.Fatalf("error writing file: %s\n", err)
	}
}

func (uhc *UserHomeConfig) Exists() bool {
	_, err := os.Stat(uhc.getFilename())
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

func NewUserHomeConfig(basePath string) *UserHomeConfig {
	return &UserHomeConfig{
		BasePath: basePath,
	}
}
