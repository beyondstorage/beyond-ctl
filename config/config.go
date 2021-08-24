package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

// Version of config
const Version = 1

type Config struct {
	sync.Mutex
	Version  int                `json:"version" toml:"version"`
	Profiles map[string]Profile `json:"profile" toml:"profile"`
}

func New() *Config {
	return &Config{
		Version:  Version, // set current version to default value, will change by parsing from
		Profiles: make(map[string]Profile),
	}
}

func LoadFromFile(path string) (*Config, error) {
	cfg := New()

	fullPath, err := expandHomeDir(path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(fullPath)
	if err != nil && os.IsNotExist(err) {
		// if config file not exist, do not load, write default config instead.
		xerr := cfg.WriteToFile(fullPath)
		if xerr != nil {
			return nil, fmt.Errorf("config file at %s not found, write default config failed: %w", fullPath, err)
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	if err = toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.Version != Version {
		return nil, fmt.Errorf("config ver. %d is expected, version of config file is not compatible", Version)
	}
	return cfg, nil
}

func (c *Config) WriteToFile(path string) error {
	c.Lock()
	defer c.Unlock()

	fullPath, err := expandHomeDir(path)
	if err != nil {
		return err
	}

	// make parent dirs
	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(c)
}

func expandHomeDir(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// replace `~/` with home dir
	return strings.Replace(path, "~/", homeDir+"/", 1), nil
}
