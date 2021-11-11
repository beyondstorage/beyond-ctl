package operations

import (
	"fmt"
	"time"

	"go.beyondstorage.io/v5/types"
)

func (so *SingleOperator) Sign(path string, expire time.Duration) (url string, err error) {
	signer, ok := so.store.(types.StorageHTTPSigner)
	if !ok {
		return "", fmt.Errorf("storage http signer unimplement")
	}

	req, err := signer.QuerySignHTTPRead(path, expire)
	if err != nil {
		return "", err
	}

	url = req.URL.String()

	return
}
