package main

import (
	"testing"
)

func TestExecute(t *testing.T) {
	setupSignalHandler = func() (stopCh <-chan struct{}) {
		return make(chan struct{})
	}

	bootInit = func() {
		t.Logf("Run boot Init")
	}

	bootRun = func(stopCh <-chan struct{}) {
		t.Logf("Run boot Run with stopCh")
	}

	main()
}
