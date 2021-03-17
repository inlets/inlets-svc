// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type InletsClientConfig struct {
	Upstreams   []string `json":upstreams"`
	URL         string   `json":url"`
	Token       string   `json":token"`
	LicenseFile string   `json:"license-file"`
	AutoTLS     bool     `json:"auto-tls"`
}

type myservice struct {
	InletsClientConfig
	Process *exec.Cmd
}

func newService() (*myservice, error) {
	m := &myservice{}
	c := InletsClientConfig{}

	b, err := ioutil.ReadFile("C:\\inlets.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	elog.Info(1, fmt.Sprintf("Should we start here?"))
	m.InletsClientConfig = c

	upstreams := "--upstream=" + strings.TrimRight(strings.Join(m.Upstreams, ","), ",")
	args := []string{
		"http",
		"client",
		"--url=" + m.URL,
		upstreams,
		"--token=" + m.Token,
		"--license-file=" + m.LicenseFile,
		"--auto-tls=" + strconv.FormatBool(m.AutoTLS),
	}

	elog.Info(1, fmt.Sprintf("inlets-pro %v", args))
	cmd := exec.Command("inlets-pro", args...)

	m.Process = cmd

	err = m.Process.Start()
	if err != nil {
		elog.Error(1, fmt.Sprintf("Error starting app %s", err.Error()))
	}

	elog.Info(1, fmt.Sprintf("PID %d", cmd.Process.Pid))
	return m, nil
}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case <-tick:
			//beep()
		//	elog.Info(1, fmt.Sprintf("beep, %v", m))
		case c := <-r:
			switch c.Cmd {

			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				testOutput := strings.Join(args, "-")
				testOutput += fmt.Sprintf("-%d", c.Context)
				elog.Info(1, testOutput)
				err := m.Process.Process.Kill()
				if err != nil {
					elog.Info(1, "Error killing process: "+err.Error())
				}
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}

	m, err := newService()
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}

	err = run(name, m)
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
