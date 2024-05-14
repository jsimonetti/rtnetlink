package testutils

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"golang.org/x/sync/errgroup"
)

// NetNS returns a file descriptor to a new network namespace.
// The netns handle is automatically closed as part of test cleanup.
func NetNS(tb testing.TB) int {
	tb.Helper()

	var ns *os.File
	var eg errgroup.Group
	eg.Go(func() error {
		// Lock the new goroutine to its OS thread. Never unlock the goroutine so
		// the thread dies when the goroutine ends to avoid having to restore the
		// thread's netns.
		runtime.LockOSThread()

		// Move the current thread to a new network namespace.
		if err := unix.Unshare(unix.CLONE_NEWNET); err != nil {
			return fmt.Errorf("unsharing netns: %w", err)
		}

		f, err := os.OpenFile(fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), unix.Gettid()),
			unix.O_RDONLY|unix.O_CLOEXEC, 0)
		if err != nil {
			return fmt.Errorf("opening netns handle: %w", err)
		}

		// Store a namespace reference in the outer scope.
		ns = f

		return nil
	})

	if err := eg.Wait(); err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		// Maintain a reference to the namespace until the end of the test, where
		// the handle will close automatically and the namespace potentially
		// disappears if there are no other references (veth/netkit peers, ..) to it.
		ns.Close()
	})

	return int(ns.Fd())
}
