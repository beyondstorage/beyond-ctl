- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-10-28
- RFC PR: [beyondstorage/beyond-ctl#83](https://github.com/beyondstorage/beyond-ctl/pull/83)
- Tracking Issue: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)

# BCP-83: Add Remove Multipart Support

## Background

We have already implemented the remove files and directories operation, but the remove `multipart` operation is not yet implemented. 

Therefore, the user cannot use the remove `multipart` operation, which is extremely inconvenient for the user. And this may cause the user service to have a lot of incomplete multipart objects left.

## Proposal

So I propose to add the remove multipart operation. For this case, I suggest adding a string flag `multipart` to `rm`.

```go
&cli.StringFlag{
    Name: "multipart",
    Aliases: []string{
        "m",
        "M",
    },
    Usage: "remove multipart object, the value is multipartID",
}
```

### Command

The command will be like:

```
byctl rm --multipart=<multipartID> <source>
```

For example:

```
byctl rm --multipart=93B75FF42AF64558A6D440FB12511287 example:testMultipart
```

This command will delete the multipart object in service `example` with path `testMultipart` and `multipartID` of `93B75FF42AF64558A6D440FB12511287`. Where the service `example` is a service added using command `byctl profile add`.

### When the path or multipartID does not exist

In this case the command will not report an error. According to [GSP-46](https://github.com/beyondstorage/go-storage/blob/master/docs/rfcs/46-idempotent-delete.md), the `ObjectNotExit` related error should be omitted, especially when remove `multipart` object.

## Rationale

### Why use StringFlag?

We need to pass in the `multipartID`, but if we pass in two parameters directly, the original `rm` operation will be affected and not very readable.

Using `StringFlag` makes it easier to understand and smoother for users to use. And it doesn't add too much to the length of the command.

It is also faster to get the multipartID entered by the user, and easier to determine whether the user has entered a multipartID or not.

## Compatibility

N/A

## Implementation

- Implement rm multipart support

