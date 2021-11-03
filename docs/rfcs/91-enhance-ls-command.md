- Author: JinnyYi <github.com/JinnyYi>
- Start Date: 2021-11-02
- RFC PR: [beyondstorage/beyond-ctl#91](https://github.com/beyondstorage/beyond-ctl/issues/91)
- Tracking Issue: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)

# BCP-91: Enhance ls Command

## Background

The `ls` command is used to list files or directories in Linux and other Unix-based operating systems. And it is used to list objects and common prefixes under a prefix in [Amazon CLI] and [Aliyun CLI].

Currently, byctl already supports the `ls` command, but the capabilities are limited. For example, it is not possible to recursively traverse a directory and retrieve summary statistics for a directory.

## Proposal

So I propose to add more command options for `ls` to change the way files or directories are listed in the user terminal.

### Synopsis

```
byctl ls [command options] [arguments...]
```

### Options

```
-a, --all
    list all files or objects, and in-progress multipart uploads
--format
    short --short
--multipart
    list in-progress multipart uploads
-R, --recursive
    list subdirectories recursively
--short
    show by short format
--summarize
    display summary information (number of files, total size in human readable format)
```

### Usage

The command list files or objects, with simple additional information (long format) by default, including object mode, size, last modified time and name. If --short option is specified, byctl will show by short format.

- all type
  - -a, --all option will show files or objects and multipart upload tasks, whose file or object name starts with the specified prefix.
  - It returns top-level subdirectory names instead of contents of the subdirectory, which in default show by long format.
  - The usage also supports --short, -R and --summarize options.

- multipart option
  - --multipart option will show the in-progress multipart upload tasks, whose object name starts with the specified prefix. byctl will show the init time and upload id meanwhile.
  - The usage also supports --short and --summarize options.

- recursive option
  - -R, --recursive option means list command will be performed on all files or objects under the specified directory or prefix.
  - The usage also supports -a, --short and --summarize options.
  
- short format
  - --short option ignores all additional information, which means that only the file or object name and upload id for multipart upload task are displayed.
  - The usage also supports -a, --multipart, and -R options.

- summarize option
  - --summarize will display summary information, including number of files and total size.
  - The usage also supports -a or --multipart, and -R options.

### Examples

```
1)  byctl ls s3:
    dir            0 Nov 02 06:53 dir1
    read      903899 Nov 02 06:53 obj1
    read          18 Nov 02 01:51 obj2
    
2)  byctl ls --short s3:
    dir1 obj1 obj2
   
3)  byctl ls --multipart s3:
    part 15754AF7980C4DFB8193F190837520BB      0 Nov 02 06:54 obj3
    part 3998971ACAF94AD9AC48EAC1988BE863      0 Nov 02 01:55 obj4
    
4)  byctl ls --multipart --short s3:
    15754AF7980C4DFB8193F190837520BB obj3
    3998971ACAF94AD9AC48EAC1988BE863 obj4

5)  byctl ls -a s3:
    dir                                        0 Nov 02 06:53 dir1
    read                                  903899 Nov 02 06:53 obj1
    read                                      18 Nov 02 01:51 obj2
    part 15754AF7980C4DFB8193F190837520BB      0 Nov 02 06:54 obj3
    part 3998971ACAF94AD9AC48EAC1988BE863      0 Nov 02 01:55 obj4
    
6)  byctl ls -a --short s3:
    dir1 obj1 obj2 
    15754AF7980C4DFB8193F190837520BB obj3
    3998971ACAF94AD9AC48EAC1988BE863 obj4
   
7)  byctl ls --summarize s3:
    dir                                        0 Nov 02 06:53 dir1
    read                                  903899 Nov 02 06:53 obj1
    read                                      18 Nov 02 01:51 obj2
    Total Objects: 3
       Total Size: 882.7 KiB
   
8)  byctl ls -a --summarize s3:
    dir                                        0 Nov 02 06:53 dir1
    read                                  903899 Nov 02 06:53 obj1
    read                                      18 Nov 02 01:51 obj2
    part 15754AF7980C4DFB8193F190837520BB      0 Nov 02 06:54 obj3
    part 3998971ACAF94AD9AC48EAC1988BE863      0 Nov 02 01:55 obj4
   
    Total Objects: 5
       Total Size: 882.7 KiB
   
9)  byctl ls -R file:test
    dir            0 Nov 02 06:53 dir1
    read           9 Nov 02 07:11 dir1/test
    read      903899 Nov 02 06:53 obj1
    read          18 Nov 02 01:51 obj2
    
10)  byctl ls -R --summarize file:test
    dir            0 Nov 02 06:53 dir1
    read           9 Nov 02 07:11 dir1/test
    read      903899 Nov 02 06:53 test1
    read          18 Nov 02 01:51 test2
   
    Total Objects: 4
       Total Size: 882.7 KiB
       
11) byctl ls -a -R --summarize file:test
    dir                                        0 Nov 02 06:53 dir1
    read                                       9 Nov 02 07:11 dir1/test
    read                                  903899 Nov 02 06:53 obj1
    read                                      18 Nov 02 01:51 obj2
    part 15754AF7980C4DFB8193F190837520BB      0 Nov 02 06:54 obj3
    part 3998971ACAF94AD9AC48EAC1988BE863      0 Nov 02 01:55 obj4
   
    Total Objects: 6
       Total Size: 882.7 KiB 
```

