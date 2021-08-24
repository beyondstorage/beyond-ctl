package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ParseProfileInput(t *testing.T) {
	cfg := New()
	_ = cfg.AddProfile("test1", Profile{
		Connection: "s3://bucket-name/dir/",
	})

	cases := []struct {
		name   string
		input  string
		conn   string
		key    string
		hasErr bool
	}{
		{
			name:   "normal profile",
			input:  "test1:object_key",
			conn:   "s3://bucket-name/dir/",
			key:    "object_key",
			hasErr: false,
		},
		{
			name:   "invalid conn",
			input:  ":/dir/to/file",
			conn:   "",
			key:    "",
			hasErr: true,
		},
		{
			name:   "profile not exist",
			input:  "test2:object_key",
			conn:   "",
			key:    "",
			hasErr: true,
		},
		{
			name:   "blank key",
			input:  "test1:",
			conn:   "s3://bucket-name/dir/",
			key:    "",
			hasErr: false,
		},
		{
			name:   "fs abs path",
			input:  "/path/to/file",
			conn:   "fs:///",
			key:    "/path/to/file",
			hasErr: false,
		},
		// {
		// 	name:   "fs rel path",
		// 	input:  "path/to/file",
		// 	conn:   "fs:///absolute/path/to/wd",
		// 	key:    "path/to/file",
		// 	hasErr: false,
		// },
	}

	for _, tt := range cases {
		conn, key, err := cfg.ParseProfileInput(tt.input)
		if tt.hasErr {
			assert.NotNil(t, err, tt.name)
			continue
		}

		assert.Nil(t, err, tt.name)
		assert.Equal(t, tt.conn, conn, tt.name)
		assert.Equal(t, tt.key, key, tt.name)
	}
}
