// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package watchdog

import (
	"os"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/fcavani/e"
	"github.com/fcavani/log"
	"github.com/fcavani/systemd/util"
)

const ErrNotRunning = "systemd is not running"
const ErrNoWatchdog = "no watchdog enabled for this daemon"
const ErrInvPeriode = "invalid periode"
const ErrInvInterval = "invalid interval"

const sdState = "WATCHDOG=1"

// Watchdog setup the watchdog and start then. This functoin will
// comunicate with systemd sending the pings to it, if this fails
// to send the ping systemd will reatart this daemon.
func Watchdog() (stop chan struct{}, err error) {
	// Check if systemd exist.
	if !util.IsRunningSystemd() {
		return nil, e.New(ErrNotRunning)
	}
	// Get the periode and check if watchdog is on for this daemon
	wPeriodeµsec := os.Getenv("WATCHDOG_USEC")
	if wPeriodeµsec == "" {
		return nil, e.New(ErrNoWatchdog)
	}
	wPerInt64, err := strconv.ParseInt(wPeriodeµsec, 10, 32)
	if err != nil {
		return nil, e.Push(err, ErrInvPeriode)
	}
	wPerHalf := time.Duration(int(wPerInt64)/2) * time.Microsecond
	if wPerHalf <= 0 {
		return nil, e.New(ErrInvInterval)
	}
	log.Tag("systemd", "watchdog").Printf("Starting the watchdog with interval of %v.", wPerHalf)
	stop = make(chan struct{})
	// Start the periodic pings
	go func() {
		for {
			select {
			case <-stop:
				log.Tag("systemd", "watchdog").Println("By request watchdog is stoping.")
				return
			case <-time.After(wPerHalf):
				// Send the ping.
				log.DebugLevel().Tag("systemd", "watchdog").Println("Ping.")
				daemon.SdNotify(sdState)
			}
		}
	}()
	return
}
