package operations

import (
	"fmt"

	"go.beyondstorage.io/v5/types"
)

const (
	defaultMultipartPartSize int64 = 128 * 1024 * 1024 // 128M
)

func calculatePartSize(store types.Storager, totalSize int64) (int64, error) {
	maxNum, numOK := store.Metadata().GetMultipartNumberMaximum()
	maxSize, maxOK := store.Metadata().GetMultipartSizeMaximum()
	minSize, minOK := store.Metadata().GetMultipartSizeMinimum()

	partSize := defaultMultipartPartSize

	// assert object size too large
	if maxOK && numOK {
		if totalSize > maxSize*int64(maxNum) {
			return 0, fmt.Errorf("calculate part size failed: {object size too large}")
		}
	}

	// if multipart number no restrict, just check max and min multipart part size
	if !numOK {
		for {
			if maxOK && partSize > maxSize {
				partSize = partSize >> 1
				continue
			}
			// if part size >= total size, consider total object as last part, which is not limited by minimum part size
			if partSize < totalSize {
				// otherwise, if partSize less than minimum, double part size
				if minOK && partSize < minSize {
					partSize = partSize << 1
					continue
				}
			} else {
				partSize = totalSize
				break
			}
			break
		}
		return partSize, nil
	}

	// if multipart number has maximum restriction, count part size dynamically
	for {
		// if part number count larger than maximum, double part size
		if totalSize/partSize >= int64(maxNum) {
			// objSize > maxSize * maxNum has been asserted before,
			// so we do not need check
			partSize = partSize << 1
			continue
		}

		// if part size >= total size, consider total object as last part, which is not limited by minimum part size
		if partSize < totalSize {
			// otherwise, if partSize less than minimum, double part size
			if minOK && partSize < minSize {
				partSize = partSize << 1
				continue
			}
		} else {
			partSize = totalSize
			break
		}

		// if partSize too large, try to count by the max part number
		if maxOK && partSize > maxSize {
			// Try to adjust partSize if it is too small and account for
			// integer division truncation.
			partSize = totalSize/int64(maxNum) + 1
			return partSize, nil
		}

		// otherwise, use the part size, which is not too large or too small
		break
	}
	return partSize, nil
}

// validatePartSize used to check user-input part size
func validatePartSize(stor types.Storager, totalSize, partSize int64) error {
	if min, ok := stor.Metadata().GetMultipartSizeMinimum(); ok && partSize < min {
		return fmt.Errorf("part size must larger than {%d}", min)
	}
	if max, ok := stor.Metadata().GetMultipartSizeMaximum(); ok && partSize > max {
		return fmt.Errorf("part size must less than {%d}", max)
	}
	num, ok := stor.Metadata().GetMultipartNumberMaximum()
	if ok {
		parts := totalSize / partSize
		if parts > int64(num) {
			return fmt.Errorf("parts count at part size <%d> "+
				"will be larger than max number {%d}", partSize, num)
		}
	}
	return nil
}
