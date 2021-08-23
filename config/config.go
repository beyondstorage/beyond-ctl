package config

import (
	"fmt"
	"io/ioutil"
	"os"
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

	data, err := ioutil.ReadFile(path)
	if err != nil {
		// if config file not exist, do not load
		if os.IsNotExist(err) {
			// write default config into path if not exist
			if err := cfg.WriteToFile(path); err != nil {
				return nil, fmt.Errorf("config file at %s not found, write default config failed: %w", path, err)
			}
			return cfg, nil
		}
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

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return enc.Encode(c)
}
