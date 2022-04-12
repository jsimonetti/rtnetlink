//go:build go1.18
// +build go1.18

package rtnetlink

import "testing"

// FuzzLinkMessage will fuzz a LinkMessage
func FuzzLinkMessage(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		m := &LinkMessage{}
		_ = m.UnmarshalBinary(data)
	})
}

// FuzzAddressMessage will fuzz an AddressMessage
func FuzzAddressMessage(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		m := &LinkMessage{}
		_ = m.UnmarshalBinary(data)
	})
}

// FuzzNeighMessage will fuzz a NeighMessage
func FuzzNeighMessage(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		m := &LinkMessage{}
		_ = m.UnmarshalBinary(data)
	})
}

// FuzzRouteMessage will fuzz a RouteMessage
func FuzzRouteMessage(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		m := &LinkMessage{}
		_ = m.UnmarshalBinary(data)
	})
}

// FuzzRuleMessage will fuzz a RuleMessage
func FuzzRuleMessage(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		m := &RuleMessage{}
		_ = m.UnmarshalBinary(data)
	})
}
