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
type Profile struct {
    Connection string `json:"connection" toml:"connection"`
}

type Config struct {
    Version  int                `json:"version" toml:"version"`
    Profiles map[string]Profile `json:"profile" toml:"profile"`
}
```

The field `Version` used to keep config compatible, and will return an error if config struct changed and cannot ensure
compatibility.

The field `Profiles` contains key to `Profile` struct as a map.

We will save config as toml format into local file, whose content will be like:

```toml
version = 1

[profile.one-profile]
connection = "type://content/of/connection?k=v&k2=v2"

[profile.another-profile]
connection = "type://another/content/of/connection?k=v&k2=v2"
```

### Commands

The commands will be like:

```
beyondctl profile add <name> <connection string>
beyondctl profile remove <name>
beyondctl profile list
```

For example:

```
beyondctl profile add test-s3 s3://bucket_name/qqq?crendential=hmac:access_key:secret_key&endpoint=xxxx:xxx:xx
```

This command will add a connection string with key `test-s3`, so every time we need input arg to initialize a storager
with given connection string, we can use `test-s3` instead, for example:

```
beyondctl ls test-s3:dir_to_list
```

This command will use the connection string, whose key is `test-s3`, to initialize a s3 storager, and list
dir `dir_to_list`.

**Notice**: Because the config file is introduced, these commands cannot be executed synchronously with the same config
file.

## Rationale

### Why toml instead of json?

`toml` is human-readable, easy to maintain, supports comments.

json is suitable for machines but not humans, we can add `--json` to support json output.

### What if add profile to an existing key?

For now, we just return a `profile_exists` error.

Maybe we can add a flag `--force` to satisfy both actions if needed:

- If `--force` flag was set, the new one would overwrite the existed one. 
- If not, a `profile_exists` error will be returned. 

### Support alias for profile commands?

Because the name of `profile` command is a little long for user input, maybe we can add an alias
for it, such as `p` in the future. The same situation for `remove`, `rm` is a common alias for it.

For now, we have no plan to support alias in our commands. Because it will lead misunderstanding for users 

### Optimization for fs?

We can add special treatment for `fs`. For example:

```
beyondctl cp local_dir test_s3:qqqq
```

If src/dst doesn't include a profile name like `local_dir` here, we can treat it as `fs://<pwd>/local_dir`. So the real command will become:

```
beyondctl cp fs:/path/to/local_dir test_s3:qqq
```

This can be implemented in another proposal.

### Interactive support for profile command?

In the further, we will introduce interactive profile creation:

```shell
> beyondctl profile add
It looks like you didn't input connection string, use interactive creation instead.
Please input the name.
> xxxx
Please choose the service type.
> s3
Please input the credential.
S3 support following credential protocol:
- hmac: "hmac:access_key:secret_key"
- env
> hmac:access_key:secret_key
Please input endpoint. (Input enter to use default value)
>
Profile xxxxx has been created, you can use it in the following ways:

    beyondctl ls xxxxx
    beyondctl ls xxxxx:/abc
    byonndctl cp file xxxxx:/abc/xxxxx
```

### Other Comparable Implementation

#### git remote

```
NAME
       git-remote - Manage set of tracked repositories

SYNOPSIS
       git remote [-v | --verbose]
       git remote add [-t <branch>] [-m <master>] [-f] [--[no-]tags] [--mirror=(fetch|push)] <name> <url>
       git remote rename <old> <new>
       git remote remove <name>
       git remote set-head <name> (-a | --auto | -d | --delete | <branch>)
       git remote set-branches [--add] <name> <branch>...
       git remote get-url [--push] [--all] <name>
       git remote set-url [--push] <name> <newurl> [<oldurl>]
       git remote set-url --add [--push] <name> <newurl>
       git remote set-url --delete [--push] <name> <url>
       git remote [-v | --verbose] show [-n] <name>...
       git remote prune [-n | --dry-run] <name>...
       git remote [-v | --verbose] update [-p | --prune] [(<group> | <remote>)...]
```

#### minio mcli

```
:) mcli alias --help
NAME:
  mcli alias - set, remove and list aliases in configuration file

USAGE:
  mcli alias COMMAND [COMMAND FLAGS | -h] [ARGUMENTS...]

COMMANDS:
  set, s      set a new alias to configuration file
  list, ls    list aliases in configuration file
  remove, rm  remove an alias from configuration file

FLAGS:
  --config-dir value, -C value  path to configuration folder (default: "/home/xuanwo/.mcli")
  --quiet, -q                   disable progress bar display
  --no-color                    disable color theme
  --json                        enable JSON lines formatted output
  --debug                       enable debug output
  --insecure                    disable SSL certificate verification
  --help, -h                    show help
```

## Compatibility

None

## Implementation

1. Add config model and implement the methods
2. Add `--config` as global flag to specify config file path
3. Support read config from file and write config into file
4. Add commands to reach profile in config file

[GSP-90]: https://github.com/beyondstorage/specs/pull/90