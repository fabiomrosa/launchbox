//go:build !windows

package main

import (
	"os"
	"syscall"
)

// fileOwnerUID returns the uid owning path, and false if it can't be determined.
func fileOwnerUID(path string) (int, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return int(stat.Uid), true
}
