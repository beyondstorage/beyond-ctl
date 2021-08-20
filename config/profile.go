package config

import "errors"

var (
	// ErrProfileAlreadyExist returned when add profile to an existing name
	ErrProfileAlreadyExist = errors.New("profile already exists")
)

type Profile struct {
	Connection string `json:"connection" toml:"connection"`
}

// Encoder used to encode struct
type Encoder interface {
	Encode(v interface{}) error
}

func (c *Config) AppendProfile(name string, prof Profile, force bool) error {
	c.Lock()
	defer c.Unlock()

	// if not force append, check existence and return error if true
	if !force {
		_, ok := c.Profiles[name]
		if ok {
			return ErrProfileAlreadyExist
		}
	}
	c.Profiles[name] = prof
	return nil
}

func (c *Config) RemoveProfile(name string) {
	c.Lock()
	defer c.Unlock()

	delete(c.Profiles, name)
}
