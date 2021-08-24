package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const profileSeparator = ":"

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

func (c *Config) ParseProfileInput(input string) (conn, key string, err error) {
	c.Lock()
	defer c.Unlock()

	sepIdx := strings.LastIndex(input, profileSeparator)

	// separator not found, treat as normal fs path
	// 1. if path is absolute path, set work dir is /
	// 2. if path is relative path, set work dir is current dir
	// in both condition, treat path as object key
	if sepIdx == -1 {
		prefix := "fs://"
		// absolute path
		if strings.HasPrefix(input, "/") {
			conn = prefix + "/"
		} else {
			curDir, err := os.Getwd()
			if err != nil {
				return "", "", err
			}
			conn = prefix + curDir
		}

		key = input
		return
	}

	// handle profile
	name := input[:sepIdx]
	prof, ok := c.Profiles[name]
	// profile not exist, treat as full connection string
	if !ok {
		return "", "", fmt.Errorf("profile with name %s not exist", name)
	}

	conn = prof.Connection

	// handle key
	// if input end with ':', set key to blank
	if sepIdx == len(input)-1 {
		key = ""
	} else {
		key = input[sepIdx+1:]
	}
	return
}
