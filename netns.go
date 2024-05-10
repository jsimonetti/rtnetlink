package rtnetlink

import (
	"os"
	"path/filepath"

	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
)

// NetNS represents a Linux network namespace
type NetNS struct {
	file *os.File
	pid  uint32
}

// NewNetNS returns a new NetNS from the given type
// When an uint32 is given simply the pid value is set
// When a string is given a namespace file is opened with the name and the file descriptor is set
// The file descriptor should be closed after use with the Close() method
func NewNetNS[T string | uint32](t T) (*NetNS, error) {
	if name, ok := any(t).(string); ok {
		file, err := os.Open(filepath.Join("/var/run/netns", name))
		if err != nil {
			return nil, err
		}

		return &NetNS{file: file}, nil
	}
	return &NetNS{pid: any(t).(uint32)}, nil
}

// Type returns either unix.IFLA_NET_NS_FD or unix.IFLA_NET_NS_PID according ns data type
func (n *NetNS) Type() uint16 {
	if n.file != nil {
		return unix.IFLA_NET_NS_FD
	}
	return unix.IFLA_NET_NS_PID
}

// Value returns either a file descriptor value or the pid value of the ns
func (n *NetNS) Value() uint32 {
	if n.file != nil {
		return uint32(n.file.Fd())
	}
	return n.pid
}

// Close closes the file descriptor
func (n *NetNS) Close() error {
	if n.file != nil {
		return n.file.Close()
	}
	return nil
}
