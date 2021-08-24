package config

import "errors"

type Profile struct {
	Connection string `json:"connection" toml:"connection"`
}

func (c *Config) AddProfile(name string, prof Profile) error {
	c.Lock()
	defer c.Unlock()

	// check existence and return error if exists
	_, ok := c.Profiles[name]
	if ok {
		return errors.New("profile already exists")
	}
	c.Profiles[name] = prof
	return nil
}

func (c *Config) RemoveProfile(name string) {
	c.Lock()
	defer c.Unlock()

	delete(c.Profiles, name)
}
