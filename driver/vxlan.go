package driver

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"github.com/mdlayher/netlink"
)

// VxlanDFMode specifies how to handle DF flag in outer IPv4 header
type VxlanDFMode uint8

const (
	// VxlanDFUnset indicates DF flag is not set (default)
	VxlanDFUnset VxlanDFMode = iota

	// VxlanDFSet indicates DF flag is set
	VxlanDFSet

	// VxlanDFInherit indicates DF flag is inherited from inner IPv4 header
	VxlanDFInherit
)

func (v VxlanDFMode) String() string {
	switch v {
	case VxlanDFUnset:
		return "unset"
	case VxlanDFSet:
		return "set"
	case VxlanDFInherit:
		return "inherit"
	default:
		return fmt.Sprintf("unknown VxlanDFMode value (%d)", v)
	}
}

// VxlanPortRange specifies the range of source UDP ports
type VxlanPortRange struct {
	Low  uint16
	High uint16
}

// Vxlan implements LinkDriver for the vxlan driver
type Vxlan struct {
	// VXLAN Network Identifier (or VXLAN Segment ID) - required
	ID *uint32

	// Multicast group IP address to join (IPv4)
	Group net.IP

	// Multicast group IP address to join (IPv6)
	Group6 net.IP

	// Physical device to use for tunnel endpoint communication
	Link *uint32

	// Source IP address to use in outgoing packets (IPv4)
	Local net.IP

	// Source IP address to use in outgoing packets (IPv6)
	Local6 net.IP

	// TTL to use in outgoing packets
	TTL *uint8

	// TOS to use in outgoing packets
	TOS *uint8

	// Enable learning of source link addresses
	Learning *bool

	// Lifetime in seconds of FDB entries learnt by the kernel
	Ageing *uint32

	// Maximum number of FDB entries
	Limit *uint32

	// Range of source UDP ports to use
	PortRange *VxlanPortRange

	// Enable ARP proxy
	Proxy *bool

	// Enable route short circuit
	RSC *bool

	// Enable netlink LLADDR miss notifications
	L2Miss *bool

	// Enable netlink IP address miss notifications
	L3Miss *bool

	// Destination port for VXLAN traffic (default 4789)
	Port *uint16

	// Enable UDP checksums on transmit for outer IPv4
	UDPCsum *bool

	// Enable zero UDP checksums on transmit for outer IPv6
	UDPZeroCsum6Tx *bool

	// Allow zero UDP checksums on receive for outer IPv6
	UDPZeroCsum6Rx *bool

	// Enable remote checksum offload on transmit
	RemCsumTx *bool

	// Enable remote checksum offload on receive
	RemCsumRx *bool

	// Enable VXLAN Group Policy extension
	GBP *bool

	// Enable remote checksum offload without partial checksums
	RemCsumNoPartial *bool

	// Enable metadata collection mode
	CollectMetadata *bool

	// MPLS label to use for VXLAN encapsulation
	Label *uint32

	// Enable Generic Protocol Extension (VXLAN-GPE)
	GPE *bool

	// Inherit TTL from inner packet
	TTLInherit *bool

	// Specifies how to handle DF flag in outer IPv4 header
	DF *VxlanDFMode

	// Enable VXLAN VNI filtering
	VNIFilter *bool
}

var _ rtnetlink.LinkDriver = &Vxlan{}

func (v *Vxlan) New() rtnetlink.LinkDriver {
	return &Vxlan{}
}

