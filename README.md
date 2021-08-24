# BeyondCTL

BeyondCTL is a command-line tool for all storage services.

## Status

We are focusing on the v0.1.0 release.

See our [v0.1.0 Roadmap](https://github.com/beyondstorage/beyond-ctl/issues/1) to keep sync.

We expect to have basic support in v0.1.0, including:

- profile (different profiles for different storage services)
- ls (list files/dirs with *color*, *long format*)
- cp (copy files/dirs across different storage services)
- rm (remove files/dirs in different storage services)
- stat (get file or storage info)
- cat (pipe data from storage services into `stdout`)
- tee (pipe data from `stdin` into storage services)

## Features

### Fully storage services support

`BeyondCTL` is powered by [go-storage](https://github.com/beyondstorage/go-storage), we will support all stable services. Please sending PRs if we missed them.

### User and machine-friendly UX/UI

We will develop both user and machine-friendly UX/UI.

For example:

- we will build progress and interactive shell for users
- we will add json output support for the machine
- ...

## Call for help!

There are so much works to do, and we are welcome all PRs.

Please visit [issues](https://github.com/beyondstorage/beyond-ctl/issues) to know what kind of help to provide.
