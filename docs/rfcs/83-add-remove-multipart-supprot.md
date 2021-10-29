- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-10-28
- RFC PR: [beyondstorage/beyond-ctl#83](https://github.com/beyondstorage/beyond-ctl/pull/83)
- Tracking Issue: [beyondstorage/beyond-ctl#85](https://github.com/beyondstorage/beyond-ctl/issues/85)

# BCP-83: Add Remove Multipart Support

## Background

We have already implemented the remove files and directories operation, but the remove `multipart` operation is not yet implemented.

Therefore, the user cannot use the remove `multipart` operation, which is extremely inconvenient for the user. And this may cause the user service to have a lot of incomplete multipart objects left.

## Proposal

So I propose to add the remove multipart operation. For this case, I suggest adding a string flag `multipart` to `rm`.

```go
&cli.BoolFlag{
Name: "multipart",
Usage: "remove multipart object",
}
```

### Command

The command will be like:

```
byctl rm --multipart <source>
```

For example:

```
byctl rm --multipart example:testMultipart
```

This command will delete all multipart objects in service `example` with path `testMultipart`. Where service `example` is a service added with the `byctl profile add` command.

We can also use it with the flag `-r` in rm, for example:

```
byctl rm --multipart -r example:testMultipart
```

This command will delete all multipart objects in the service `example` with the prefix `multipart`.

### When the path or multipartID does not exist

In this case the command will not report an error. According to [GSP-46](https://github.com/beyondstorage/go-storage/blob/master/docs/rfcs/46-idempotent-delete.md), the `ObjectNotExit` related error should be omitted, especially when remove `multipart` object.

## Rationale

### Why use BoolFlag?

For users, it is very difficult to get the multipartID. So we use `BoolFlag` to remove all multipart objects with path `<source>`. We can also use it better with flag `-r`, which allows users to delete all multipart objects prefixed with `<source>`.

Also, for the user, they are supposed to care about the multipart object with path `<source>`, not the `multipartID`.

## Compatibility

N/A

## Implementation

- Implement rm multipart support
