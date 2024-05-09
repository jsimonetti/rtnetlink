package driver

import (
	"github.com/jsimonetti/rtnetlink"
)

// init registers predefined drivers with the rtnetlink package.
//
// Currently, registering driver implementations that conflict with existing ones isn't supported.
// Since most users don't need this feature, we'll keep it as is.
// If required, we could consider implementing rtnetlink.UnregisterDriver to address this.
func init() {
	for _, drv := range []rtnetlink.LinkDriver{
		&Bond{},
		&BondSlave{},
		&Netkit{},
		&Veth{},
	} {
		_ = rtnetlink.RegisterDriver(drv)
	}
}
