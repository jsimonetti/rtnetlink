package driver

import (
	"net"
	"testing"

	"github.com/mdlayher/netlink"
)

func TestVxlanEncodeDecode(t *testing.T) {
	var (
		vni100    uint32 = 100
		port4789  uint16 = 4789
		ttl64     uint8  = 64
		tos1      uint8  = 1
		ageing300 uint32 = 300
		limit5000 uint32 = 5000
		trueVal          = true
		falseVal         = false
		dfInherit        = VxlanDFInherit
	)

	tests := []struct {
		name   string
		vxlan  *Vxlan
		verify func(*testing.T, *Vxlan)
	}{
		{
			name: "minimal configuration",
			vxlan: &Vxlan{
				ID: &vni100,
			},
			verify: func(t *testing.T, v *Vxlan) {
				if v.ID == nil || *v.ID != vni100 {
					t.Errorf("expected ID %d, got %v", vni100, v.ID)
				}
			},
		},
		{
			name: "full IPv4 configuration",
			vxlan: &Vxlan{
				ID:      &vni100,
				Group:   net.ParseIP("239.1.1.1"),
				Local:   net.ParseIP("192.168.1.1"),
				Port:    &port4789,
				TTL:     &ttl64,
				TOS:     &tos1,
				Ageing:  &ageing300,
				Limit:   &limit5000,
				Proxy:   &trueVal,
				L2Miss:  &trueVal,
				L3Miss:  &trueVal,
				UDPCsum: &trueVal,
			},
			verify: func(t *testing.T, v *Vxlan) {
				if v.ID == nil || *v.ID != vni100 {
					t.Errorf("expected ID %d, got %v", vni100, v.ID)
				}
				if v.Group == nil || !v.Group.Equal(net.ParseIP("239.1.1.1")) {
					t.Errorf("expected Group 239.1.1.1, got %v", v.Group)
				}
				if v.Local == nil || !v.Local.Equal(net.ParseIP("192.168.1.1")) {
					t.Errorf("expected Local 192.168.1.1, got %v", v.Local)
				}
				if v.Port == nil || *v.Port != port4789 {
					t.Errorf("expected Port %d, got %v", port4789, v.Port)
				}
				if v.TTL == nil || *v.TTL != ttl64 {
					t.Errorf("expected TTL %d, got %v", ttl64, v.TTL)
				}
				if v.Ageing == nil || *v.Ageing != ageing300 {
					t.Errorf("expected Ageing %d, got %v", ageing300, v.Ageing)
				}
			},
		},
		{
			name: "IPv6 configuration",
			vxlan: &Vxlan{
				ID:     &vni100,
				Group6: net.ParseIP("ff05::100"),
				Local6: net.ParseIP("fe80::1"),
			},
			verify: func(t *testing.T, v *Vxlan) {
				if v.ID == nil || *v.ID != vni100 {
					t.Errorf("expected ID %d, got %v", vni100, v.ID)
				}
				if v.Group6 == nil || !v.Group6.Equal(net.ParseIP("ff05::100")) {
					t.Errorf("expected Group6 ff05::100, got %v", v.Group6)
				}
				if v.Local6 == nil || !v.Local6.Equal(net.ParseIP("fe80::1")) {
					t.Errorf("expected Local6 fe80::1, got %v", v.Local6)
				}
			},
		},
		{
			name: "learning disabled",
			vxlan: &Vxlan{
				ID:       &vni100,
				Learning: &falseVal,
			},
			verify: func(t *testing.T, v *Vxlan) {
				if v.Learning == nil || *v.Learning != false {
					t.Errorf("expected Learning false, got %v", v.Learning)
				}
			},
		},
		{
			name: "DF mode inherit",
			vxlan: &Vxlan{
				ID: &vni100,
				DF: &dfInherit,
			},
			verify: func(t *testing.T, v *Vxlan) {
				if v.DF == nil || *v.DF != VxlanDFInherit {
					t.Errorf("expected DF inherit, got %v", v.DF)
				}
			},
		},
		{
			name: "port range",
			vxlan: &Vxlan{
				ID: &vni100,
				PortRange: &VxlanPortRange{
					Low:  10000,
					High: 20000,
				},
			},
			verify: func(t *testing.T, v *Vxlan) {
				if v.PortRange == nil {
					t.Error("expected PortRange to be set")
					return
				}
				if v.PortRange.Low != 10000 || v.PortRange.High != 20000 {
					t.Errorf("expected PortRange 10000-20000, got %d-%d", v.PortRange.Low, v.PortRange.High)
				}
			},
		},
		{
			name: "GBP and GPE flags",
			vxlan: &Vxlan{
				ID:  &vni100,
				GBP: &trueVal,
				GPE: &trueVal,
			},
			verify: func(t *testing.T, v *Vxlan) {
				if v.GBP == nil || *v.GBP != true {
					t.Errorf("expected GBP true, got %v", v.GBP)
				}
				if v.GPE == nil || *v.GPE != true {
					t.Errorf("expected GPE true, got %v", v.GPE)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			ae := netlink.NewAttributeEncoder()
			if err := tt.vxlan.Encode(ae); err != nil {
				t.Fatalf("failed to encode: %v", err)
			}

			encoded, err := ae.Encode()
			if err != nil {
				t.Fatalf("failed to encode attributes: %v", err)
			}

			// Decode
			decoded := &Vxlan{}
			ad, err := netlink.NewAttributeDecoder(encoded)
			if err != nil {
				t.Fatalf("failed to create attribute decoder: %v", err)
			}

			if err := decoded.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			// Verify
			tt.verify(t, decoded)
		})
	}
}

func TestVxlanDFModeString(t *testing.T) {
	tests := []struct {
		mode VxlanDFMode
		want string
	}{
		{VxlanDFUnset, "unset"},
		{VxlanDFSet, "set"},
		{VxlanDFInherit, "inherit"},
		{VxlanDFMode(99), "unknown VxlanDFMode value (99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("VxlanDFMode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVxlanEncodeErrors(t *testing.T) {
	tests := []struct {
		name    string
		vxlan   *Vxlan
		wantErr string
	}{
		{
			name: "invalid IPv4 group",
			vxlan: &Vxlan{
				Group: net.ParseIP("ff05::100"), // IPv6 address for IPv4 field
			},
			wantErr: "group must be an IPv4 address",
		},
		{
			name: "invalid IPv6 group",
			vxlan: &Vxlan{
				Group6: net.ParseIP("239.1.1.1"), // IPv4 address for IPv6 field
			},
			wantErr: "group6 must be an IPv6 address",
		},
		{
			name: "invalid IPv4 local",
			vxlan: &Vxlan{
				Local: net.ParseIP("fe80::1"), // IPv6 address for IPv4 field
			},
			wantErr: "local must be an IPv4 address",
		},
		{
			name: "invalid IPv6 local",
			vxlan: &Vxlan{
				Local6: net.ParseIP("192.168.1.1"), // IPv4 address for IPv6 field
			},
			wantErr: "local6 must be an IPv6 address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ae := netlink.NewAttributeEncoder()
			err := tt.vxlan.Encode(ae)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestVxlanKind(t *testing.T) {
	v := &Vxlan{}
	if kind := v.Kind(); kind != "vxlan" {
		t.Errorf("expected kind %q, got %q", "vxlan", kind)
	}
}

func TestVxlanNew(t *testing.T) {
	v := &Vxlan{}
	newV := v.New()
	if _, ok := newV.(*Vxlan); !ok {
		t.Errorf("expected *Vxlan, got %T", newV)
	}
}
