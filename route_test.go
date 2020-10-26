package rtnetlink

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jsimonetti/rtnetlink/internal/unix"
)

// Tests will only pass on little endian machines

func TestRouteMessageMarshalUnmarshalBinary(t *testing.T) {
	skipBigEndian(t)

	timeout := uint32(255)
	tests := []struct {
		name string
		m    *RouteMessage
		b    []byte
	}{
		{
			name: "empty",
			m:    &RouteMessage{},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "no attributes",
			m: &RouteMessage{
				Family:    unix.AF_INET,
				DstLength: 8,
				Type:      unix.RTN_UNICAST,
			},
			b: []byte{
				0x02, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "full",
			m: &RouteMessage{
				Family:    2,
				DstLength: 8,
				Table:     unix.RT_TABLE_MAIN,
				Protocol:  unix.RTPROT_STATIC,
				Scope:     unix.RT_SCOPE_UNIVERSE,
				Type:      unix.RTN_UNICAST,
				Attributes: RouteAttributes{
					Dst:      net.IPv4(10, 0, 0, 0),
					Src:      net.IPv4(10, 100, 10, 1),
					Gateway:  net.IPv4(10, 0, 0, 1),
					OutIface: 5,
					Priority: 1,
					Table:    2,
					Mark:     3,
					Expires:  &timeout,
					Metrics: &RouteMetrics{
						AdvMSS:   1,
						Features: 0xffffffff,
						InitCwnd: 2,
						MTU:      1500,
					},
					Multipath: []NextHop{
						{
							Hop: RTNextHop{
								Length:  16,
								IfIndex: 1,
							},
							Gateway: net.IPv4(10, 0, 0, 2),
						},
						{
							Hop: RTNextHop{
								Length:  16,
								IfIndex: 2,
							},
							Gateway: net.IPv4(10, 0, 0, 3),
						},
					},
				},
			},
			b: []byte{
				// RouteMessage struct literal
				//
				// Family
				0x02,
				// DstLength
				0x08,
				// SrcLength
				0x00,
				// Tos
				0x00,
				// Table
				0xfe,
				// Protocol
				0x04,
				// Scope
				0x00,
				// Type
				0x01,
				// Flags
				0x00, 0x00, 0x00, 0x00,
				// RouteAttributes
				// 2 bytes length, 2 bytes type, N bytes value
				//
				// Dst
				0x08, 0x00, 0x01, 0x00,
				0x0a, 0x00, 0x00, 0x00,
				// Src
				0x08, 0x00, 0x07, 0x00,
				0x0a, 0x64, 0x0a, 0x01,
				// Gateway
				0x08, 0x00, 0x05, 0x00,
				0x0a, 0x00, 0x00, 0x01,
				// OutIface
				0x08, 0x00, 0x04, 0x00,
				0x05, 0x00, 0x00, 0x00,
				// Priority
				0x08, 0x00, 0x06, 0x00,
				0x01, 0x00, 0x00, 0x00,
				// Table
				0x08, 0x00, 0x0f, 0x00,
				0x02, 0x00, 0x00, 0x00,
				// Mark
				0x08, 0x00, 0x10, 0x00,
				0x03, 0x00, 0x00, 0x00,
				// Expires
				0x08, 0x00, 0x17, 0x00,
				0xff, 0x00, 0x00, 0x00,
				// RouteMetrics
				// Length must be manually adjusted as more fields are added.
				0x24, 0x00, 0x08, 0x80,
				// AdvMSS
				0x08, 0x00, 0x08, 0x00,
				0x01, 0x00, 0x00, 0x00,
				// Features
				0x08, 0x00, 0x0c, 0x00,
				0xff, 0xff, 0xff, 0xff,
				// InitCwnd
				0x08, 0x00, 0x0b, 0x00,
				0x02, 0x00, 0x00, 0x00,
				// MTU
				0x08, 0x00, 0x02, 0x00,
				0xdc, 0x05, 0x00, 0x00,
				// Multipath
				//
				// 2 bytes length, 2 bytes type, then repeated 8 byte rtnexthop
				// structures followed by their nested netlink attributes.
				0x24, 0x00, 0x09, 0x00,
				// rtnexthop
				0x10, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x00,
				// rtnexthop attributes
				0x08, 0x00, 0x05, 0x00,
				// Gateway
				10, 0, 0, 2,
				// rtnexthop
				0x10, 0x00, 0x00, 0x00,
				0x02, 0x00, 0x00, 0x00,
				// rtnexthop attributes
				0x08, 0x00, 0x05, 0x00,
				// Gateway
				10, 0, 0, 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// It's important to be able to parse raw bytes into valid
			// structures so we start with that step first. After, we'll do a
			// marshaling round-trip to ensure that the structure's byte output
			// and parsed form match what is expected, while also comparing
			// against the expected fixtures throughout.
			var m1 RouteMessage
			if err := m1.UnmarshalBinary(tt.b); err != nil {
				t.Fatalf("failed to unmarshal first message from binary: %v", err)
			}

			if diff := cmp.Diff(tt.m, &m1); diff != "" {
				t.Fatalf("unexpected first message (-want +got):\n%s", diff)
			}

			b, err := m1.MarshalBinary()
			if err != nil {
				t.Fatalf("failed to marshal first message binary: %v", err)
			}

			if diff := cmp.Diff(tt.b, b); diff != "" {
				t.Fatalf("unexpected first message bytes (-want +got):\n%s", diff)
			}

			var m2 RouteMessage
			if err := m2.UnmarshalBinary(b); err != nil {
				t.Fatalf("failed to unmarshal second message from binary: %v", err)
			}

			if diff := cmp.Diff(&m1, &m2); diff != "" {
				t.Fatalf("unexpected parsed messages (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRouteMessageUnmarshalBinaryErrors(t *testing.T) {
	skipBigEndian(t)

	tests := []struct {
		name string
		b    []byte
		m    Message
		err  error
	}{
		{
			name: "empty",
			err:  errInvalidRouteMessage,
		},
		{
			name: "short",
			b:    make([]byte, 3),
			err:  errInvalidRouteMessage,
		},
		{
			name: "invalid attr",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x04, 0x00, 0x01, 0x00, 0x04, 0x00, 0x02, 0x00,
				0x05, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			err: errInvalidRouteMessageAttr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m RouteMessage
			err := m.UnmarshalBinary(tt.b)

			if diff := cmp.Diff(tt.err, err, cmp.Comparer(compareErrors)); diff != "" {
				t.Fatalf("unexpected error (-want +got):\n%s", diff)
			}
		})
	}
}

func compareErrors(x, y error) bool {
	// This is lazy but should be sufficient for the typical stringified errors
	// returned by this package.
	return x.Error() == y.Error()
}
