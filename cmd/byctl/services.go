package main

import (
	_ "go.beyondstorage.io/services/azblob/v3"
	_ "go.beyondstorage.io/services/cos/v3"
	_ "go.beyondstorage.io/services/dropbox/v3"
	_ "go.beyondstorage.io/services/fs/v4"
	_ "go.beyondstorage.io/services/ftp"
	_ "go.beyondstorage.io/services/gcs/v3"
	_ "go.beyondstorage.io/services/ipfs"
	_ "go.beyondstorage.io/services/kodo/v3"
	_ "go.beyondstorage.io/services/memory"
	_ "go.beyondstorage.io/services/minio"
	_ "go.beyondstorage.io/services/oss/v3"
	_ "go.beyondstorage.io/services/qingstor/v4"
	_ "go.beyondstorage.io/services/s3/v3"
	_ "go.beyondstorage.io/services/uss/v3"
)
