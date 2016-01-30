// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package watchdog

import (
	"os"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/coreos/go-systemd/util"
	"github.com/fcavani/e"
	"github.com/fcavani/log"
)

const ErrNotRunning = "systemd is not running"
const ErrNoWatchdog = "no watchdog enabled for this daemon"
const ErrInvPeriode = "invalid periode"

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
		return e.Push(err, ErrInvPeriode)
	}
	wPerHalf := time.Duration(int(wPerInt64) / 2)
	log.Tag("systemd", "watchdog").Println("Starting the watchdog.")
	stop = make(chan struct{})
	// Start the periodic pings
	go func() {
		select {
		case <-stop:
			log.Tag("systemd", "watchdog").Println("By request watchdog is stoping.")
			return
		case <-time.After(wPerHalf):
			// Send the ping.
			log.DebugLevel().Tag("systemd", "watchdog").Println("Ping.")
			daemon.SdNotify(sdState)
		}
	}()
	return
}
