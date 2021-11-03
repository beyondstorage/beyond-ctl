- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-10-28
- RFC PR: [beyondstorage/beyond-ctl#82](https://github.com/beyondstorage/beyond-ctl/pull/82)
- Tracking Issue: [beyondstorage/beyond-ctl#84](https://github.com/beyondstorage/beyond-ctl/issues/84)

# BCP-82: Add Tee Support

## Background

The `Linux` `tee` command is used to read data from standard input and output its contents to a file. The `tee` command reads data from a standard input device, outputs its contents to a standard output device, and also saves it to a file.

We will refer to the `Linux` `tee` command to implement support for the `BeyondCTL` `tee` command.

## Proposal

I propose to add support for `tee` command to `BeyondCTL`. `BeyondCTL` will tee the content to stdout like `Linux` `tee` command does.

```
byctl tee [command options] [target]
```

### Command options

```
--expect-size  expected size of the input file (default: "128MiB")
```
- `--expect-size` 

### Example

#### Use with the cat command

The `tee` command can also be used with the `cat` command, for example:

```
cat exampleTee | byctl tee example:testTee
```

This command will write the contents of the local file `exampleTee` to the file with the path `testTee` in the specified service `example` via the Linux pipeline `|` connection `tee` command.

Once the upload is complete, the user will be prompted to save the file successfully.

```
Stdin is saved to </testTee>
```

We can also use the flag `--expect-size` to determine the size of the input file. For example, the size of the file `exampleTee` is 2MiB:

```
cat exampleTee | byctl tee --expect-size=2MiB example:testTee
```

#### Type in the command line

Use the command `tee` to save the data entered by the user to a file with the path `testTee` in the service `example`, enter the following command:

```
byctl tee example:/testTee
```

After the above command is executed, the user will be prompted to enter the data to be saved to a file, as follows:

```
test tee command    #Prompt the user to enter data 
test tee command    #Output data for feedback output
```

After the user enters `Ctrl D`, the user will be prompted that the data entered by the user has been saved to the path entered by the user.

```
Stdin is saved to </testTee>
```

In this case, you can open the file `testTee` and check if the content is the input to see if the command `tee` is executed successfully.

## Rationale

### What is a pipe character in Linux

Linux pipes use the vertical line `|` to connect multiple commands, which is called a pipe character. The specific syntax format of a Linux pipe is as follows:

```
command1 | command2
```

When a pipe is set up between two commands, the output of the left command of the pipe character `|` becomes the input of the right command. As long as the first command writes to the standard output and the second command reads from the standard input, then the two commands can form a pipe.

### What do we use to upload the content to the specified service?

For commands that use pipeline character(`|`), we use `multipart` uploads where the user can enter `--expect-size=xxx` to upload the approximate size of the file (128MiB by default).

For terminal input, we use `Write` to upload.

## Compatibility

N/A

## Implementation

- Implement tee support