func (v *Vxlan) Encode(ae *netlink.AttributeEncoder) error {
	if v.ID != nil {
		ae.Uint32(unix.IFLA_VXLAN_ID, *v.ID)
	}
	if v.Group != nil {
		ip := v.Group.To4()
		if ip == nil {
			return fmt.Errorf("group must be an IPv4 address")
		}
		ae.Bytes(unix.IFLA_VXLAN_GROUP, ip)
	}
	if v.Group6 != nil {
		// Check if it's actually an IPv6 address (not an IPv4 address)
		if v.Group6.To4() != nil {
			return fmt.Errorf("group6 must be an IPv6 address")
		}
		ip := v.Group6.To16()
		if ip == nil {
			return fmt.Errorf("group6 must be an IPv6 address")
		}
		ae.Bytes(unix.IFLA_VXLAN_GROUP6, ip)
	}
	if v.Link != nil {
		ae.Uint32(unix.IFLA_VXLAN_LINK, *v.Link)
	}
	if v.Local != nil {
		ip := v.Local.To4()
		if ip == nil {
			return fmt.Errorf("local must be an IPv4 address")
		}
		ae.Bytes(unix.IFLA_VXLAN_LOCAL, ip)
	}
	if v.Local6 != nil {
		// Check if it's actually an IPv6 address (not an IPv4 address)
		if v.Local6.To4() != nil {
			return fmt.Errorf("local6 must be an IPv6 address")
		}
		ip := v.Local6.To16()
		if ip == nil {
			return fmt.Errorf("local6 must be an IPv6 address")
		}
		ae.Bytes(unix.IFLA_VXLAN_LOCAL6, ip)
	}
	if v.TTL != nil {
		ae.Uint8(unix.IFLA_VXLAN_TTL, *v.TTL)
	}
	if v.TOS != nil {
		ae.Uint8(unix.IFLA_VXLAN_TOS, *v.TOS)
	}
	if v.Learning != nil {
		var val uint8
		if *v.Learning {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_LEARNING, val)
	}
	if v.Ageing != nil {
		ae.Uint32(unix.IFLA_VXLAN_AGEING, *v.Ageing)
	}
	if v.Limit != nil {
		ae.Uint32(unix.IFLA_VXLAN_LIMIT, *v.Limit)
	}
	if v.PortRange != nil {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint16(buf[0:2], v.PortRange.Low)
		binary.BigEndian.PutUint16(buf[2:4], v.PortRange.High)
		ae.Bytes(unix.IFLA_VXLAN_PORT_RANGE, buf)
	}
	if v.Proxy != nil {
		var val uint8
		if *v.Proxy {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_PROXY, val)
	}
	if v.RSC != nil {
		var val uint8
		if *v.RSC {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_RSC, val)
	}
	if v.L2Miss != nil {
		var val uint8
		if *v.L2Miss {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_L2MISS, val)
	}
	if v.L3Miss != nil {
		var val uint8
		if *v.L3Miss {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_L3MISS, val)
	}
	if v.Port != nil {
		// Port must be in network byte order (big-endian)
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, *v.Port)
		ae.Bytes(unix.IFLA_VXLAN_PORT, buf)
	}
	if v.UDPCsum != nil {
		var val uint8
		if *v.UDPCsum {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_UDP_CSUM, val)
	}
	if v.UDPZeroCsum6Tx != nil {
		var val uint8
		if *v.UDPZeroCsum6Tx {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_UDP_ZERO_CSUM6_TX, val)
	}
	if v.UDPZeroCsum6Rx != nil {
		var val uint8
		if *v.UDPZeroCsum6Rx {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_UDP_ZERO_CSUM6_RX, val)
	}
	if v.RemCsumTx != nil {
		var val uint8
		if *v.RemCsumTx {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_REMCSUM_TX, val)
	}
	if v.RemCsumRx != nil {
		var val uint8
		if *v.RemCsumRx {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_REMCSUM_RX, val)
	}
	if v.GBP != nil {
		if *v.GBP {
			ae.Uint32(unix.IFLA_VXLAN_GBP, 0)
		}
	}
	if v.RemCsumNoPartial != nil {
		if *v.RemCsumNoPartial {
			ae.Uint32(unix.IFLA_VXLAN_REMCSUM_NOPARTIAL, 0)
		}
	}
	if v.CollectMetadata != nil {
		var val uint8
		if *v.CollectMetadata {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_COLLECT_METADATA, val)
	}
	if v.Label != nil {
		ae.Uint32(unix.IFLA_VXLAN_LABEL, *v.Label)
	}
	if v.GPE != nil {
		if *v.GPE {
			ae.Uint32(unix.IFLA_VXLAN_GPE, 0)
		}
	}
	if v.TTLInherit != nil {
		var val uint8
		if *v.TTLInherit {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_TTL_INHERIT, val)
	}
	if v.DF != nil {
		ae.Uint8(unix.IFLA_VXLAN_DF, uint8(*v.DF))
	}
	if v.VNIFilter != nil {
		var val uint8
		if *v.VNIFilter {
			val = 1
		}
		ae.Uint8(unix.IFLA_VXLAN_VNIFILTER, val)
	}

	return nil
}

