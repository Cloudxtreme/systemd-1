// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

// From github.com/coreos/go-systemd/util/ but no cgo
package util

import "os"

// IsRunningSystemd checks whether the host was booted with systemd as its init
// system. This functions similarly to systemd's `sd_booted(3)`: internally, it
// checks whether /run/systemd/system/ exists and is a directory.
// http://www.freedesktop.org/software/systemd/man/sd_booted.html
func IsRunningSystemd() bool {
	fi, err := os.Lstat("/run/systemd/system")
	if err != nil {
		return false
	}
	return fi.IsDir()
}
