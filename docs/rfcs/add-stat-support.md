- Author: abyss-w <mad.hatter@foxmail.com>
- Start Date: 2021-10-19
- RFC PR: [beyondstorage/beyond-ctl#0](https://github.com/beyondstorage/beyond-ctl/issues/0)
- Tracking Issue: [beyondstorage/beyond-ctl#55](https://github.com/beyondstorage/beyond-ctl/issues/55)

# BCP-0: Add Stat Support

## Background

In order to facilitate access to basic information about the file or storage, we decided to support stat operations.

## Proposal

I propose to add stat support for users to get basic information about a given file or storage.

For file, the following fields are supported:

```go
type fileMessage struct {
	key           string           // file name
	mode          types.ObjectMode // file mode
	lastModified  time.Time        // lastModified
	contentLength int64            // ContentLength
	etag          string           // Etag
	contentType   string           // ContentType
}
```

For storage, the following fields are supported:

```go
type storageMessage struct {
	service    string // service name
	bucketName string // bucket name
	workDir    string // work dir
	location   string // bucket location
}
```

- The field `location` is not required. `location` is only valid when the `storage metadata` is set to `location`.

### Command Options

```
 --directory, -d, -D     stat directory(default: false)
```

### Command

The command will be like:

```
byctl stat [command options] [path]
byctl stat -d [dir path]
```

For example `stat` local file `test.mp4`:

```
byctl stat test.mp4
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

#### minio [mc](https://github.com/minio/mc)

The fields supported by mc stat file are:

```go
type statMessage struct {
	Status            string            `json:"status"`
	Key               string            `json:"name"`
	Date              time.Time         `json:"lastModified"`
	Size              int64             `json:"size"`
	ETag              string            `json:"etag"`
	Type              string            `json:"type,omitempty"`
	Expires           time.Time         `json:"expires,omitempty"`
	Expiration        time.Time         `json:"expiration,omitempty"`
	ExpirationRuleID  string            `json:"expirationRuleID,omitempty"`
	ReplicationStatus string            `json:"replicationStatus,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	VersionID         string            `json:"versionID,omitempty"`
	DeleteMarker      bool              `json:"deleteMarker,omitempty"`
	singleObject      bool
}
```

The fields supported by mc stat bucket are:

```go
type bucketInfoMessage struct {
	Op       string
	URL      string     `json:"url"`
	Status   string     `json:"status"`
	Metadata BucketInfo `json:"metadata"`
}

// BucketInfo holds info about a bucket
type BucketInfo struct {
	URL        ClientURL   `json:"-"`
	Key        string      `json:"name"`
	Date       time.Time   `json:"lastModified"`
	Size       int64       `json:"size"`
	Type       os.FileMode `json:"-"`
	Versioning struct {
		Status    string `json:"status"`
		MFADelete string `json:"MFADelete"`
	} `json:"Versioning,omitempty"`
	Encryption struct {
		Algorithm string `json:"algorithm,omitempty"`
		KeyID     string `json:"keyId,omitempty"`
	} `json:"Encryption,omitempty"`
	Locking struct {
		Enabled  string              `json:"enabled"`
		Mode     minio.RetentionMode `json:"mode"`
		Validity string              `json:"validity"`
	} `json:"ObjectLock,omitempty"`
	Replication struct {
		Enabled bool               `json:"enabled"`
		Config  replication.Config `json:"config,omitempty"`
	} `json:"Replication"`
	Policy struct {
		Type string `json:"type"`
		Text string `json:"policy,omitempty"`
	} `json:"Policy,omitempty"`
	Location string            `json:"location"`
	Tagging  map[string]string `json:"tagging,omitempty"`
	ILM      struct {
		Config *lifecycle.Configuration `json:"config,omitempty"`
	} `json:"ilm,omitempty"`
	Notification struct {
		Config notification.Configuration `json:"config,omitempty"`
	} `json:"notification,omitempty"`
}
```

## Compatibility

N/A

## Implementation

- Implement stat support