func (v *Vxlan) Decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.IFLA_VXLAN_ID:
			val := ad.Uint32()
			v.ID = &val
		case unix.IFLA_VXLAN_GROUP:
			v.Group = net.IP(ad.Bytes())
		case unix.IFLA_VXLAN_GROUP6:
			v.Group6 = net.IP(ad.Bytes())
		case unix.IFLA_VXLAN_LINK:
			val := ad.Uint32()
			v.Link = &val
		case unix.IFLA_VXLAN_LOCAL:
			v.Local = net.IP(ad.Bytes())
		case unix.IFLA_VXLAN_LOCAL6:
			v.Local6 = net.IP(ad.Bytes())
		case unix.IFLA_VXLAN_TTL:
			val := ad.Uint8()
			v.TTL = &val
		case unix.IFLA_VXLAN_TOS:
			val := ad.Uint8()
			v.TOS = &val
		case unix.IFLA_VXLAN_LEARNING:
			val := ad.Uint8() != 0
			v.Learning = &val
		case unix.IFLA_VXLAN_AGEING:
			val := ad.Uint32()
			v.Ageing = &val
		case unix.IFLA_VXLAN_LIMIT:
			val := ad.Uint32()
			v.Limit = &val
		case unix.IFLA_VXLAN_PORT_RANGE:
			buf := ad.Bytes()
			if len(buf) >= 4 {
				v.PortRange = &VxlanPortRange{
					Low:  binary.BigEndian.Uint16(buf[0:2]),
					High: binary.BigEndian.Uint16(buf[2:4]),
				}
			}
		case unix.IFLA_VXLAN_PROXY:
			val := ad.Uint8() != 0
			v.Proxy = &val
		case unix.IFLA_VXLAN_RSC:
			val := ad.Uint8() != 0
			v.RSC = &val
		case unix.IFLA_VXLAN_L2MISS:
			val := ad.Uint8() != 0
			v.L2Miss = &val
		case unix.IFLA_VXLAN_L3MISS:
			val := ad.Uint8() != 0
			v.L3Miss = &val
		case unix.IFLA_VXLAN_PORT:
			// Port is in network byte order (big-endian)
			buf := ad.Bytes()
			if len(buf) >= 2 {
				val := binary.BigEndian.Uint16(buf)
				v.Port = &val
			}
		case unix.IFLA_VXLAN_UDP_CSUM:
			val := ad.Uint8() != 0
			v.UDPCsum = &val
		case unix.IFLA_VXLAN_UDP_ZERO_CSUM6_TX:
			val := ad.Uint8() != 0
			v.UDPZeroCsum6Tx = &val
		case unix.IFLA_VXLAN_UDP_ZERO_CSUM6_RX:
			val := ad.Uint8() != 0
			v.UDPZeroCsum6Rx = &val
		case unix.IFLA_VXLAN_REMCSUM_TX:
			val := ad.Uint8() != 0
			v.RemCsumTx = &val
		case unix.IFLA_VXLAN_REMCSUM_RX:
			val := ad.Uint8() != 0
			v.RemCsumRx = &val
		case unix.IFLA_VXLAN_GBP:
			val := true
			v.GBP = &val
		case unix.IFLA_VXLAN_REMCSUM_NOPARTIAL:
			val := true
			v.RemCsumNoPartial = &val
		case unix.IFLA_VXLAN_COLLECT_METADATA:
			val := ad.Uint8() != 0
			v.CollectMetadata = &val
		case unix.IFLA_VXLAN_LABEL:
			val := ad.Uint32()
			v.Label = &val
		case unix.IFLA_VXLAN_GPE:
			val := true
			v.GPE = &val
		case unix.IFLA_VXLAN_TTL_INHERIT:
			val := ad.Uint8() != 0
			v.TTLInherit = &val
		case unix.IFLA_VXLAN_DF:
			val := VxlanDFMode(ad.Uint8())
			v.DF = &val
		case unix.IFLA_VXLAN_VNIFILTER:
			val := ad.Uint8() != 0
			v.VNIFilter = &val
		}
	}
	return nil
}

func (*Vxlan) Kind() string {
	return "vxlan"
}
