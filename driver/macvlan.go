package driver

import (
	"fmt"

	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/unix"

	"github.com/mdlayher/netlink"
)

// MacvlanMode represents the MACVLAN operating mode.
type MacvlanMode uint32

// MACVLAN modes.
const (
	MacvlanModePrivate  MacvlanMode = 0x1
	MacvlanModeVEPA     MacvlanMode = 0x2
	MacvlanModeBridge   MacvlanMode = 0x4
	MacvlanModePassthru MacvlanMode = 0x8
	MacvlanModeSource   MacvlanMode = 0x10
)

// String returns a string representation of the MacvlanMode.
func (m MacvlanMode) String() string {
	switch m {
	case MacvlanModePrivate:
		return "private"
	case MacvlanModeVEPA:
		return "vepa"
	case MacvlanModeBridge:
		return "bridge"
	case MacvlanModePassthru:
		return "passthru"
	case MacvlanModeSource:
		return "source"
	default:
		return fmt.Sprintf("unknown MacvlanMode value (%d)", m)
	}
}

// MacvlanFlag represents MACVLAN flags.
type MacvlanFlag uint16

// MACVLAN flags.
const (
	MacvlanFlagNopromisc MacvlanFlag = 0x1
	MacvlanFlagNodst     MacvlanFlag = 0x2
)

// MacvlanMacaddrMode represents the MACVLAN MAC address mode.
type MacvlanMacaddrMode uint32

// MACVLAN MAC address modes.
const (
	MacvlanMacaddrAdd   MacvlanMacaddrMode = 0x0
	MacvlanMacaddrDel   MacvlanMacaddrMode = 0x1
	MacvlanMacaddrFlush MacvlanMacaddrMode = 0x2
	MacvlanMacaddrSet   MacvlanMacaddrMode = 0x3
)

// Macvlan represents a MACVLAN device configuration.
type Macvlan struct {
	// Mode specifies the MACVLAN mode (private, vepa, bridge, passthru, source).
	Mode *MacvlanMode

	// Flags specifies MACVLAN flags (nopromisc, nodst).
	Flags *MacvlanFlag

	// MacaddrMode specifies the MAC address mode for source mode.
	MacaddrMode *MacvlanMacaddrMode

	// MacaddrData contains MAC addresses for source mode.
	MacaddrData [][]byte

	// MacaddrCount specifies the number of MAC addresses in source mode.
	MacaddrCount *uint32

	// BcQueueLen specifies the broadcast queue length.
	BcQueueLen *uint32

	// BcQueueLenUsed indicates if the broadcast queue length is being used.
	BcQueueLenUsed *uint8

	// BcCutoff specifies the broadcast cutoff value.
	BcCutoff *int32
}

var _ rtnetlink.LinkDriver = &Macvlan{}

// New creates a new Macvlan instance.
func (m *Macvlan) New() rtnetlink.LinkDriver {
	return &Macvlan{}
}

// Kind returns the MACVLAN interface kind.
func (m *Macvlan) Kind() string {
	return "macvlan"
}

// Encode encodes the MACVLAN configuration into netlink attributes.
func (m *Macvlan) Encode(ae *netlink.AttributeEncoder) error {

	if m.Mode != nil {
		ae.Uint32(unix.IFLA_MACVLAN_MODE, uint32(*m.Mode))
	}

	if m.Flags != nil {
		ae.Uint16(unix.IFLA_MACVLAN_FLAGS, uint16(*m.Flags))
	}

	if m.MacaddrMode != nil {
		ae.Uint32(unix.IFLA_MACVLAN_MACADDR_MODE, uint32(*m.MacaddrMode))
	}

	if len(m.MacaddrData) > 0 {
		ae.Nested(unix.IFLA_MACVLAN_MACADDR_DATA, func(nae *netlink.AttributeEncoder) error {
			for i, mac := range m.MacaddrData {
				nae.Bytes(uint16(i), mac)
			}
			return nil
		})
	}

	if m.MacaddrCount != nil {
		ae.Uint32(unix.IFLA_MACVLAN_MACADDR_COUNT, *m.MacaddrCount)
	}

	if m.BcQueueLen != nil {
		ae.Uint32(unix.IFLA_MACVLAN_BC_QUEUE_LEN, *m.BcQueueLen)
	}

	if m.BcQueueLenUsed != nil {
		ae.Uint8(unix.IFLA_MACVLAN_BC_QUEUE_LEN_USED, *m.BcQueueLenUsed)
	}

	if m.BcCutoff != nil {
		ae.Int32(unix.IFLA_MACVLAN_BC_CUTOFF, *m.BcCutoff)
	}

	return nil
}

// Decode decodes netlink attributes into the MACVLAN configuration.
func (m *Macvlan) Decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.IFLA_MACVLAN_MODE:
			mode := MacvlanMode(ad.Uint32())
			m.Mode = &mode
		case unix.IFLA_MACVLAN_FLAGS:
			flags := MacvlanFlag(ad.Uint16())
			m.Flags = &flags
		case unix.IFLA_MACVLAN_MACADDR_MODE:
			macaddrMode := MacvlanMacaddrMode(ad.Uint32())
			m.MacaddrMode = &macaddrMode
		case unix.IFLA_MACVLAN_MACADDR_DATA:
			ad.Nested(func(nad *netlink.AttributeDecoder) error {
				for nad.Next() {
					m.MacaddrData = append(m.MacaddrData, nad.Bytes())
				}
				return nad.Err()
			})
		case unix.IFLA_MACVLAN_MACADDR_COUNT:
			count := ad.Uint32()
			m.MacaddrCount = &count
		case unix.IFLA_MACVLAN_BC_QUEUE_LEN:
			qlen := ad.Uint32()
			m.BcQueueLen = &qlen
		case unix.IFLA_MACVLAN_BC_QUEUE_LEN_USED:
			used := ad.Uint8()
			m.BcQueueLenUsed = &used
		case unix.IFLA_MACVLAN_BC_CUTOFF:
			cutoff := ad.Int32()
			m.BcCutoff = &cutoff
		}
	}

	return ad.Err()
}
