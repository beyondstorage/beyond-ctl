package main

import "strings"

// ParseKeyFromConnectionString parse key and keep connection string as original without object key
// since we do not check connection string valid or not here,
// so we will only split key from original connection string
// and return the key-split connection string, key, not return an error
func ParseKeyFromConnectionString(input string) (conn, key string) {
	var b strings.Builder

	question := strings.Index(input, "?")
	var connStr, paramStr string

	// no question mark found, like: fs:///path/to/dir/key, means no param found
	if question == -1 {
		connStr = input
	} else {
		connStr = input[:question]
		paramStr = input[question:] // paramStr keeps ? as its start so that we can build result directly
	}

	// if connStr end with /, we handle it as a dir, split key from connStr
	// otherwise, we handle it as /dir/to/key
	if !strings.HasSuffix(connStr, "/") {
		split := strings.LastIndex(connStr, "/")
		key = connStr[split+1:]
		connStr = connStr[:split+1]
	}

	// build new connection string without key
	b.WriteString(connStr)
	b.WriteString(paramStr)
	conn = b.String()
	return
}
