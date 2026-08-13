//go:build !windows

package source

import (
	"os"
	"syscall"
)

func fileLinkCount(info os.FileInfo) uint64 {
	metadata, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0
	}
	if metadata.Nlink > 1 {
		return 2
	}
	return 1
}
