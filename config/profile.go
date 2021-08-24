package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	profileSeparator = ":"

	profileEnvPrefix = "BEYOND_CTL_PROFILE_"
)

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

func (c *Config) MergeProfileFromEnv() {
	c.Lock()
	defer c.Unlock()

	env := getEnvWith(profileEnvPrefix)
	for k, v := range env {
		c.Profiles[k] = Profile{Connection: v}
	}
}

// getEnvWith get env var with given prefix and return a map with key trimmed prefix
func getEnvWith(prefix string) map[string]string {
	res := make(map[string]string)

	for _, env := range os.Environ() {
		// os.Environ format as `key=value`
		// environ with same name would be overwrite, so we can use map as result
		parts := strings.Split(env, "=")
		key, value := parts[0], parts[1]

		if prefix != "" && !strings.HasPrefix(key, prefix) {
			continue
		}

		res[strings.TrimPrefix(key, prefix)] = value
	}

	return res
}
