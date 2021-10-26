- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-10-19
- RFC PR: [beyondstorage/beyond-ctl#75](https://github.com/beyondstorage/beyond-ctl/pull/75)
- Tracking Issue: [beyondstorage/beyond-ctl#55](https://github.com/beyondstorage/beyond-ctl/issues/55)

# BCP-75: Add Stat Support

## Background

In order to facilitate access to basic information about the file or storage, we decided to support stat operations.

## Proposal

I propose to add stat support for users to get basic information about a given file or storage. We support both normal output and `json` output.

For file, the following fields are supported:

```go
ID             // file absolute path
path           // file relative path
mode           // file mode
lastModified   // lastModified
contentLength  // ContentLength
etag           // Etag
contentType    // ContentType
userMetadata   // user metadata
systemMetadata // system metadata
```

For storage, the following fields are supported:

```go
service    // service name
bucketName // bucket name
workDir    // work dir
location   // bucket location
```

- The field `location` is not required. `location` is only valid when the `storage metadata` is set to `location`.

### Command Options

```
--json   Output in json format(defualt: false)
```

### Command

The command will be like:

```
byctl stat [command options] [path]
```

### Example

- `stat` local file `test.jpg`:

```
byctl stat test.jpg
```

After the above command is executed, the basic information of the file `test.jpg` will be displayed:

```
ID: path/to/workdir/test.jpg
Path: test.jpg
Mode: Read
LastModified: 2021-10-25 09:06:57 +0000 UTC
Etag: "7e4611c52075590896dd26905ac0c4cf"
ContentType: image/jpeg

SystemMetadata: 
StorageClass: STANDARD
xxxxxxxxx: xxx

UserMetadata: 
xxxx: xxxx
xxx: xxx   
```

If a file has two modes, we will output it in the following way:

```
Mode: Append|Part
```

- `stat` storager `example` (Added via profile, qingstor.):

```
byctl stat example:
```

When the user enters only service, but not the specified stat path, we consider this case to be a `stat` `storager`. Execute the above command and output the result:

```
Service: qingstor
BucketName: test-stat
WorkDir: /workdir/
Location: sh1a
```

## Rationale

### Why BeyondCTL Support these fields

Combining the implementation of each service with the available accessible fields and the user experience of other projects, the current field design is essentially the most logical.

### Other Comparable Implementation

#### [qsctl](https://github.com/qingstor/qsctl)

The fields supported by qsctl stat file are:

```
Key      -file name
Size     -total size, in bytes
Type     -file type
Etag     -content etag of the file
StorageClass 
UpdateAt -time of last data modification
```

The fields supported by qsctl stat storage are:

```
Name     -bucket name
Location -bucket location
Size     -total size, in bytes
Count    -count of files in this bucket
```

For example:

`stat` file:

```
:) qsctl stat qs://bucket/test.jpg
         Key: test.jpg
        Size: 31089B
        Type: image/jpeg
        ETag: "7e4611c52075590896dd26905ac0c4cf"
StorageClass: STANDARD
   UpdatedAt: 2021-10-21 03:03:11 +0000 UTC
```

`stat` storager:

```
:) qsctl stat qs://bucket/
    Name: bucket
Location: sh1a
    Size: 62207B
   Count: 4
```

#### minio [mc](https://github.com/minio/mc)

`stat` file:

```go
:) mc stat test.jpg
Name      : test.jpg
Date      : 2021-10-26 11:01:22 CST
Size      : 30 KiB
Type      : file
Metadata  :
Content-Type       : image/jpeg
X-Amz-Meta-Mc-Attrs: xxxxx
```

## Compatibility

N/A

## Implementation

- Implement stat support