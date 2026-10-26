package driver

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"github.com/mdlayher/netlink"
)

func TestVlanProtocolNetworkOrder(t *testing.T) {
	for _, protocol := range []struct {
		name  string
		value VlanProtocol
		wire  []byte
	}{
		{"802.1Q", VlanProtocol8021Q, []byte{0x81, 0x00}},
		{"802.1ad", VlanProtocol8021AD, []byte{0x88, 0xa8}},
	} {
		for _, order := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
			for _, driver := range []struct {
				name         string
				new          func() rtnetlink.LinkDriver
				configured   rtnetlink.LinkDriver
				protocolType uint16
				nativeType   uint16
			}{
				{"vlan", func() rtnetlink.LinkDriver { return &Vlan{} }, &Vlan{Protocol: &protocol.value}, unix.IFLA_VLAN_PROTOCOL, unix.IFLA_VLAN_ID},
				{"bridge", func() rtnetlink.LinkDriver { return &Bridge{} }, &Bridge{VlanProtocol: &protocol.value}, unix.IFLA_BR_VLAN_PROTOCOL, unix.IFLA_BR_GROUP_FWD_MASK},
			} {
				t.Run(protocol.name+"/"+order.String()+"/"+driver.name, func(t *testing.T) {
					ae := netlink.NewAttributeEncoder()
					ae.ByteOrder = order
					if err := driver.configured.Encode(ae); err != nil {
						t.Fatal(err)
					}
					// Protocol byte order must not change later native-endian attributes.
					ae.Uint16(driver.nativeType, 0x1234)
					b, err := ae.Encode()
					if err != nil {
						t.Fatal(err)
					}
					attrs, err := netlink.UnmarshalAttributes(b)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(attrs[0].Data, protocol.wire) {
						t.Errorf("protocol payload: got %x, want %x", attrs[0].Data, protocol.wire)
					}
					if got := order.Uint16(attrs[1].Data); got != 0x1234 {
						t.Errorf("native attribute: got %x, want 1234", got)
					}

					// Decode an independent kernel-format fixture, not the encoder output.
					ae = netlink.NewAttributeEncoder()
					ae.ByteOrder = order
					ae.Bytes(driver.protocolType, protocol.wire)
					ae.Uint16(driver.nativeType, 0x1234)
					b, err = ae.Encode()
					if err != nil {
						t.Fatal(err)
					}
					ad, err := netlink.NewAttributeDecoder(b)
					if err != nil {
						t.Fatal(err)
					}
					ad.ByteOrder = order
					decoded := driver.new()
					if err := decoded.Decode(ad); err != nil {
						t.Fatal(err)
					}
					var gotProtocol *VlanProtocol
					var gotNative *uint16
					switch d := decoded.(type) {
					case *Vlan:
						gotProtocol, gotNative = d.Protocol, d.ID
					case *Bridge:
						gotProtocol, gotNative = d.VlanProtocol, d.GroupFwdMask
					}
					if gotProtocol == nil || *gotProtocol != protocol.value {
						t.Errorf("decoded protocol: got %v, want %v", gotProtocol, protocol.value)
					}
					if gotNative == nil || *gotNative != 0x1234 {
						t.Errorf("decoded native attribute: got %v, want 1234", gotNative)
					}
				})
			}
		}
	}
}
