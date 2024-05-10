//go:build integration
// +build integration

package driver

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"

	"github.com/jsimonetti/rtnetlink"
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

// setupInterface create a interface for testing
func setupInterface(conn *rtnetlink.Conn, name string, index, master uint32, driver rtnetlink.LinkDriver) error {
	attrs := &rtnetlink.LinkAttributes{
		Name: name,
		Info: &rtnetlink.LinkInfo{Kind: driver.Kind(), Data: driver},
	}
	flag := uint32(unix.IFF_UP)
	if master > 0 {
		attrs.Master = &master
		flag = 0
	}
	// construct an interface to test drivers
	err := conn.Link.New(&rtnetlink.LinkMessage{
		Family:     unix.AF_UNSPEC,
		Index:      index,
		Flags:      flag,
		Change:     flag,
		Attributes: attrs,
	})
	if err != nil {
		conn.Link.Delete(index)
	}
	return err
}

func getInterface(conn *rtnetlink.Conn, index uint32) (*rtnetlink.LinkMessage, error) {
	interf, err := conn.Link.Get(index)
	if err != nil {
		conn.Link.Delete(interf.Index)
		return nil, err
	}
	return &interf, err
}

// creates a network namespace by utilizing ip commandline tool
// returns NetNS and clean function
func createNS(name string) (*rtnetlink.NetNS, func(), error) {
	cmdPath, err := exec.LookPath("ip")
	if err != nil {
		return nil, nil, fmt.Errorf("getting ip command path failed, %w", err)
	}
	_, err = exec.Command(cmdPath, "netns", "add", name).Output()
	if err != nil {
		return nil, nil, fmt.Errorf("ip netns add %s, failed: %w", name, err)
	}

	ns, err := rtnetlink.NewNetNS(name)
	if err != nil {
		return nil, nil, fmt.Errorf("reading ns %s, failed: %w", name, err)
	}
	return ns, func() {
		if err := ns.Close(); err != nil {
			fmt.Printf("closing ns file failed: %v", err)
		}
		_, err := exec.Command(cmdPath, "netns", "del", name).Output()
		if err != nil {
			fmt.Printf("removing netns %s failed, %v", name, err)
		}
	}, nil
}
