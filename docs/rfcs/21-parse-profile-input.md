- Author: Prnyself <281239768@qq.com>
- Start Date: 2021-08-24
- RFC PR: [beyondstorage/beyond-ctl#21](https://github.com/beyondstorage/beyond-ctl/pull/21)
- Tracking Issue: [beyondstorage/beyond-ctl#22](https://github.com/beyondstorage/beyond-ctl/issues/22)

# BCP-21: Parse Profile Input

## Background

From [BCP-7], we have introduced profile support, and we can simplify our input with profile name, like:

```
beyondctl ls test-s3:dir_to_list
```

However, for local fs, all it need is to initialize a storager is the `work_dir`, so it introduces another inconvenience
when using local fs.

Moreover, shall we also support input connection string instead of adding profile before? Shall we support adding
profile by environment variables?

## Proposal

So I propose to add following rules when parse input:

- using `:` as the separator between `profile` and `key`, such as `test-s3:dir_to_list`.
- try to find `profile` with given `name`
    - if found, use `profile`'s connection string, error would return if connection string
      invalid
    - if not, a `profile not found` error would return
- if no separator found, treat input as local fs path
    - if the path is an absolute path, set `/` as `work_dir`, and the input as `key`
    - if the path is a relative path, set `pwd` as `work_dir`, and the input as `key`

### env support

For now, we initialize `profile` list from config file, and modify config file when execute `add` or `remove` `profile`
command. To support set profile by environment, we should specify the prefix of environment variables which will be used
as profile key and value.

For `BEYOND_CTL_PROFILE_xxxx = s3://xxxx`, a profile with name `xxxx` and connection string
`s3://xxxx` will be merged into config profile, and can be used in input, without adding into config file.

Moreover, to avoid writing environment variable profile into config file, we do not support using environment variables
in `profile` commands.

## Rationale

### Why not support raw connection string from input?

- First, the connection is usually too long to input.
- Second, there would be some sensitive data such as sk or password in connection string, which is not appropriate to
  show in input string.
- Third, we add support for get profile from environment variable, so it is unnecessary for user to input raw
  connection.

### Why not support env vars in profile commands?

It will introduce complex problems and situation into profiles. For example

```
BEYOND_CTL_PROFILE_key1 = s3://xxxx beyondctl add key1 fs:///xxx 
```

This will lead into a conflict between input and environment variable, which is meaningless.

## Compatibility

None.

## Implementation

1. Add input parser by rules listed in proposal
2. Support reading environment variables with given prefix, and merge them into current `profile`

[BCP-7]: https://github.com/beyondstorage/beyond-ctl/blob/master/docs/rfcs/7-add-profile-support.md
