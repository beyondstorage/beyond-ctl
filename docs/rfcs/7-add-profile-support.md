- Author: Prnyself <281239768@qq.com>
- Start Date: 2021-08-19
- RFC PR: [beyondstorage/beyond-ctl#7](https://github.com/beyondstorage/beyond-ctl/pull/7)
- Tracking Issue: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)

# BCP-7: Add Profile Support

## Background

In [GSP-90], we introduced connection string to initialize a storager by string input.

So we had originally planned not to introduce config file for beyond-ctl, for which, users can input the config of a
storager in command line. For example:

```
beyondctl cp fs:///path/to/key s3://bucket_name/object_name
```

However, the connection may be too long for users to input every time, especially when it contains some params
like `?endpoint=xxxx:xxx:xx`, or credential info which is usually sensitive.

So we can write these profile info into local config file, and support commands to `get`,
`set`, `delete` profile.

## Proposal

So I propose to add profile support to record user's connection string with an alias in config file.

### Config

The config struct will be like:

```go
type Config struct {
    Version  int                   `json:"version" toml:"version"`
    Profiles map[string]Connection `json:"profiles" toml:"profiles"`
}

type Connection struct {
    Type    string              `json:"type" toml:"type"`
    Name    string              `json:"name" toml:"name"`
    WorkDir string              `json:"work_dir" toml:"work_dir"`
    Pairs   map[string]string   `json:"pairs" toml:"pairs"`
}

func (c Connection) String() string {
    // Build connection string from Connection
}

func NewConnectionFromString(input string) Connection {
    // Parse connection string into Connection
}
```

The field `Version` used to keep config compatible, and will return an error if config struct changed and cannot ensure
compatibility.

The field `Identities` contains key to connection string as a map.

We will save config as toml format into local file.  

### Commands

The commands will be like:

```
beyondctl profile set <key> <connection string>
beyondctl profile delete <key>
beyondctl profile list
```

For example:

```
beyondctl profile set test-s3 s3://bucket_name/qqq?crendential=hmac:access_key:secret_key&endpoint=xxxx:xxx:xx
```

This command will set a connection string with key `test-s3`, so every time we need input arg to initialize a storager
with given connection string, we can use `test-s3` instead, for example:

```
beyondctl ls test-s3:dir_to_list
```

This command will use the connection string, whose key is `test-s3`, to initialize a s3 storager, and list
dir `dir_to_list`.

**Notice**: Because the config file is introduced, these commands cannot be executed synchronously with the same config
file.

## Rationale

Minio mc alias
implementation: <https://github.com/minio/mc/blob/a88acb58d41eec81791e97672c5318f4582e00dc/cmd/alias-set.go#L154>

### Why toml instead of json?

`toml` is human-readable, easy to maintain, supports comments.

json is suitable for machines but not humans, we can add `--json` to support json output.

### What if add profile to an existing key?

For now, we just return a `profile_exists` error.

Maybe we can add a flag `--force` to satisfy both actions if needed:

- If `--force` flag was set, the new one would overwrite the existed one. 
- If not, a `profile_exists` error will be returned. 

## Compatibility

None

## Implementation

1. Add config model and implement the methods
2. Add `--config` as global flag to specify config file path
3. Support read config from file and write config into file
4. Add commands to reach profile in config file

[GSP-90]: https://github.com/beyondstorage/specs/pull/90