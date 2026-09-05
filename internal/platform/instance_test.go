package platform

import (
	"path/filepath"
	"testing"
	"time"
)

func TestHandoverDeliversShowRequest(t *testing.T) {
	portPath := filepath.Join(t.TempDir(), "instance.port")
	shown := make(chan struct{}, 1)

	handover := NewHandover(portPath)
	handover.Serve(func() { shown <- struct{}{} })
	defer handover.Close()

	if !Notify(portPath) {
		t.Fatal("the running copy did not take the request")
	}

	select {
	case <-shown:
	case <-time.After(2 * time.Second):
		t.Fatal("show was never delivered")
	}
}

func TestNotifyWithoutARunningCopy(t *testing.T) {
	if Notify(filepath.Join(t.TempDir(), "absent.port")) {
		t.Fatal("reported a handover with nothing listening")
	}
}

func TestCloseRemovesThePortFile(t *testing.T) {
	portPath := filepath.Join(t.TempDir(), "instance.port")

	handover := NewHandover(portPath)
	handover.Serve(func() {})
	handover.Close()

	// A port left behind would send the next launch at whatever takes that port next.
	if Notify(portPath) {
		t.Fatal("the port file outlived the listener")
	}
}
