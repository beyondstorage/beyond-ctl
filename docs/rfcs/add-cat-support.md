- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-11-2
- RFC PR: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)
- Tracking Issue: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)

# BCP-0: Add Cat Support

## Background

In `Linux`, the `cat` (English full spelling: `concatenate`) command is used  to concatenate files and print to a standard output device.

We will refer to the `Linux` `cat` command to implement support for the `BeyondCTL` `cat` command.

## Proposal

I propose to add support for `cat` command to `BeyondCTL`. The command format is as follows:

```
byctl cat [command options] [source]
```

### Example

Standard output of the file named `testCat` in the service `example`:

```
byctl cat example:/testCat
```

The contents of `testCat` are:

```
Hello World
Hello BeyondCTL
Hello Open Source
```

After entering the above command, it will display:

```
Hello World
Hello BeyondCTL
Hello Open Source
```

At this point we can check if the content is the same as the file `testCat`, if it is, the execution will succeed, otherwise it will fail.

`cat` can be used together with `tee`. For example:

```
byctl cat example:testCat | byctl tee another:testCat
```

Executing the above command will upload the contents of the file `testCat` in service `example` to the file `testCat` in service `another`.

#### Other

```
byctl cat example:testCat > localfile
byctl cat example:testCat >> localfile
```

The first command will input the contents of the file `testCat` in the service `example` into the local file `localfile`. If the file `localfile` does not exist, a new one will be created, and if it exists, it will be overwritten.

The second command appends the contents of the file `testCat` in the service `example` to the local file `localfile`.

## Rationale

### How the files are processed

Whether it's a regular file or an image file, we output whatever we read. We don't do any special processing.

## Compatibility

N/A

## Implementation

- Implement cat support
