- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-10-28
- RFC PR: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)
- Tracking Issue: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)

# BCP-0: Add Tee Support

## Background

The `Linux` `tee` command is used to read data from standard input and output its contents to a file. The `tee` command reads data from a standard input device, outputs its contents to a standard output device, and also saves it to a file.

We will refer to the `Linux` `tee` command to implement support for the `BeyondCTL` `tee` command.

## Proposal

I propose to add support for `tee` command to `BeyondCTL`. `BeyondCTL` will tee the content to stdout like `Linux` `tee` command does.

```
byctl tee [command options] [target]
```

This command will be terminated by typing `Ctrl C`, after which the user will be prompted that the input was saved successfully.

### Example

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

### How to capture the `Ctrl C`?

We use `golang's` `signal` library and the `syscall.SIGINT` and `syscall.SIGTERM` parameters to capture the terminal triggered by typing `Ctrl C`.

| Signal  |                     Description                     |
| :-----: | :-------------------------------------------------: |
| SIGINT  |    User sends INTR character (Ctrl+C) to trigger    |
| SIGTERM | End the program (can be caught, blocked or ignored) |

### What do we use to upload the content to the specified service?

Since `multipart` uploads basically have a minimum byte limit, and the content entered through the console is usually not particularly large, we will `not` use `multipart` uploads to upload content.

Then what method do we use to upload content?

We will first merge the user's input line by line into a `single slice`, and then upload the slice to the specified service when the user types `Ctrl C`.

## Compatibility

N/A

## Implementation

- Implement tee support

