package rtnetlink

import (
	"bytes"
	"reflect"
	"testing"
)

func TestLinkMessageMarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		m    Message
		b    []byte
		err  error
	}{
		{
			name: "empty",
			m:    &LinkMessage{},
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
		},
		{
			name: "no attributes",
			m: &LinkMessage{
				Family: 0,
				Type:   1,
				Index:  2,
				Flags:  0,
				Change: 0,
			},
			b: []byte{
				0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x04, 0x00, 0x01, 0x00, 0x04, 0x00, 0x02, 0x00,
				0x05, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "attributes",
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:   []byte{0, 0, 0, 0, 0, 0},
					Broadcast: []byte{0, 0, 0, 0, 0, 0},
					Name:      "lo",
				},
			},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x02, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x03, 0x00, 0x6c, 0x6f, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "attributes ipip",
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:   []byte{10, 0, 0, 1},
					Broadcast: []byte{255, 255, 255, 255},
					Name:      "ipip",
				},
			},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x01, 0x00, 0x0a, 0x00, 0x00, 0x01,
				0x08, 0x00, 0x02, 0x00, 0xff, 0xff, 0xff, 0xff,
				0x09, 0x00, 0x03, 0x00, 0x69, 0x70, 0x69, 0x70,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x04, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x05, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x06, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "info",
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:   []byte{0, 0, 0, 0, 0, 0},
					Broadcast: []byte{0, 0, 0, 0, 0, 0},
					Name:      "lo",
					Info: &LinkInfo{
						Kind:      "data",
						Data:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
						SlaveKind: "foo",
						SlaveData: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
					},
				},
			},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x02, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x03, 0x00, 0x6c, 0x6f, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x38, 0x00, 0x12, 0x00, 0x09, 0x00, 0x01, 0x00,
				0x64, 0x61, 0x74, 0x61, 0x00, 0x00, 0x00, 0x00,
				0x0d, 0x00, 0x02, 0x00, 0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08, 0x09, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x03, 0x00, 0x66, 0x6f, 0x6f, 0x00,
				0x0d, 0x00, 0x04, 0x00, 0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08, 0x09, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "operational state",
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:          []byte{10, 0, 0, 1},
					Broadcast:        []byte{255, 255, 255, 255},
					Name:             "ipip",
					OperationalState: OperStateUp,
				},
			},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x01, 0x00, 0x0a, 0x00, 0x00, 0x01,
				0x08, 0x00, 0x02, 0x00, 0xff, 0xff, 0xff, 0xff,
				0x09, 0x00, 0x03, 0x00, 0x69, 0x70, 0x69, 0x70,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x04, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x05, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x06, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x10, 0x00,
				0x06, 0x00, 0x00, 0x00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.m.MarshalBinary()

			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error:\n- want: %v\n-  got: %v", want, got)
			}
			if err != nil {
				return
			}

			if want, got := tt.b, b; !bytes.Equal(want, got) {
				t.Fatalf("unexpected Message bytes:\n- want: [%# x]\n-  got: [%# x]", want, got)
			}
		})
	}
}

func TestLinkMessageUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		m    Message
		err  error
	}{
		{
			name: "empty",
			err:  errInvalidLinkMessage,
		},
		{
			name: "short",
			b:    make([]byte, 3),
			err:  errInvalidLinkMessage,
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
			err: errInvalidLinkMessageAttr,
		},
		{
			name: "zero value",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x02, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:   []byte{0, 0, 0, 0, 0, 0},
					Broadcast: []byte{0, 0, 0, 0, 0, 0},
				},
			},
		},
		{
			name: "no data",
			b: []byte{
				0x00, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x02, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			m: &LinkMessage{
				Family: 0,
				Type:   1,
				Index:  2,
				Flags:  0,
				Change: 0,
				Attributes: LinkAttributes{
					Address:   []byte{0, 0, 0, 0, 0, 0},
					Broadcast: []byte{0, 0, 0, 0, 0, 0},
				},
			},
		},
		{
			name: "data",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x02, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x03, 0x00, 0x6c, 0x6f, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:   []byte{0, 0, 0, 0, 0, 0},
					Broadcast: []byte{0, 0, 0, 0, 0, 0},
					Name:      "lo",
				},
			},
		},
		{
			name: "attributes ipip",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x01, 0x00, 0x0a, 0x00, 0x00, 0x01,
				0x08, 0x00, 0x02, 0x00, 0xff, 0xff, 0xff, 0xff,
				0x09, 0x00, 0x03, 0x00, 0x69, 0x70, 0x69, 0x70,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x04, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x05, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x06, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:   []byte{10, 0, 0, 1},
					Broadcast: []byte{255, 255, 255, 255},
					Name:      "ipip",
				},
			},
		},
		{
			name: "info",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x0a, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x02, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x07, 0x00, 0x03, 0x00, 0x6c, 0x6f, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x38, 0x00, 0x12, 0x00, 0x09, 0x00, 0x01, 0x00,
				0x64, 0x61, 0x74, 0x61, 0x00, 0x00, 0x00, 0x00,
				0x0d, 0x00, 0x02, 0x00, 0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08, 0x09, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x03, 0x00, 0x66, 0x6f, 0x6f, 0x00,
				0x0d, 0x00, 0x04, 0x00, 0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08, 0x09, 0x00, 0x00, 0x00,
			},
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:   []byte{0, 0, 0, 0, 0, 0},
					Broadcast: []byte{0, 0, 0, 0, 0, 0},
					Name:      "lo",
					Info: &LinkInfo{
						Kind:      "data",
						Data:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
						SlaveKind: "foo",
						SlaveData: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
					},
				},
			},
		},
		{
			name: "operational state",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x01, 0x00, 0x0a, 0x00, 0x00, 0x01,
				0x08, 0x00, 0x02, 0x00, 0xff, 0xff, 0xff, 0xff,
				0x09, 0x00, 0x03, 0x00, 0x69, 0x70, 0x69, 0x70,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x04, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x05, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x06, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x10, 0x00,
				0x06, 0x00, 0x00, 0x00,
			},
			m: &LinkMessage{
				Attributes: LinkAttributes{
					Address:          []byte{10, 0, 0, 1},
					Broadcast:        []byte{255, 255, 255, 255},
					Name:             "ipip",
					OperationalState: OperStateUp,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LinkMessage{}
			err := (m).UnmarshalBinary(tt.b)

			if want, got := tt.err, err; want != got {
				t.Fatalf("unexpected error:\n- want: %v\n-  got: %v", want, got)
			}
			if err != nil {
				return
			}

			if want, got := tt.m, m; !reflect.DeepEqual(want, got) {
				t.Fatalf("unexpected Message:\n- want: %#v\n-  got: %#v", want, got)
			}
		})
	}
}
