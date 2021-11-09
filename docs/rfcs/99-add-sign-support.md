- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-11-08
- RFC PR: [beyondstorage/beyond-ctl#99](https://github.com/beyondstorage/beyond-ctl/pull/99)
- Tracking Issue: [beyondstorage/beyond-ctl#101](https://github.com/beyondstorage/beyond-ctl/issues/101)

# BCP-99: Add Sign Support

## Background

In [GSP-729](https://github.com/beyondstorage/go-storage/blob/master/docs/rfcs/729-redesign-http-signer.md), we propose
to add the `StorageHTTPSigner` interface, where `QuerySignHTTPRead` supports signing operations on read.

## Proposal

I propose that `BeyondCTL` support `sign` operation.

```
byctl sign [command options] [source]
```

`sign` generates a signed URL for the source object. Anyone who receives this URL within the given expiration time can
retrieve the object via an HTTP GET request.

### Command options

```
--expire value   the number of seconds until the signed URL expires (default: 300)
```

`--expire` is an `IntFlag`, in seconds, with a default value of 300 seconds.

### Example

```
byctl sign example:test.mp4
```

Executing the previous command will return a signed URL to download the file `test.mp4` from the service `example`. This
URL is valid for 300 seconds by default.

If you want to set your own expiration time, you can do so:

```
byctl sign --expire=150 example:test.mp4
```

At this point, the URL has an expiration time of 150 seconds.

## Rationale

N/A

## Compatibility

N/A

## Implementation

- Implement sign support
- Add the sign document to the [site](https://github.com/beyondstorage/site)
