//go:build integration
// +build integration

package rtnetlink

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"github.com/cilium/ebpf/rlimit"
	"github.com/jsimonetti/rtnetlink/v2/internal/testutils"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

// lo accesses the loopback interface present in every network namespace.
var lo uint32 = 1

func getKernelVersion() (kernel, major, minor int, err error) {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return 0, 0, 0, err
	}

	end := bytes.IndexByte(uname.Release[:], 0)
	versionStr := uname.Release[:end]

	if count, _ := fmt.Sscanf(string(versionStr), "%d.%d.%d", &kernel, &major, &minor); count < 2 {
		err = fmt.Errorf("failed to parse kernel version from: %q", string(versionStr))
	}
	return
}

// kernelMinReq checks if the runtime kernel is sufficient
// for the test
func kernelMinReq(t *testing.T, kernel, major int) {
	k, m, _, err := getKernelVersion()
	if err != nil {
		t.Fatalf("failed to get host kernel version: %v", err)
	}
	if k < kernel || k == kernel && m < major {
		t.Skipf("host kernel (%d.%d) does not meet test's minimum required version: (%d.%d)",
			k, m, kernel, major)
	}
}

func xdpPrograms(tb testing.TB) (int32, int32) {
	tb.Helper()

	// Load XDP_PASS into the return value register.
	bpfProgram := &ebpf.ProgramSpec{
		Type: ebpf.XDP,
		Instructions: asm.Instructions{
			asm.LoadImm(asm.R0, int64(2), asm.DWord),
			asm.Return(),
		},
		License: "MIT",
	}
	prog1, err := ebpf.NewProgram(bpfProgram)
	if err != nil {
		tb.Fatal(err)
	}

	prog2, err := ebpf.NewProgram(bpfProgram)
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		prog1.Close()
		prog2.Close()
	})

	// Use the file descriptor of the programs
	return int32(prog1.FD()), int32(prog2.FD())
}

func attachXDP(tb testing.TB, conn *Conn, ifIndex uint32, xdp *LinkXDP) {
	tb.Helper()

	message := LinkMessage{
		Family: unix.AF_UNSPEC,
		Index:  ifIndex,
		Attributes: &LinkAttributes{
			XDP: xdp,
		},
	}

	if err := conn.Link.Set(&message); err != nil {
		tb.Fatalf("attaching program with fd %d to link at ifindex %d: %s", xdp.FD, ifIndex, err)
	}
}

// getXDP returns the XDP attach, XDP prog ID and errors when the
// interface could not be fetched
func getXDP(tb testing.TB, conn *Conn, ifIndex uint32) (uint8, uint32) {
	tb.Helper()

	interf, err := conn.Link.Get(ifIndex)
	if err != nil {
		tb.Fatalf("getting link xdp properties: %s", err)
	}

	return interf.Attributes.XDP.Attached, interf.Attributes.XDP.ProgID
}

func TestLinkXDPAttach(t *testing.T) {
	if err := rlimit.RemoveMemlock(); err != nil {
		t.Fatal(err)
	}

	conn, err := Dial(&netlink.Config{NetNS: testutils.NetNS(t)})
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	progFD1, progFD2 := xdpPrograms(t)

	tests := []struct {
		name string
		xdp  *LinkXDP
	}{
		{
			name: "with FD, no expected FD",
			xdp: &LinkXDP{
				FD:    progFD1,
				Flags: unix.XDP_FLAGS_SKB_MODE,
			},
		},
		{
			name: "with FD, expected FD == FD",
			xdp: &LinkXDP{
				FD:         progFD1,
				ExpectedFD: progFD1,
				Flags:      unix.XDP_FLAGS_SKB_MODE,
			},
		},
		{
			name: "with FD, expected FD != FD",
			xdp: &LinkXDP{
				FD:         progFD1,
				ExpectedFD: progFD2,
				Flags:      unix.XDP_FLAGS_SKB_MODE,
			},
		},
		{
			name: "with FD, expected FD < 0",
			xdp: &LinkXDP{
				FD:         progFD1,
				ExpectedFD: -1,
				Flags:      unix.XDP_FLAGS_SKB_MODE,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachXDP(t, conn, lo, tt.xdp)

			attached, progID := getXDP(t, conn, lo)
			if attached != unix.XDP_FLAGS_SKB_MODE {
				t.Fatalf("XDP attached state does not match. Got: %d, wanted: %d", attached, unix.XDP_FLAGS_SKB_MODE)
			}
			if attached == unix.XDP_FLAGS_SKB_MODE && progID == 0 {
				t.Fatalf("XDP program should be attached but program ID is 0")
			}
		})
	}
}

