- Author: JinnyYi <github.com/JinnyYi>
- Start Date: 2021-11-02
- RFC PR: [beyondstorage/beyond-ctl#91](https://github.com/beyondstorage/beyond-ctl/issues/91)
- Tracking Issue: [beyondstorage/beyond-ctl#94](https://github.com/beyondstorage/beyond-ctl/issues/94)

# BCP-91: Add Summarize Support in ls

## Background

The `ls` command is used to list files or directories in Linux and other Unix-based operating systems.

Currently, byctl already supports the `ls` command to list files or objects with simple additional information (long format):

```shell
> ./byctl ls -l s3:
read      903899 Nov 02 06:53 obj1
read          18 Nov 02 01:51 obj2
```

There is no intuitive information to show summary statistics for a directory.

## Proposal

So I propose to add summarize support in `ls` to display summary information. For this case, I suggest adding a string flag `summarize` to `ls`.

```go
&cli.BoolFlag{
	Name:  "summarize", 
	Usage: "display summary information",
}
```

### Command

Usage:

```
byctl ls [command options] <source>
```

Options:

```
--summarize    display summary information (default: false)
```

- `summarize` option will display the total number of objects and total size at the end of the result listing.

### Examples

```
1)  byctl ls -l s3:
    dir            0 Nov 02 06:53 dir1
    read      903899 Nov 02 06:53 obj1
    read          18 Nov 02 01:51 obj2
   
2)  byctl ls -l --summarize s3:
    dir            0 Nov 02 06:53 dir1
    read      903899 Nov 02 06:53 obj1
    read          18 Nov 02 01:51 obj2
   
    Total Objects: 3
       Total Size: 882.7 KiB 
```

## Rationale

### Other Comparable Implementation

#### [Amazon CLI](https://docs.aws.amazon.com/cli/latest/reference/s3/ls.html)

Synopsis

```
 ls
<S3Uri> or NONE
[--recursive]
[--page-size <value>]
[--human-readable]
[--summarize]
[--request-payer <value>]
```

Example

```shell
> aws s3 ls s3://bucketnanme --recursive --human-readable --summarize
2013-09-02 21:37:53   10 Bytes a.txt
2013-09-02 21:37:53  2.9 MiB foo.zip
2013-09-02 21:32:57   23 Bytes foo/bar/.baz/a
2013-09-02 21:32:58   41 Bytes foo/bar/.baz/b
2013-09-02 21:32:57  281 Bytes foo/bar/.baz/c
2013-09-02 21:32:57   73 Bytes foo/bar/.baz/d
2013-09-02 21:32:57  452 Bytes foo/bar/.baz/e
2013-09-02 21:32:57  896 Bytes foo/bar/.baz/hooks/bar
2013-09-02 21:32:57  189 Bytes foo/bar/.baz/hooks/foo
2013-09-02 21:32:57  398 Bytes z.txt

Total Objects: 10
   Total Size: 2.9 MiB
```

#### [Aliyun CLI](https://github.com/aliyun/aliyun-cli/blob/master/oss/lib/ls.go)

syntaxText:

```go
var specEnglishList = SpecText{
	synopsisText: "List Buckets or Objects", 
	paramText: "[cloud_url] [options]", 
	syntaxText: `ossutil ls [oss://bucket[/prefix]] [-s] [-d] [-m] [--limited-num num] [--marker marker] [--upload-id-marker umarker] [--payer requester] [--include include-pattern] [--exclude exclude-pattern]  [--version-id-marker id_marker] [--all-versions]  [-c file]`,
}
```

sampleText:

```
1) ossutil ls oss://bucket1
   LastModifiedTime              Size(B)  StorageClass   ETAG                              ObjectName
   2015-06-05 14:06:29 +0000 CST  201933      Standard   7E2F4A7F1AC9D2F0996E8332D5EA5B41  oss://bucket1/dir1/obj11
   2015-06-05 14:36:21 +0000 CST  201933      Standard   6185CA2E8EB8510A61B3A845EAFE4174  oss://bucket1/obj1
   2016-04-08 14:50:47 +0000 CST 6476984      Standard   4F16FDAE7AC404CEC8B727FCC67779D6  oss://bucket1/sample.txt
   Object Number is: 3
```

## Compatibility

N/A

## Implementation

- Implement summarize in `ls`
