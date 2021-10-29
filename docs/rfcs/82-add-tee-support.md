- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-10-28
- RFC PR: [beyondstorage/beyond-ctl#82](https://github.com/beyondstorage/beyond-ctl/pull/82)
- Tracking Issue: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)

# BCP-82: Add Tee Support

## Background

The `Linux` `tee` command is used to read data from standard input and output its contents to a file. The `tee` command reads data from a standard input device, outputs its contents to a standard output device, and also saves it to a file.

We will refer to the `Linux` `tee` command to implement support for the `BeyondCTL` `tee` command.

## Proposal

I propose to add support for `tee` command to `BeyondCTL`. `BeyondCTL` will tee the content to stdout like `Linux` `tee` command does.

```
byctl tee [command options] [target]
```

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

After the user enters `Ctrl C`, the user will be prompted that the data entered by the user has been saved to the path entered by the user.

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

### How to capture the `Ctrl C`?

We use `golang's` `signal` library and the `syscall.SIGINT` and `syscall.SIGTERM` parameters to capture the terminal triggered by typing `Ctrl C`.

| Signal  |                     Description                     |
| :-----: | :-------------------------------------------------: |
| SIGINT  |    User sends INTR character (Ctrl+C) to trigger    |
| SIGTERM | End the program (can be caught, blocked or ignored) |

### If we don't want to continue with the current command, how do we interrupt?

Since we use `Ctrl C` to determine if the user has finished typing, we can no longer use `Ctrl C` to interrupt the program. In this case, we can just use `Ctrl Z` to achieve our goal.

### What do we use to upload the content to the specified service?

Since `multipart` uploads basically have a minimum byte limit, and the content entered through the console is usually not particularly large, we will `not` use `multipart` uploads to upload content.

Then what method do we use to upload content?

We will first merge the user's input line by line into a `single slice`, and then upload the slice to the specified service when the user types `Ctrl C`.

## Compatibility

N/A

## Implementation

- Implement tee support