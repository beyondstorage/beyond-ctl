- Author: JinnyYi <github.com/JinnyYi>
- Start Date: 2021-11-04
- RFC PR: [beyondstorage/beyond-ctl#96](https://github.com/beyondstorage/beyond-ctl/issues/96)
- Tracking Issue: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)

# BCP-96: Add Glob Pattern Support in cp

## Background

For bulk file copying, we only support directory copying now.

Currently, there is no support for the use of UNIX style wildcards in a command's path arguments. However, most commands have `--exclude <value>` and `--include <value>` parameters that can achieve the desired result. These parameters perform pattern matching to either exclude or include a particular file or object.

## Proposal

I propose to add glob support in `cp`. For this case, I suggest add string slice flags `exclude` and `include` to `cp`.

### Command

Usage:

```
byctl cp [command options] [source] [target]
```

Options:

```
--exclude value    The UNIX-style wildcard to ignore, except by include statements
--include value    The UNIX-style wildcard to act upon
```

Description:

- The following pattern symbols are supported:
  - *: Matches everything
  - ?: Matches any single character
  - [sequence]: Matches any character in sequence
  - [!sequence]: Matches any character not in sequence

- Each filter is evaluated against the source directory. If the source location is a file instead of a directory, the directory containing the file is used as the source directory.
  ```
  cp /tmp/foo /tmp/dir 
  The source directory is /tmp/foo, any include/exclude filters will be evaluated with the source directory prepended.
  ```

- Any number of these options can be passed to a command. When there are multiple filters, the rule is the filters that appear later in the command take precedence over filters that appear earlier in the command.
  ```txt
  --exclude "*" --include "*.txt"
  All files will be excluded from the command except for files ending with ".txt"
  ```
  
- All files are included by default. This means that providing only an `--include` filter will not change what files are transferred. `--include` will only re-include files that have been excluded from an `--exclude` filter.

### Examples

Given the following directory:

```txt
/tmp/foo/
  .git/
  |---config
  |---description
  foo.txt
  bar.txt
  baz.jpg
```

The command `byctl cp -R --exclude "ba*" /tmp/foo/ s3:dir1` will exclude `/tmp/foo/bar.txt` and `/tmp/foo/baz.jpg`:

```txt
dir1/
  .git/
  |---config
  |---description
  foo.txt
```

The command `byctl cp -R --exclude "*" --include "*.jpg" --include "*.txt" /tmp/foo/ s3:dir2` will only include files with name end with `.jpg` or `.txt`:

```txt
dir2/
  foo.txt
  bar.txt
  baz.jpg
```

## Rationale

### Other Comparable Implementation

#### [Amazon CLI](https://docs.aws.amazon.com/cli/latest/reference/s3/#use-of-exclude-and-include-filters)

GlobPattern in Amazon CLI: Use exclude and include filters.

Example:

```txt
aws s3 cp --recursive --exclude '*' --include 'part*' s3://my-amazing-bucket/ s3://my-other-bucket/
```

#### [Aliyun CLI](https://github.com/aliyun/aliyun-cli/blob/master/oss/lib/cp.go)

syntaxText:

```txt
ossutil cp src_url dest_url [options]

--include and --exclude option:

    When --recursive is specified, these parameters perform pattern matching to either exclude or
    include a particular file or object. By default, all files/objects are included.
```

sampleText:

```txt
e.g., 3 files in current dir:
    testfile1.jpg
    testfiel2.txt
    testfile33.jpg

$ ossutil cp . oss://my-bucket/path --exclude '*.jpg'
upload testfile2.txt to oss://my-bucket/path/testfile2.txt

$ ossutil cp . oss://my-bucket/path --exclude '*.jpg' --include 'testfile*.jpg'
upload testfile1.jpg to oss://my-bucket/path/testfile1.jpg
upload testfile33.jpg to oss://my-bucket/path/testfile33.jpg
upload testfile2.txt to oss://my-bucket/path/testfile2.txt

$ ossutil cp . oss://my-bucket/path --exclude '*.jpg' --include 'testfile*.jpg' --exclude 'testfile?.jpg'
upload testfile2.txt to oss://my-bucket/path/testfile2.txt
upload testfile33.jpg to oss://my-bucket/path/testfile33.jpg
```

## Compatibility

N/A

## Implementation

- Implement glob pattern in `cp`
