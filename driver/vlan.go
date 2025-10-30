package driver

import (
	"fmt"

	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/unix"

	"github.com/mdlayher/netlink"
)

// VlanProtocol represents the VLAN protocol type.
type VlanProtocol uint16

// VLAN protocols.
const (
	// VlanProtocol8021Q represents 802.1Q VLAN tagging (standard VLAN).
	VlanProtocol8021Q VlanProtocol = 0x0081

	// VlanProtocol8021AD represents 802.1ad QinQ (VLAN stacking).
	VlanProtocol8021AD VlanProtocol = 0xA888
)

// String returns a string representation of the VlanProtocol.
func (p VlanProtocol) String() string {
	switch p {
	case VlanProtocol8021Q:
		return "802.1Q"
	case VlanProtocol8021AD:
		return "802.1ad"
	default:
		return fmt.Sprintf("unknown VlanProtocol value (0x%x)", uint16(p))
	}
}

// VlanFlag represents VLAN flags.
type VlanFlag uint32

// VLAN flags.
const (
	VlanFlagReorderHdr    VlanFlag = 0x1
	VlanFlagGVRP          VlanFlag = 0x2
	VlanFlagLooseBinding  VlanFlag = 0x4
	VlanFlagMVRP          VlanFlag = 0x8
	VlanFlagBridgeBinding VlanFlag = 0x10
)

// VlanQosMapping represents a QoS priority mapping.
type VlanQosMapping struct {
	From uint32
	To   uint32
}

// Vlan represents a VLAN device configuration.
type Vlan struct {
	// ID specifies the VLAN ID (1-4094).
	ID *uint16

	// Protocol specifies the VLAN protocol (802.1Q or 802.1ad).
	Protocol *VlanProtocol

	// Flags specifies VLAN flags.
	Flags *VlanFlag

	// EgressQos specifies egress QoS mappings.
	EgressQos []VlanQosMapping

	// IngressQos specifies ingress QoS mappings.
	IngressQos []VlanQosMapping
}

var _ rtnetlink.LinkDriver = &Vlan{}

// New creates a new Vlan instance.
func (v *Vlan) New() rtnetlink.LinkDriver {
	return &Vlan{}
}

// Kind returns the VLAN interface kind.
func (v *Vlan) Kind() string {
	return "vlan"
}

// Encode encodes a Vlan into netlink attributes.
func (v *Vlan) Encode(ae *netlink.AttributeEncoder) error {
	if v.ID != nil {
		ae.Uint16(unix.IFLA_VLAN_ID, *v.ID)
	}

	if v.Protocol != nil {
		ae.Uint16(unix.IFLA_VLAN_PROTOCOL, uint16(*v.Protocol))
	}

	if v.Flags != nil {
		// VLAN flags are encoded as a nested attribute with flags and mask
		ae.Nested(unix.IFLA_VLAN_FLAGS, func(nae *netlink.AttributeEncoder) error {
			nae.Uint32(1, uint32(*v.Flags)) // flags
			nae.Uint32(2, uint32(*v.Flags)) // mask (same as flags)
			return nil
		})
	}

	if len(v.EgressQos) > 0 {
		ae.Nested(unix.IFLA_VLAN_EGRESS_QOS, func(nae *netlink.AttributeEncoder) error {
			for _, mapping := range v.EgressQos {
				nae.Nested(unix.IFLA_VLAN_QOS_MAPPING, func(mae *netlink.AttributeEncoder) error {
					mae.Uint32(1, mapping.From)
					mae.Uint32(2, mapping.To)
					return nil
				})
			}
			return nil
		})
	}

	if len(v.IngressQos) > 0 {
		ae.Nested(unix.IFLA_VLAN_INGRESS_QOS, func(nae *netlink.AttributeEncoder) error {
			for _, mapping := range v.IngressQos {
				nae.Nested(unix.IFLA_VLAN_QOS_MAPPING, func(mae *netlink.AttributeEncoder) error {
					mae.Uint32(1, mapping.From)
					mae.Uint32(2, mapping.To)
					return nil
				})
			}
			return nil
		})
	}

	return nil
}

// Decode decodes netlink attributes into a Vlan.
func (v *Vlan) Decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.IFLA_VLAN_ID:
			id := ad.Uint16()
			v.ID = &id

		case unix.IFLA_VLAN_PROTOCOL:
			protocol := VlanProtocol(ad.Uint16())
			v.Protocol = &protocol

		case unix.IFLA_VLAN_FLAGS:
			ad.Nested(func(nad *netlink.AttributeDecoder) error {
				for nad.Next() {
					switch nad.Type() {
					case 1: // flags
						flags := VlanFlag(nad.Uint32())
						v.Flags = &flags
					case 2: // mask (we don't need to store this)
						nad.Uint32()
					}
				}
				return nil
			})

		case unix.IFLA_VLAN_EGRESS_QOS:
			ad.Nested(func(nad *netlink.AttributeDecoder) error {
				for nad.Next() {
					if nad.Type() == unix.IFLA_VLAN_QOS_MAPPING {
						var mapping VlanQosMapping
						nad.Nested(func(mad *netlink.AttributeDecoder) error {
							for mad.Next() {
								switch mad.Type() {
								case 1: // from
									mapping.From = mad.Uint32()
								case 2: // to
									mapping.To = mad.Uint32()
								}
							}
							return nil
						})
						v.EgressQos = append(v.EgressQos, mapping)
					}
				}
				return nil
			})

		case unix.IFLA_VLAN_INGRESS_QOS:
			ad.Nested(func(nad *netlink.AttributeDecoder) error {
				for nad.Next() {
					if nad.Type() == unix.IFLA_VLAN_QOS_MAPPING {
						var mapping VlanQosMapping
						nad.Nested(func(mad *netlink.AttributeDecoder) error {
							for mad.Next() {
								switch mad.Type() {
								case 1: // from
									mapping.From = mad.Uint32()
								case 2: // to
									mapping.To = mad.Uint32()
								}
							}
							return nil
						})
						v.IngressQos = append(v.IngressQos, mapping)
					}
				}
				return nil
			})
		}
	}

	return ad.Err()
}
