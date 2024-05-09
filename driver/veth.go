package driver

import (
	"fmt"

	"github.com/jsimonetti/rtnetlink"
	"github.com/mdlayher/netlink"
)

const veth_info_peer = 0x1

// Veth implements LinkDriverVerifier for the veth driver
type Veth struct {
	PeerInfo *rtnetlink.LinkMessage // Specifies peer link information
}

var _ rtnetlink.LinkDriverVerifier = &Veth{}

func (v *Veth) New() rtnetlink.LinkDriver {
	return &Veth{}
}

func (v *Veth) Encode(ae *netlink.AttributeEncoder) error {
	b, err := v.PeerInfo.MarshalBinary()
	if err != nil {
		return err
	}
	ae.Bytes(veth_info_peer, b)

	return nil
}

func (v *Veth) Decode(ad *netlink.AttributeDecoder) error {
	return nil
}

func (*Veth) Kind() string {
	return "veth"
}

const (
	eth_min_mtu = 68    // Min IPv4 MTU per RFC791
	eth_max_mtu = 65535 // 65535, same as IP_MAX_MTU
)

func (v *Veth) Verify(msg *rtnetlink.LinkMessage) error {
	if msg.Attributes != nil && msg.Attributes.MTU > 0 && (msg.Attributes.MTU < eth_min_mtu || msg.Attributes.MTU > eth_max_mtu) {
		return fmt.Errorf("invalid MTU value %d, must be between %d %d", msg.Attributes.MTU, eth_min_mtu, eth_max_mtu)
	}
	return nil
}
