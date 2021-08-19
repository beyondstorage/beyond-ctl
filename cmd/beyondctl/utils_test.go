package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleParseKeyFromConnectionString() {
	var conn, key string

	conn, key = ParseKeyFromConnectionString("fs:///path/to/file")
	fmt.Printf("conn: %s, key: %s;\n", conn, key)

	conn, key = ParseKeyFromConnectionString("fs:///path/to/dir/")
	fmt.Printf("conn: %s, key: %s;\n", conn, key)

	conn, key = ParseKeyFromConnectionString("fs:///path/to/file?k1=v1")
	fmt.Printf("conn: %s, key: %s;\n", conn, key)

	conn, key = ParseKeyFromConnectionString("fs:///path/to/dir/?k1=v1")
	fmt.Printf("conn: %s, key: %s;\n", conn, key)

	// Output:
	// conn: fs:///path/to/, key: file;
	// conn: fs:///path/to/dir/, key: ;
	// conn: fs:///path/to/?k1=v1, key: file;
	// conn: fs:///path/to/dir/?k1=v1, key: ;
}

func TestParseKeyFromConnectionString(t *testing.T) {
	cases := []struct {
		name  string
		input string
		conn  string
		key   string
	}{
		{
			name:  "file",
			input: "fs:///path/to/file",
			conn:  "fs:///path/to/",
			key:   "file",
		},
		{
			name:  "dir",
			input: "s3://bucket-name/path/to/dir/",
			conn:  "s3://bucket-name/path/to/dir/",
			key:   "",
		},
		{
			name:  "file with param",
			input: "s3://bucket-name/path/to/dir/file.txt?pair1=value1&pair2=value2",
			conn:  "s3://bucket-name/path/to/dir/?pair1=value1&pair2=value2",
			key:   "file.txt",
		},
		{
			name:  "dir with param",
			input: "s3://bucket-name/path/to/dir/?pair1=value1&pair2=value2",
			conn:  "s3://bucket-name/path/to/dir/?pair1=value1&pair2=value2",
			key:   "",
		},
		{
			name:  "invalid path",
			input: "fs://",
			conn:  "fs://",
			key:   "",
		},
		{
			name:  "invalid path with key",
			input: "fs://file.txt",
			conn:  "fs://",
			key:   "file.txt",
		},
	}

	for _, tt := range cases {
		conn, key := ParseKeyFromConnectionString(tt.input)
		assert.Equal(t, tt.conn, conn, tt.name)
		assert.Equal(t, tt.key, key, tt.name)
	}
}
