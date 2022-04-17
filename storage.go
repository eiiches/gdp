package main

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"strings"
)

type Storage interface {
	Store(name string, data []byte) error
	Load(name string) ([]byte, error)
	List() ([]string, error)
	Delete(name string) error
}

type localStorage struct {
	configDir string
}

func NewLocalStorage() (*localStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrapf(err, "os.UserHomeDir() failed")
	}

	configDir := fmt.Sprintf("%s/.config/gdp", homeDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, errors.Wrapf(err, "os.MkdirAll(configDir, 0755) failed")
	}

	return &localStorage{
		configDir: configDir,
	}, nil
}

func (this *localStorage) Store(name string, data []byte) error {
	if err := os.WriteFile(fmt.Sprintf("%s/%s.json", this.configDir, name), data, 0644); err != nil {
		return errors.Wrapf(err, "os.WriteFile(path, data, 0644) failed")
	}
	return nil
}

func (this *localStorage) Load(name string) ([]byte, error) {
	data, err := os.ReadFile(fmt.Sprintf("%s/%s.json", this.configDir, name))
	if err != nil {
		return nil, errors.Wrapf(err, "os.ReadFile(path) failed")
	}
	return data, nil
}

func (this *localStorage) List() ([]string, error) {
	files, err := ioutil.ReadDir(this.configDir)
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil.ReadDir(%+v) failed", this.configDir)
	}

	profiles := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		profiles = append(profiles, file.Name()[:len(file.Name())-len(".json")])
	}

	return profiles, nil
}

func (this *localStorage) Delete(name string) error {
	err := os.Remove(fmt.Sprintf("%s/%s.json", this.configDir, name))
	if err != nil {
		return errors.Wrapf(err, "os.Remove(path) failed")
	}
	return nil
}
