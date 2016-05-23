// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package systemd

import (
	"sync"
	"time"

	sd "github.com/coreos/go-systemd/dbus"
	"github.com/fcavani/e"
	"github.com/fcavani/log"
)

const ErrInvServName = "invalid service name"

type ProcessStatus uint8

const (
	Unknow ProcessStatus = iota
	Running
	Stopped
	Deleted
	Other
)

func (p ProcessStatus) String() string {
	switch p {
	case Unknow:
		return "unknow"
	case Running:
		return "running"
	case Stopped:
		return "stopped"
	case Deleted:
		return "deleted"
	case Other:
		return "other"
	default:
		return "invalid status"
	}
}

type ServiceStatus struct {
	// Name of the service
	Name string
	// Interval betwen probes.
	Interval time.Duration
	status   ProcessStatus
	lck      sync.Mutex
	conn     *sd.Conn
}

func (s *ServiceStatus) Init() error {
	var err error
	if s.Name == "" {
		return e.New(ErrInvServName)
	}
	s.conn, err = sd.NewSystemdConnection()
	if err != nil {
		return e.Forward(err)
	}
	err = s.conn.Subscribe()
	if err != nil {
		return e.Forward(err)
	}
	ch, cherr := s.conn.SubscribeUnits(s.Interval)
	go func() {
		for {
			select {
			case status := <-ch:
				if status == nil {
					continue
				}
				st, found := status[s.Name]
				if !found {
					log.Tag("systemd").Printf("Status of %v is unknow.", s.Name)
					s.setStatus(Unknow)
					continue
				}
				if s == nil {
					s.setStatus(Deleted)
					continue
				}
				if st.LoadState == "loaded" && st.ActiveState == "active" && st.SubState == "running" {
					s.setStatus(Running)
				} else if st.LoadState == "loaded" && st.ActiveState == "active" && st.SubState == "active" {
					s.setStatus(Running)
				} else if st.LoadState == "not-found" && st.ActiveState == "active" && st.SubState == "exited" {
					s.setStatus(Stopped)
				} else if st.LoadState == "loaded" && st.ActiveState == "inactive" && st.SubState == "dead" {
					s.setStatus(Stopped)
				} else {
					//ActiveState: deactivating, LoadState: loaded, SubState: stop-sigterm
					log.Tag("systemd").Printf("ActiveState: %v, LoadState: %v, SubState: %v", st.ActiveState, st.LoadState, st.SubState)
					s.setStatus(Other)
				}
			case err := <-cherr:
				log.Tag("systemd", "service", "status").Fatalln("SubscribeUnits error:", err)
			}
		}
	}()
	return nil
}

func (s *ServiceStatus) setStatus(ps ProcessStatus) {
	s.lck.Lock()
	defer s.lck.Unlock()
	s.status = ps
}

func (s *ServiceStatus) GetStatus() ProcessStatus {
	s.lck.Lock()
	defer s.lck.Unlock()
	return s.status
}

func (s *ServiceStatus) Close() error {
	err := s.conn.Unsubscribe()
	if err != nil {
		return e.Forward(err)
	}
	s.conn.Close()
	return nil
}
