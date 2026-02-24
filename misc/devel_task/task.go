// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

// Package task is a manager for tasks
// that need to shut down gracefully
// before the program exits.
//
// It's mainly a solution to the problem of child processes
// getting orphaned and lingering on after the Go program
// has been stopped by a signal.
//
// This mechanism only covers the program exiting
// due to signals or ending naturally.
// Panics or os.Exit calls will still leave orphaned child processes;
// handling those as well would require mechanisms
// such as spawning a child reaper process.
//
// Note that since the GlobalManager is supposed to be rendered controllable
// in the main() function, and since all the init() functions run before main(),
// it's futile to use the task manager in init() functions.
//
// # Usage
//
//	main() {
//	    stop, wait := task.StopTasksWhenKilled()
//	    defer func() {
//	        stop()
//	        wait()
//	    }
//	    ...
//	}
package task

// TODO: Can [UnhandleSignals] be returned as a closure by [HandleSignals]?

import (
	"context"
	"os"
	"sync"
	"syscall"
)

type Manager interface {
	Wg() WaitGroup
	Ctx() context.Context
}

type WaitGroup interface {
	Add(int)
	Done()
	Go(func())
}

var GlobalManager Manager = NewFakeManager()

func Wg() WaitGroup {
	return GlobalManager.Wg()
}

func Ctx() context.Context {
	return GlobalManager.Ctx()
}

var (
	globalCancel    func()
	globalWait      func()
	unhandleSignals func()
)

func NewManager() (m Manager, stop, wait func()) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	return &manager{wg, ctx}, cancel, wg.Wait
}

func NewFakeManager() Manager {
	return &manager{
		wg:  fakeWaitGroup{},
		ctx: context.Background(),
	}
}

func ManageTasks() (cancelTasks, waitForTasks func()) {
	if globalCancel != nil {
		panic("task.ControlGlobalManager: already controlled")
	}
	GlobalManager, globalCancel, globalWait = NewManager()
	return globalCancel, globalWait
}

// HandleSignals will capture the typical process termination signals.
// Upon receiving one such signal,
// the [GlobalManager]'s context is canceled,
// then it's wait group is waited on,
// then [os.Exit] is called.
//
// This function calls [ManageTasks]
// if it hasn't been already.
func HandleSignals() (cancelTasks, waitForTasks func()) {
	if globalCancel == nil {
		ManageTasks()
	}
	if unhandleSignals != nil {
		panic("task.HandleSignals: already handling signals")
	}
	sigChan, _unhandleSignals := trap() // trap is defined internally. It installs signal handlers.
	unhandleSignals = _unhandleSignals
	go func() {
		sig, ok := <-sigChan
		if !ok {
			return
		}
		globalCancel()
		globalWait()
		os.Exit(signalExitCode(sig))
	}()
	return globalCancel, globalWait
}

// UnhandleSignals reverts the effects of [HandleSignals].
func UnhandleSignals() {
	if unhandleSignals == nil {
		panic("task.UnhandleSignals: HandleSignals not called yet")
	}
	unhandleSignals()
}

// signalExitCode returns a conventional exit code
// based on the given signal that the program received.
//
// https://www.gnu.org/software/bash/manual/html_node/Exit-Status.html
func signalExitCode(sig os.Signal) int {
	if sigNum, ok := sig.(syscall.Signal); ok {
		return 128 + int(sigNum)
	}
	return 1
}

type manager struct {
	wg  WaitGroup
	ctx context.Context
}

func (m manager) Wg() WaitGroup {
	return m.wg
}

func (m manager) Ctx() context.Context {
	return m.ctx
}

type fakeWaitGroup struct {
	WaitGroup
}

func (fakeWaitGroup) Go(fn func()) {
	go fn()
}
