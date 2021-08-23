package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
)

// Version of config
const Version = 1

var (
	// ErrVerNotCompatible returned when version check failed
	ErrVerNotCompatible = errors.New("the version of config file is not compatible")
)

type Config struct {
	sync.Mutex
	Version  int                `json:"version" toml:"version"`
	Profiles map[string]Profile `json:"profile" toml:"profile"`
}

func NewConfig() *Config {
	return &Config{
		Profiles: make(map[string]Profile),
	}
}

func DefaultConfig() *Config {
	conf := NewConfig()
	conf.Version = Version
	return conf
}

func (c *Config) LoadConfigFromFile(path string) error {
	c.Lock()
	defer c.Unlock()

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err = toml.Unmarshal(data, c); err != nil {
		return err
	}

	return c.Check()
}

func (c *Config) Check() error {
	if c.Version != Version {
		return fmt.Errorf("config ver. %d is expected, %w", Version, ErrVerNotCompatible)
	}
	return nil
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
