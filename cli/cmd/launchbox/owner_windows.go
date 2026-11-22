//go:build windows

package main

// fileOwnerUID is not supported on Windows.
func fileOwnerUID(path string) (int, bool) {
	return 0, false
}
