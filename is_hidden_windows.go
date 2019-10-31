// is_hidden_windows.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2019-10-31
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2019-10-31

// +build windows
package main

import "syscall"

// Check whether a file is hidden or not. (For Windows)
func isHidden(filename string) bool {
	p, _ := syscall.UTF16PtrFromString(filename)
	attr, _ := syscall.GetFileAttributes(p)
	return attr&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}
