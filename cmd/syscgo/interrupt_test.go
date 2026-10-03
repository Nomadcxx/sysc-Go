package main

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestDeliverInterruptAfterStopDoesNotPanic(t *testing.T) {
	w := newInterruptWatcher()
	w.stop()
	w.deliverInterrupt()
}

func TestStopIsIdempotent(t *testing.T) {
	w := newInterruptWatcher()
	w.stop()
	w.stop()
}

func TestStopUnblocksHandler(t *testing.T) {
	w := newInterruptWatcher()
	sig := make(chan os.Signal)
	exited := make(chan struct{})
	go func() {
		defer close(exited)
		w.handle(sig)
	}()

	w.stop()
	select {
	case <-exited:
	case <-time.After(time.Second):
		t.Fatal("handler did not exit after stop")
	}
}

func TestStopAndSignalRaceDoesNotPanic(t *testing.T) {
	for i := 0; i < 100; i++ {
		w := newInterruptWatcher()
		sig := make(chan os.Signal, 1)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			w.handle(sig)
		}()
		go func() {
			defer wg.Done()
			w.stop()
			sig <- syscall.SIGTERM
		}()
		wg.Wait()
	}
}

func TestHandleSignalNotifiesQuit(t *testing.T) {
	w := newInterruptWatcher()
	sig := make(chan os.Signal, 1)
	sig <- syscall.SIGTERM
	w.handle(sig)

	select {
	case <-w.quit:
	case <-time.After(time.Second):
		t.Fatal("expected quit after signal")
	}
}
