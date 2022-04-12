package rtnetlink

import (
	"bytes"
	"errors"
	"net"
	"reflect"
	"testing"
)

func TestRuleMessage(t *testing.T) {
	skipBigEndian(t)

	tests := map[string]struct {
		m            Message
		b            []byte
		marshalErr   error
		unmarshalErr error
	}{
		"empty": {
			m: &RuleMessage{},
			b: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		"no attributes": {
			m: &RuleMessage{
				Family:    1,
				DstLength: 2,
				SrcLength: 3,
				TOS:       4,
				Table:     5,
				Action:    6,
				Flags:     7,
			},
			b: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x00, 0x00, 0x06, 0x07, 0x00, 0x00, 0x00},
		},
		"with attributes": {
			m: &RuleMessage{
				Family:    7,
				DstLength: 6,
				SrcLength: 5,
				TOS:       4,
				Table:     3,
				Action:    2,
				Flags:     1,
				Attributes: &RuleAttributes{
					Src:               netIPPtr(net.ParseIP("8.8.8.8")),
					Dst:               netIPPtr(net.ParseIP("1.1.1.1")),
					IIFName:           strPtr("eth0"),
					OIFName:           strPtr("br0"),
					Goto:              uint32Ptr(1),
					Priority:          uint32Ptr(2),
					FwMark:            uint32Ptr(3),
					FwMask:            uint32Ptr(5),
					L3MDev:            uint8Ptr(7),
					DstRealm:          uint16Ptr(11),
					SrcRealm:          uint16Ptr(13),
					TunID:             uint64Ptr(17),
					Protocol:          uint8Ptr(19),
					IPProto:           uint8Ptr(23),
					Table:             uint32Ptr(29),
					SuppressPrefixLen: uint32Ptr(31),
					SuppressIFGroup:   uint32Ptr(37),
					UIDRange: &RuleUIDRange{
						Start: 22,
						End:   25,
					},
					SPortRange: &RulePortRange{
						Start: 23,
						End:   26,
					},
					DPortRange: &RulePortRange{
						Start: 24,
						End:   27,
					},
				},
			},
			b: []byte{
				0x07, 0x06, 0x05, 0x04, 0x03, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x08, 0x00,
				0x0f, 0x00, 0x1d, 0x00, 0x00, 0x00, 0x05, 0x00, 0x15, 0x00, 0x13, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x02, 0x00, 0x08, 0x08, 0x08, 0x08, 0x08, 0x00, 0x01, 0x00, 0x01, 0x01,
				0x01, 0x01, 0x09, 0x00, 0x03, 0x00, 0x65, 0x74, 0x68, 0x30, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x11, 0x00, 0x62, 0x72, 0x30, 0x00, 0x08, 0x00, 0x04, 0x00, 0x01, 0x00,
				0x00, 0x00, 0x08, 0x00, 0x06, 0x00, 0x02, 0x00, 0x00, 0x00, 0x08, 0x00, 0x0a, 0x00,
				0x03, 0x00, 0x00, 0x00, 0x08, 0x00, 0x10, 0x00, 0x05, 0x00, 0x00, 0x00, 0x08, 0x00,
				0x0b, 0x00, 0x0b, 0x00, 0x0d, 0x00, 0x0c, 0x00, 0x0c, 0x00, 0x11, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x13, 0x00, 0x07, 0x00, 0x00, 0x00, 0x05, 0x00,
				0x16, 0x00, 0x17, 0x00, 0x00, 0x00, 0x08, 0x00, 0x0d, 0x00, 0x25, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x0e, 0x00, 0x1f, 0x00, 0x00, 0x00, 0x08, 0x00, 0x14, 0x00, 0x16, 0x00,
				0x19, 0x00, 0x08, 0x00, 0x17, 0x00, 0x17, 0x00, 0x1a, 0x00, 0x08, 0x00, 0x18, 0x00,
				0x18, 0x00, 0x1b, 0x00,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var b []byte
			t.Run("marshal", func(t *testing.T) {
				var marshalErr error
				b, marshalErr = tt.m.MarshalBinary()

				if !errors.Is(marshalErr, tt.marshalErr) {
					t.Fatalf("Expected error '%v' but got '%v'", tt.marshalErr, marshalErr)
				}
			})

			t.Run("compare bytes", func(t *testing.T) {
				if want, got := tt.b, b; !bytes.Equal(want, got) {
					t.Fatalf("unexpected Message bytes:\n- want: [%# x]\n-  got: [%# x]", want, got)
				}
			})

			m := &RuleMessage{}
			t.Run("unmarshal", func(t *testing.T) {
				unmarshalErr := (m).UnmarshalBinary(b)
				if !errors.Is(unmarshalErr, tt.unmarshalErr) {
					t.Fatalf("Expected error '%v' but got '%v'", tt.unmarshalErr, unmarshalErr)
				}
			})

			t.Run("compare messages", func(t *testing.T) {
				if !reflect.DeepEqual(tt.m, m) {
					t.Fatalf("unexpected Message:\n- want: %#v\n-  got: %#v", tt.m, m)
				}
			})
		})
	}

	t.Run("invalid length", func(t *testing.T) {
		m := &RuleMessage{}
		unmarshalErr := (m).UnmarshalBinary([]byte{0x00, 0x01, 0x2, 0x03})
		if !errors.Is(unmarshalErr, errInvalidRuleMessage) {
			t.Fatalf("Expected 'errInvalidRuleMessage' but got '%v'", unmarshalErr)
		}
	})

	t.Run("skipped attributes", func(t *testing.T) {
		m := &RuleMessage{}
		unmarshalErr := (m).UnmarshalBinary([]byte{
			0x01, 0x00, 0x00, 0x02, 0x03, 0x00, 0x00, 0x04, 0x05, 0x00, 0x00, 0x00, 0x04, 0x00,
			0x00, 0x00, 0x04, 0x00, 0x05, 0x00, 0x04, 0x00, 0x07, 0x00, 0x04, 0x00, 0x08, 0x00,
			0x04, 0x00, 0x09, 0x00, 0x04, 0x00, 0x12, 0x00,
		})
		if !errors.Is(unmarshalErr, nil) {
			t.Fatalf("Expected no error but got '%v'", unmarshalErr)
		}
		expected := &RuleMessage{
			Family:     1,
			TOS:        2,
			Table:      3,
			Action:     4,
			Flags:      5,
			Attributes: &RuleAttributes{},
		}
		if !reflect.DeepEqual(expected, m) {
			t.Fatalf("unexpected Message:\n- want: %#v\n-  got: %#v", expected, m)
		}
	})

	t.Run("invalid attribute", func(t *testing.T) {
		m := &RuleMessage{}
		unmarshalErr := (m).UnmarshalBinary([]byte{
			0x01, 0x00, 0x00, 0x02, 0x03, 0x00, 0x00, 0x04, 0x05, 0x00, 0x00, 0x00, 0x04, 0x00,
			0x2a, 0x00,
		})
		if !errors.Is(unmarshalErr, errInvalidRuleAttribute) {
			t.Fatalf("Expected 'errInvalidRuleAttribute' error but got '%v'", unmarshalErr)
		}
	})
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func uint32Ptr(v uint32) *uint32 {
	return &v
}

func uint16Ptr(v uint16) *uint16 {
	return &v
}

func uint8Ptr(v uint8) *uint8 {
	return &v
}

func netIPPtr(v net.IP) *net.IP {
	if ip4 := v.To4(); ip4 != nil {
		// By default net.IP returns the 16 byte representation.
		// But netlink requires us to provide only four bytes
		// for legacy IPs.
		return &ip4
	}
	return &v
}

func strPtr(v string) *string {
	return &v
}
