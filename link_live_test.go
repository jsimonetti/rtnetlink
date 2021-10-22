//go:build integration
// +build integration

package rtnetlink

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"golang.org/x/sys/unix"
)

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

// SetupDummyInterface create a dummy interface for testing and returns its
// properties
func SetupDummyInterface(conn *Conn, name string) (*LinkMessage, error) {
	// construct dummy interface to test XDP program against
	if err := conn.Link.New(&LinkMessage{
		Family: unix.AF_UNSPEC,
		Index:  1001,
		Flags:  unix.IFF_UP,
		Attributes: &LinkAttributes{
			Name: name,
			Info: &LinkInfo{Kind: "dummy"},
		},
	}); err != nil {
		return nil, err
	}

	// get info for the dummy interface
	interf, err := conn.Link.Get(1001)
	if err != nil {
		conn.Link.Delete(interf.Index)
		return nil, err
	}
	return &interf, err
}

func GetBPFPrograms() (int32, int32, error) {
	// load a BPF test program. If it fails error out of the tests
	// and clean up dummy interface. The program loads XDP_PASS
	// into the return value register.
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
		return 0, 0, err
	}

	prog2, err := ebpf.NewProgram(bpfProgram)
	if err != nil {
		return 0, 0, err
	}

	// Use the file descriptor of the programs
	return int32(prog1.FD()), int32(prog2.FD()), nil
}

// SendXDPMsg sends a XDP netlink msg with the specified LinkXDP properties
func SendXPDMsg(conn *Conn, ifIndex uint32, xdp *LinkXDP) error {
	message := LinkMessage{
		Family: unix.AF_UNSPEC,
		Index:  ifIndex,
		Attributes: &LinkAttributes{
			XDP: xdp,
		},
	}

	return conn.Link.Set(&message)
}

// GetXDPProperties returns the XDP attach, XDP prog ID and errors when the
// interface could not be fetched
func GetXDPProperties(conn *Conn, ifIndex uint32) (uint8, uint32, error) {
	interf, err := conn.Link.Get(ifIndex)
	if err != nil {
		return 0, 0, err
	}
	return interf.Attributes.XDP.Attached, interf.Attributes.XDP.ProgID, nil
}

func TestLinkXDPAttach(t *testing.T) {
	// BPF loading requires a high RLIMIT_MEMLOCK.
	n := uint64(1024 * 1024 * 10)
	err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{Cur: n, Max: n})
	if err != nil {
		t.Fatalf("failed to increase RLIMIT_MEMLOCK: %v", err)
	}

	// establish a netlink connections
	conn, err := Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	// setup dummy interface for the test
	interf, err := SetupDummyInterface(conn, "dummyXDPAttach")
	if err != nil {
		t.Fatalf("failed to setup dummy interface: %v", err)
	}
	defer conn.Link.Delete(interf.Index)

	// get a BPF program
	progFD1, progFD2, err := GetBPFPrograms()
	if err != nil {
		t.Fatalf("failed to load bpf programs: %v", err)
	}

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
			// attach the BPF program to the link
			err = SendXPDMsg(conn, interf.Index, tt.xdp)
			if err != nil {
				t.Fatalf("failed to attach XDP program to link: %v", err)
			}

			// validate the XDP properites of the link
			attached, progID, err := GetXDPProperties(conn, interf.Index)
			if err != nil {
				t.Fatalf("failed to get XDP properties from the link: %v", err)
			}

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
	// BPF loading requires a high RLIMIT_MEMLOCK.
	n := uint64(1024 * 1024 * 10)
	err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{Cur: n, Max: n})
	if err != nil {
		t.Fatalf("failed to increase RLIMIT_MEMLOCK: %v", err)
	}

	// establish a netlink connections
	conn, err := Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	// setup dummy interface for the test
	interf, err := SetupDummyInterface(conn, "dummyXDPClear")
	if err != nil {
		t.Fatalf("failed to setup dummy interface: %v", err)
	}
	defer conn.Link.Delete(interf.Index)

	// get a BPF program
	progFD1, _, err := GetBPFPrograms()
	if err != nil {
		t.Fatalf("failed to load bpf programs: %v", err)
	}

	// attach the BPF program to the link
	err = SendXPDMsg(conn, interf.Index, &LinkXDP{
		FD:    progFD1,
		Flags: unix.XDP_FLAGS_SKB_MODE,
	})
	if err != nil {
		t.Fatalf("failed to attach XDP program to link: %v", err)
	}

	// clear the BPF program from the link
	err = SendXPDMsg(conn, interf.Index, &LinkXDP{
		FD:    -1,
		Flags: unix.XDP_FLAGS_SKB_MODE,
	})
	if err != nil {
		t.Fatalf("failed to clear XDP program to link: %v", err)
	}

	attached, progID, err := GetXDPProperties(conn, interf.Index)
	if err != nil {
		t.Fatalf("failed to get XDP program ID 1 from interface: %v", err)
	}

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

	// BPF loading requires a high RLIMIT_MEMLOCK.
	n := uint64(1024 * 1024 * 10)
	err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{Cur: n, Max: n})
	if err != nil {
		t.Fatalf("failed to increase RLIMIT_MEMLOCK: %v", err)
	}

	// establish a netlink connections
	conn, err := Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	// setup dummy interface for the test
	interf, err := SetupDummyInterface(conn, "dummyXDPReplace")
	if err != nil {
		t.Fatalf("failed to setup dummy interface: %v", err)
	}
	defer conn.Link.Delete(interf.Index)

	// get BPF programs
	progFD1, progFD2, err := GetBPFPrograms()
	if err != nil {
		t.Fatalf("failed to load bpf programs: %v", err)
	}

	err = SendXPDMsg(conn, interf.Index, &LinkXDP{
		FD:    progFD1,
		Flags: unix.XDP_FLAGS_SKB_MODE,
	})
	if err != nil {
		t.Fatalf("failed to attach XDP program 1 to link: %v", err)
	}

	_, progID1, err := GetXDPProperties(conn, interf.Index)
	if err != nil {
		t.Fatalf("failed to get XDP program ID 1 from interface: %v", err)
	}

	err = SendXPDMsg(conn, interf.Index, &LinkXDP{
		FD:         progFD2,
		ExpectedFD: progFD2,
		Flags:      unix.XDP_FLAGS_SKB_MODE | unix.XDP_FLAGS_REPLACE,
	})
	if err == nil {
		t.Fatalf("replaced XDP program while expected FD did not match: %v", err)
	}

	_, progID2, err := GetXDPProperties(conn, interf.Index)
	if err != nil {
		t.Fatalf("failed to get XDP program ID 2 from interface: %v", err)
	}
	if progID2 != progID1 {
		t.Fatal("XDP prog ID does not match previous program ID, which it should")
	}

	err = SendXPDMsg(conn, interf.Index, &LinkXDP{
		FD:         progFD2,
		ExpectedFD: progFD1,
		Flags:      unix.XDP_FLAGS_SKB_MODE | unix.XDP_FLAGS_REPLACE,
	})
	if err != nil {
		t.Fatalf("could not replace XDP program: %v", err)
	}

	_, progID2, err = GetXDPProperties(conn, interf.Index)
	if err != nil {
		t.Fatalf("failed to get XDP program ID 2 from interface: %v", err)
	}
	if progID2 == progID1 {
		t.Fatal("XDP prog ID does match previous program ID, which it shouldn't")
	}
}
