//go:build windows

package source

import "os"

func fileLinkCount(os.FileInfo) uint64 {
	return 1
}
