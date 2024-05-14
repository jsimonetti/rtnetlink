package testutils

import (
	"bytes"
	"fmt"
	"testing"

	"golang.org/x/sys/unix"
)

func getKernelVersion(tb testing.TB) (maj, min, patch uint32) {
	tb.Helper()

	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		tb.Fatalf("getting uname: %s", err)
	}

	end := bytes.IndexByte(uname.Release[:], 0)
	versionStr := uname.Release[:end]

	if count, _ := fmt.Sscanf(string(versionStr), "%d.%d.%d", &maj, &min, &patch); count < 2 {
		tb.Fatalf("failed to parse kernel version from %s", string(versionStr))
	}
	return
}

// SkipOnOldKernel skips the test if the host's kernel is lower than the given
// x.y target version.
func SkipOnOldKernel(tb testing.TB, target, reason string) {
	maj, min, _ := getKernelVersion(tb)

	var maj_t, min_t, patch_t uint32
	if count, _ := fmt.Sscanf(target, "%d.%d.%d", &maj_t, &min_t, &patch_t); count < 2 {
		tb.Fatalf("failed to parse target version from %s", target)
	}

	if maj < maj_t || maj == maj_t && min < min_t {
		tb.Skipf("host kernel (%d.%d) too old (minimum %d.%d): %s", maj, min, maj_t, min_t, reason)
	}
}
