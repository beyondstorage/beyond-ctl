- Author: JinnyYi <github.com/JinnyYi>
- Start Date: 2021-11-04
- RFC PR: [beyondstorage/beyond-ctl#96](https://github.com/beyondstorage/beyond-ctl/issues/96)
- Tracking Issue: [beyondstorage/beyond-ctl#98](https://github.com/beyondstorage/beyond-ctl/issues/98)

# BCP-96: Add Glob Patterns Support

## Background

Globs, also known as glob patterns, are patterns that can expand a wildcard pattern into a list of path that match the given pattern.

In the command shell, a wildcard is a short textual pattern, that can match another character (or characters) in a file path. It’s kind of a shortcut that allows you to specify a whole set of related path names using a single, concise pattern.

A string can be considered a wildcard pattern if it contains one of the following characters (unquoted and unescaped): `*`, `?`, `[` or `{`:

```txt
*       - (asterisk, star) matches zero or more of any characters
?       - (question mark) matches any one (1) character
[ ]     - (square brackets)
- Matches any one (1) character in the list of characters, e.g. [abc] matches one a or one b or one c (only one of the three).
- The list can be inverted/complemented by using ! at the start, e.g. [!abc] means "any one character not a or b or c".
{ }     - (curly brackets) matches on any of a series of sub-patterns you specify, e.g. {a,b,c} matches one a, one b and one c.
```

## Proposal

I propose to add glob support by using UNIX style wildcards in the path arguments of the command.

Each wildcard will be evaluated against the source path. The following pattern symbols are supported:

- ?: Matches any single non-separator character.
- *: Matches any sequence of non-separator characters.
- **: Two asterisks, matches any sequence of characters. It works like * but crosses directory boundaries (ie complete paths) in a file system.
- [...]: Match character set. This kind of wildcard specifies an "or" relationship.
- [!...] or [^...]: Match inverse character set. This is a logical NOT.
- {...}: Brace expansion, terms are separated by commas (without spaces) and each term must be the name of something or a wildcard.
- \: Backslash, used as an "escape" character.

**Notice:**
Instead of expanding the braces before even looking for files, byctl attempts to determine whether the listed file name matches the file name pattern.

Glob patterns can be used in the following commands:

- cat
- cp
- ls
- mv
- rm
- stat
- sync

### Examples

Given the following directory:

```txt
/tmp/foo/
  .git/
  |---config
  |---description
  dag.txt
  deg.txt
  big.txt
  bog.txt
  bug.txt
```

```shell
> ./byctl ls /tmp/foo/da?.txt
/tmp/foo/dag.txt

> ./byctl ls /tmp/foo/*.txt
/tmp/foo/dag.txt /tmp/foo/deg.txt /tmp/foo/dig.txt /tmp/foo/dog.txt /tmp/foo/dug.txt

> ./byctl ls /tmp/foo/d[aeiou]g.txt
/tmp/foo/dag.txt /tmp/foo/deg.txt /tmp/foo/dig.txt /tmp/foo/dog.txt /tmp/foo/dug.txt

> ./byctl ls /tmp/foo/d[a-e]g.txt
/tmp/foo/dag.txt /tmp/foo/deg.txt

> ./byctl ls /tmp/foo/d{a,e,i,o,u}g.txt
/tmp/foo/dag.txt /tmp/foo/deg.txt /tmp/foo/dig.txt /tmp/foo/dog.txt /tmp/foo/dug.txt

> ./byctl ls /tmp/foo/d{a..e}g.txt
/tmp/foo/dag.txt /tmp/foo/deg.txt
```

## Rationale

### Alternative Way: Use exclude and include filters

Most commands have `--exclude <value>` and `--include <value>` parameters to perform pattern matching to either exclude or include a particular file or object.

So we can add `exclude` and `include` options for commands. Take `cp` as an example:

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
  - [!sequence] or [^sequence]: Matches any character not in sequence

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

Drawbacks:

Does not support patterns containing directory info. e.g., --include "/usr/**/test/*.jpg"

## Compatibility

N/A

## Implementation

- Implement glob patterns