func TestLinkXDPClear(t *testing.T) {
	if err := rlimit.RemoveMemlock(); err != nil {
		t.Fatal(err)
	}

	conn, err := Dial(&netlink.Config{NetNS: testutils.NetNS(t)})
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	progFD1, _ := xdpPrograms(t)

	attachXDP(t, conn, lo, &LinkXDP{
		FD:    progFD1,
		Flags: unix.XDP_FLAGS_SKB_MODE,
	})

	// clear the BPF program from the link
	attachXDP(t, conn, lo, &LinkXDP{
		FD:    -1,
		Flags: unix.XDP_FLAGS_SKB_MODE,
	})

	attached, progID := getXDP(t, conn, lo)
	if progID != 0 {
		t.Fatalf("there is still a program loaded, while we cleared the link")
	}
	if attached != 0 {
		t.Fatalf(
			"XDP attached state does not match. Got: %d, wanted: %d\nThere should be no program loaded",
			attached, 0,
		)
	}
}

func TestLinkXDPReplace(t *testing.T) {
	// As of kernel version 5.7, the use of EXPECTED_FD and XDP_FLAGS_REPLACE
	// is supported. We check here if the test host kernel fills this
	// requirement. If the requirement is not met, we skip this test and
	// output a notice. Running the code on a kernel version lower then 5.7
	// will throw an "invalid argument" error.
	// source kernel 5.6:
	// https://elixir.bootlin.com/linux/v5.6/source/net/core/dev.c#L8662
	// source kernel 5.7:
	// https://elixir.bootlin.com/linux/v5.7/source/net/core/dev.c#L8674
	kernelMinReq(t, 5, 7)

	if err := rlimit.RemoveMemlock(); err != nil {
		t.Fatal(err)
	}

	conn, err := Dial(&netlink.Config{NetNS: testutils.NetNS(t)})
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	progFD1, progFD2 := xdpPrograms(t)

	attachXDP(t, conn, lo, &LinkXDP{
		FD:    progFD1,
		Flags: unix.XDP_FLAGS_SKB_MODE,
	})

	_, progID1 := getXDP(t, conn, lo)

	if err := conn.Link.Set(&LinkMessage{
		Family: unix.AF_UNSPEC,
		Index:  lo,
		Attributes: &LinkAttributes{
			XDP: &LinkXDP{
				FD:         progFD2,
				ExpectedFD: progFD2,
				Flags:      unix.XDP_FLAGS_SKB_MODE | unix.XDP_FLAGS_REPLACE,
			},
		},
	}); err == nil {
		t.Fatalf("replaced XDP program while expected FD did not match: %v", err)
	}

	_, progID2 := getXDP(t, conn, lo)
	if progID2 != progID1 {
		t.Fatal("XDP prog ID does not match previous program ID, which it should")
	}

	attachXDP(t, conn, lo, &LinkXDP{
		FD:         progFD2,
		ExpectedFD: progFD1,
		Flags:      unix.XDP_FLAGS_SKB_MODE | unix.XDP_FLAGS_REPLACE,
	})

	_, progID2 = getXDP(t, conn, lo)
	if progID2 == progID1 {
		t.Fatal("XDP prog ID does match previous program ID, which it shouldn't")
	}
}