## Rationale

### Other Comparable Implementation

#### [Amazon CLI]

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
>
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

#### [Aliyun CLI]

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
1) ossutil ls oss://bucket1 -a 
   LastModifiedTime              Size(B)  StorageClass   ETAG                              ObjectName
   2015-06-05 14:06:29 +0000 CST  201933      Standard   7E2F4A7F1AC9D2F0996E8332D5EA5B41  oss://bucket1/dir1/obj11
   2015-06-05 14:36:21 +0000 CST  201933      Standard   6185CA2E8EB8510A61B3A845EAFE4174  oss://bucket1/obj1
   2016-04-08 14:50:47 +0000 CST 6476984      Standard   4F16FDAE7AC404CEC8B727FCC67779D6  oss://bucket1/sample.txt
   Object Number is: 3
   InitiatedTime                  UploadID                          ObjectName
   2017-01-13 03:45:26 +0000 CST  15754AF7980C4DFB8193F190837520BB  oss://bucket1/obj1
   2017-01-13 03:43:13 +0000 CST  2A1F9B4A95E341BD9285CC42BB950EE0  oss://bucket1/obj1
   2017-01-13 03:45:25 +0000 CST  3998971ACAF94AD9AC48EAC1988BE863  oss://bucket1/obj2
   2017-01-20 11:16:21 +0800 CST  A20157A7B2FEC4670626DAE0F4C0073C  oss://bucket1/tobj
   UploadID Number is: 4
   
2) ossutil ls oss://bucket1/obj -a -s 
   oss://bucket1/obj1
   bject Number is: 1
   UploadID                          ObjectName
   15754AF7980C4DFB8193F190837520BB  oss://bucket1/obj1
   2A1F9B4A95E341BD9285CC42BB950EE0  oss://bucket1/obj1
   3998971ACAF94AD9AC48EAC1988BE863  oss://bucket1/obj2
   UploadID Number is: 3
```

#### More possible command options

```
-C  list entries by columns
--format 
    across -x, commas -m, horizontal -x, single-column -1, verbose -l, vertical -C
-m  fill width with a comma separated list of entries
-s, --size
    print the allocated size of each file, in blocks
-t  sort by modification time, newest first
-x  list entries by lines instead of by columns
-1  list one file per line. Avoid '\n' with -q or -b
```

## Compatibility

N/A

## Implementation

- Implement options for `ls`

[Amazon CLI]: https://docs.aws.amazon.com/cli/latest/reference/s3/ls.html
[Aliyun CLI]: https://github.com/aliyun/aliyun-cli/blob/master/oss/lib/ls.go
