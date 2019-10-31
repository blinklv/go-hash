// is_hidden.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2019-10-31
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2019-10-31

// +build !windows
package main

// Check whether a file is hidden or not. (For Unix-Like System)
func isHidden(filename string) bool {
	return filename[0:1] == "."
}
