package driver

import (
	"errors"
	"fmt"

	"github.com/jsimonetti/rtnetlink"
	"github.com/jsimonetti/rtnetlink/internal/unix"
	"github.com/mdlayher/netlink"
)

// NetkitMode specifies netkit operation mode
type NetkitMode uint32

func (n NetkitMode) String() string {
	switch n {
	case NetkitModeL2:
		return "layer2"
	case NetkitModeL3:
		return "layer3"
	default:
		return fmt.Sprintf("unknown NetkitMode value (%d)", n)
	}
}

const (
	// Netkit operates on layer2
	NetkitModeL2 NetkitMode = unix.NETKIT_L2

	// Netkit operates on layer3, this is the default mode
	NetkitModeL3 NetkitMode = unix.NETKIT_L3
)

// NetkitPolicy specifies default packet policy when no eBPF program is attached
type NetkitPolicy int32

func (n NetkitPolicy) String() string {
	switch n {
	case NetkitPolicyPass:
		return "forward"
	case NetkitPolicyDrop:
		return "blackhole"
	default:
		return fmt.Sprintf("unknown NetkitPolicy value (%d)", n)
	}
}

const (
	// Default policy to forwards packets
	NetkitPolicyPass NetkitPolicy = unix.NETKIT_PASS

	// Default policy to drops packets
	NetkitPolicyDrop NetkitPolicy = unix.NETKIT_DROP
)

// Netkit implements LinkDriverVerifier for the netkit driver
type Netkit struct {
	Mode       *NetkitMode            // Specifies driver operation mode
	Policy     *NetkitPolicy          // Specifies default policy
	PeerPolicy *NetkitPolicy          // Specifies default peer policy
	Primary    bool                   // Shows primary link
	PeerInfo   *rtnetlink.LinkMessage // Specifies peer link information
}

var _ rtnetlink.LinkDriverVerifier = &Netkit{}

func (n *Netkit) New() rtnetlink.LinkDriver {
	return &Netkit{}
}

func (n *Netkit) Verify(msg *rtnetlink.LinkMessage) error {
	if msg.Attributes.Address != nil || (n.PeerInfo != nil && n.PeerInfo.Attributes != nil && n.PeerInfo.Attributes.Address != nil) {
		return errors.New("netkit does not support setting Ethernet address")
	}
	return nil
}

func (n *Netkit) Decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.IFLA_NETKIT_MODE:
			v := NetkitMode(ad.Uint32())
			n.Mode = &v
		case unix.IFLA_NETKIT_POLICY:
			v := NetkitPolicy(ad.Int32())
			n.Policy = &v
		case unix.IFLA_NETKIT_PEER_POLICY:
			v := NetkitPolicy(ad.Int32())
			n.PeerPolicy = &v
		case unix.IFLA_NETKIT_PRIMARY:
			n.Primary = ad.Uint8() != 0
		}
	}
	return nil
}

func (n *Netkit) Encode(ae *netlink.AttributeEncoder) error {
	if n.Mode != nil {
		ae.Uint32(unix.IFLA_NETKIT_MODE, uint32(*n.Mode))
	}
	if n.Policy != nil {
		ae.Int32(unix.IFLA_NETKIT_POLICY, int32(*n.Policy))
	}
	if n.PeerPolicy != nil {
		ae.Int32(unix.IFLA_NETKIT_PEER_POLICY, int32(*n.PeerPolicy))
	}
	if n.PeerInfo != nil {
		b, err := n.PeerInfo.MarshalBinary()
		if err != nil {
			return err
		}
		ae.Bytes(unix.IFLA_NETKIT_PEER_INFO, b)
	}
	return nil
}

func (n *Netkit) Kind() string {
	return "netkit"
}
