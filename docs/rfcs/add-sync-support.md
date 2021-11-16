- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-11-16
- RFC PR: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)
- Tracking Issue: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)

# BCP-0: Add Sync Support

## Background

In daily use, we may modify a folder in one service, but we may have put this folder in another service as well, so we need to synchronize the folders in both services. This ensures that our changes are not lost and that the folders are up to date in each service.

## Proposal

So I propose to add `sync` support to BeyondCTL.

```
byctl sync [command options] [source] [target]
```

### Command options

```
--existing         skip creating new files in target dirs
--ignore-existing  skip updating files in target dirs, only copy those not exist
--recursive        recurse into sub directories
--update           skip files that are newer in target dirs
--remove           remove extraneous object(s) on target
--exclude value    regular expression for files to exclude
--include value    regular expression for files to include (not work if exclude not set)
```

### Examples

Service `example` given the following directory:

```
test/
	|--dog.go
	|--dog.txt
	|--cat.go
	|--fruit/
		    |--apple.go
```

```
byctl sync example:test/ another:test/
```

The result in the service `another`:

```
test/
	|--dog.go
	|--dog.txt
	|--cat.go
```

#### Sync with `--existing`

Service `another` has the following directory:

```
test/
	|--dog.go
	|--dog.txt
```

```
byctl sync --existing example:test/ another:test/
```

The result in the service `another`:

```
test/
	|--dog.go
	|--dog.txt
```

Only the existing files in service `another` are updated.

#### Sync with `--ignore-existing`

Service `another` has the following directory:

```
test/
	|--dog.go
	|--dog.txt
```

```
byctl sync --ignore-existing example:test/ another:test/
```

The result in the service `another`:

```
test/
	|--dog.go
	|--dog.txt
	|--cat.go
```

There are no changes to `dog.go` and `dog.txt`, only `cat.go` has been updated.

#### Sync with `--recursive`

```
byctl sync --recursive example:test/ another:test/
```

The result in the service `another`:

```
test/
	|--dog.go
	|--dog.txt
	|--cat.go
	|--fruit/
	   	    |--apple.go
```

#### Sync with `--update`

Suppose service `example` and `another` have been synchronized once and then the file `dog.go` in `example` has been modified:

```
byctl sync --update example:test/ another:test/
```

In this case, only `dog.go` is updated in the service `another` and the other files are left unchanged.

#### Sync with `--remove`

Service `another` has the following directory:

```
test/
	|--pig.go
```

```
byctl sync --remove example:test/ another:test/
```

The result in the service `another`:

```
test/
	|--dog.go
	|--dog.txt
	|--cat.go
```

In this case all files in service `another`'s folder `test` that are not related to service `example`'s folder `test` will be deleted.

#### Sync with `--exclude` and `include`

- `--exclude`

```
byctl sync --exclude="(.*).go" example:test/ another:test/
```

The result in the service `another`:

```
test/
	|--dog.txt
```

````
byctl sync --exclude="dog.(.*)" example:test/ another:test/
````

The result in the service `another`:

```
test/
	|--cat.go
```

- `--exclude` and `include`

```
byctl sync --exclude="(.*).go" --include="dog.go" example:test/ another:test/
```

The result in the service `another`:

```
test/
	|--dog.go
	|--dog.txt
```

```
byctl sync --exclude="dog.(.*)" --include="dog.go" example:test/ another:test/
```

The result in the service `another`:

```
test/
	|--dog.go
	|--cat.go
```

## Rationale

N/A

## Compatibility

N/A

## Implementation

- Implement sync support
- Add the sync document to the [site](https://github.com/beyondstorage/site)