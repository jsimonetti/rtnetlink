package driver

import (
	"fmt"
	"net"

	"github.com/jsimonetti/rtnetlink"
	"github.com/jsimonetti/rtnetlink/internal/unix"
	"github.com/mdlayher/netlink"
)

// BondMode specifies one of the bonding policies.
type BondMode uint8

const (
	// Round-robin policy: Transmit packets in sequential order from the first available slave through the last
	// This is the default value
	BondModeBalanceRR BondMode = iota

	// Active-backup policy: Only one slave in the bond is active. A different slave becomes active if, and only if,
	// the active slave fails. The bond’s MAC address is externally visible on only one port (network adapter) to
	// avoid confusing the switch.
	BondModeActiveBackup

	// XOR policy: Transmit based on the selected transmit hash policy.
	// The default policy is BOND_XMIT_HASH_POLICY_LAYER2
	// Alternate transmit policies may be selected via the XmitHashPolicy option
	BondModeBalanceXOR

	// Broadcast policy: transmits everything on all slave interfaces
	BondModeBroadcast

	// IEEE 802.3ad Dynamic link aggregation. Creates aggregation groups that share the same speed and duplex settings.
	// Utilizes all slaves in the active aggregator according to the 802.3ad specification
	BondMode802_3AD

	// Adaptive transmit load balancing: channel bonding that does not require any special switch support
	// Outgoing traffic is configured by TlbDynamicLb field
	// Incoming traffic is received by the current slave. If the receiving slave fails,
	// another slave takes over the MAC address of the failed receiving slave.
	BondModeBalanceTLB

	// Adaptive load balancing: includes balance-tlb plus receive load balancing (rlb) for IPV4 traffic,
	// and does not require any special switch support
	BondModeBalanceALB

	BondModeUnknown
)

func (b BondMode) String() string {
	switch b {
	case BondModeBalanceRR:
		return "balance-rr"
	case BondModeActiveBackup:
		return "active-backup"
	case BondModeBalanceXOR:
		return "balance-xor"
	case BondModeBroadcast:
		return "broadcast"
	case BondMode802_3AD:
		return "802.3ad"
	case BondModeBalanceTLB:
		return "balance-tld"
	case BondModeBalanceALB:
		return "balance-alb"
	default:
		return fmt.Sprintf("unknown BondMode value (%d)", b)
	}
}

// BondArpValidate specifies whether or not ARP probes and replies should be validated in any mode that
// supports arp monitoring, or whether non-ARP traffic should be filtered (disregarded) for link monitoring purposes.
type BondArpValidate uint32

const (
	// No validation or filtering is performed
	BondArpValidateNone BondArpValidate = iota

	// Validation is performed only for the active slave
	BondArpValidateActive

	// Validation is performed only for backup slaves
	BondArpValidateBackup

	// Validation is performed for all slaves
	BondArpValidateAll

	// Filtering is applied to all slaves. No validation is performed
	BondArpValidateFilter

	// Filtering is applied to all slaves, validation is performed only for the active slave
	BondArpValidateFilterActive

	// Filtering is applied to all slaves, validation is performed only for backup slaves
	BondArpValidateFilterBackup
)

func (b BondArpValidate) String() string {
	switch b {
	case BondArpValidateNone:
		return "none"
	case BondArpValidateActive:
		return "active"
	case BondArpValidateBackup:
		return "backup"
	case BondArpValidateAll:
		return "all"
	case BondArpValidateFilter:
		return "filter"
	case BondArpValidateFilterActive:
		return "filter_active"
	case BondArpValidateFilterBackup:
		return "filter_backup"
	default:
		return fmt.Sprintf("unknown BondArpValidate value (%d)", b)
	}
}

// BondArpAllTargets specifies the quantity of arp_ip_targets that must be reachable in order for the ARP monitor
// to consider a slave as being up. This option affects only active-backup mode for slaves with arp_validation enabled.
type BondArpAllTargets uint32

const (
	// Consider the slave up only when any of the arp_ip_targets is reachable
	BondArpAllTargetsAny BondArpAllTargets = iota

	// Consider the slave up only when all of the arp_ip_targets are reachable
	BondArpAllTargetsAll
)

func (b BondArpAllTargets) String() string {
	switch b {
	case BondArpAllTargetsAny:
		return "any"
	case BondArpAllTargetsAll:
		return "all"
	default:
		return fmt.Sprintf("unknown BondArpAllTargets value (%d)", b)
	}
}

// Specifies the reselection policy for the primary slave. This affects how the primary slave is
// chosen to become the active slave when failure of the active slave or recovery of the primary slave occurs.
// This option is designed to prevent flip-flopping between the primary slave and other slaves
type BondPrimaryReselect uint8

const (
	// The primary slave becomes the active slave whenever it comes back up, this is the default value
	BondPrimaryReselectAlways BondPrimaryReselect = iota

	// The primary slave becomes the active slave when it comes back up,
	// if the speed and duplex of the primary slave is better than the speed and duplex of the current active slave
	BondPrimaryReselectBetter

	// The primary slave becomes the active slave only if the current active slave fails and the primary slave is up
	BondPrimaryReselectFailure
)

func (b BondPrimaryReselect) String() string {
	switch b {
	case BondPrimaryReselectAlways:
		return "always"
	case BondPrimaryReselectBetter:
		return "better"
	case BondPrimaryReselectFailure:
		return "failure"
	default:
		return fmt.Sprintf("unknown BondPrimaryReselect value (%d)", b)
	}
}

// BondFailOverMac specifies whether active-backup mode should set all slaves to the same MAC address at enslavement
// (the traditional behavior), or, when enabled, perform special handling of the bond’s MAC address
// in accordance with the selected policy.
type BondFailOverMac uint8

const (
	// This setting disables fail_over_mac, and causes bonding to set all slaves of an active-backup bond
	// to the same MAC address at enslavement time
	BondFailOverMacNone BondFailOverMac = iota

	// The “active” fail_over_mac policy indicates that the MAC address of the bond should always be
	// the MAC address of the currently active slave. The MAC address of the slaves is not changed;
	// instead, the MAC address of the bond changes during a failover
	BondFailOverMacActive

	// The “follow” fail_over_mac policy causes the MAC address of the bond to be selected normally
	// (normally the MAC address of the first slave added to the bond)
	// However, the second and subsequent slaves are not set to this MAC address while they are in a backup role;
	// a slave is programmed with the bond’s MAC address at failover time
	// (and the formerly active slave receives the newly active slave’s MAC address)
	BondFailOverMacFollow
)

func (b BondFailOverMac) String() string {
	switch b {
	case BondFailOverMacNone:
		return "none"
	case BondFailOverMacActive:
		return "active"
	case BondFailOverMacFollow:
		return "follow"
	default:
		return fmt.Sprintf("unknown BondPrimaryReselect value (%d)", b)
	}
}

// BondXmitHashPolicy specifies the transmit hash policy to use for
// slave selection in balance-xor, 802.3ad, and tlb modes.
type BondXmitHashPolicy uint8

const (
	// Uses XOR of hardware MAC addresses and packet type ID field to generate the hash
	BondXmitHashPolicyLayer2 BondXmitHashPolicy = iota

	// This policy uses upper layer protocol information, when available, to generate the hash
	// This allows for traffic to a particular network peer to span multiple slaves,
	// although a single connection will not span multiple slaves
	BondXmitHashPolicyLayer3_4

	// This policy uses a combination of layer2 and layer3 protocol information to generate the hash
	// Uses XOR of hardware MAC addresses and IP addresses to generate the hash
	BondXmitHashPolicyLayer2_3

	// This policy uses the same formula as layer2+3 but it relies on skb_flow_dissect to obtain
	// the header fields which might result in the use of inner headers if an encapsulation protocol is used
	BondXmitHashPolicyEncap2_3

	// This policy uses the same formula as layer3+4 but it relies on skb_flow_dissect to obtain
	// the header fields which might result in the use of inner headers if an encapsulation protocol is used
	BondXmitHashPolicyEncap3_4

	// This policy uses a very rudimentary vlan ID and source mac hash to load-balance traffic per-vlan,
	// with failover should one leg fail
	BondXmitHashPolicyVlanSrcMAC
)

func (b BondXmitHashPolicy) String() string {
	switch b {
	case BondXmitHashPolicyLayer2:
		return "layer2"
	case BondXmitHashPolicyLayer3_4:
		return "layer3+4"
	case BondXmitHashPolicyLayer2_3:
		return "layer2+3"
	case BondXmitHashPolicyEncap2_3:
		return "encap2+3"
	case BondXmitHashPolicyEncap3_4:
		return "encap3+4"
	case BondXmitHashPolicyVlanSrcMAC:
		return "vlan+srcmac"
	default:
		return fmt.Sprintf("unknown BondXmitHashPolicy value (%d)", b)
	}
}

// BondAdLacpActive specifies whether to send LACPDU frames periodically.
type BondAdLacpActive uint8

const (
	// LACPDU frames acts as “speak when spoken to”
	BondAdLacpActiveOff BondAdLacpActive = iota

	// LACPDU frames are sent along the configured links periodically with the rate configured with BondLacpRate
	// This is the default value
	BondAdLacpActiveOn
)

func (b BondAdLacpActive) String() string {
	switch b {
	case BondAdLacpActiveOff:
		return "off"
	case BondAdLacpActiveOn:
		return "on"
	default:
		return fmt.Sprintf("unknown BondLacpActive value (%d)", b)
	}
}

// Option specifying the rate in which we’ll ask our link partner to transmit LACPDU packets in 802.3ad mode.
type BondLacpRate uint8

const (
	// Request partner to transmit LACPDUs every 30 seconds
	// This is the default value
	BondLacpRateSlow BondLacpRate = iota

	// Request partner to transmit LACPDUs every 1 second
	BondLacpRateFast
)

func (b BondLacpRate) String() string {
	switch b {
	case BondLacpRateSlow:
		return "slow"
	case BondLacpRateFast:
		return "fast"
	default:
		return fmt.Sprintf("unknown BondLacpRate value (%d)", b)
	}
}

// BondAdSelect specifies the 802.3ad aggregation selection logic to use.
type BondAdSelect uint8

const (
	// The active aggregator is chosen by largest aggregate bandwidth
	// Reselection of the active aggregator occurs only when all slaves of the active aggregator
	// are down or the active aggregator has no slaves
	// This is the default value.
	BondAdSelectStable BondAdSelect = iota

	// The active aggregator is chosen by largest aggregate bandwidth.
	// Reselection occurs if:
	//  - A slave is added to or removed from the bond
	//  - Any slave’s link state changes
	//  - Any slave’s 802.3ad association state changes
	//  - The bond’s administrative state changes to up
	BondAdSelectBandwidth

	// The active aggregator is chosen by the largest number of ports (slaves)
	// Reselection rules are the same with BOND_AD_SELECT_BANDWIDTH
	BondAdSelectCount
)

func (b BondAdSelect) String() string {
	switch b {
	case BondAdSelectStable:
		return "stable"
	case BondAdSelectBandwidth:
		return "bandwidth"
	case BondAdSelectCount:
		return "count"
	default:
		return fmt.Sprintf("unknown BondAdSelect value (%d)", b)
	}
}

// BondAdInfo specifies the 802.3ad aggregation information
type BondAdInfo struct {
	AggregatorId uint16
	NumPorts     uint16
	ActorKey     uint16
	PartnerKey   uint16
	PartnerMac   net.HardwareAddr
}

const bondMaxTargets = 16

// Bond implements LinkDriver for the bond driver
type Bond struct {
	// For more detailed information see https://www.kernel.org/doc/html/latest/networking/bonding.html

	// Specifies the bonding policy. The default is balance-rr (round robin)
	Mode BondMode

	// Specifies the new active slave for modes that support it (active-backup, balance-alb and balance-tlb)
	ActiveSlave *uint32

	// Specifies the MII link monitoring frequency in milliseconds
	Miimon *uint32

	// Specifies the time, in milliseconds, to wait before enabling a slave after a link recovery has been detected
	UpDelay *uint32

	// Specifies the time, in milliseconds, to wait before disabling a slave after a link failure has been detected
	DownDelay *uint32

	// Specify the delay, in milliseconds, between each peer notification
	PeerNotifyDelay *uint32

	// Specifies whether or not miimon should use MII or ETHTOOL
	UseCarrier *uint8

	// Specifies the ARP link monitoring frequency in milliseconds
	ArpInterval *uint32

	// Specifies the IP addresses to use as ARP monitoring peers when arp_interval is > 0
	ArpIpTargets []net.IP

	// Specifies the IPv6 addresses to use as IPv6 monitoring peers when arp_interval is > 0
	NsIP6Targets []net.IP

	// Specifies whether or not ARP probes and replies should be validated
	ArpValidate *BondArpValidate

	// Specifies the quantity of arp_ip_targets that must be reachable in order for the ARP monitor to consider a slave as being up
	ArpAllTargets *BondArpAllTargets

	// A device index specifying which slave is the primary device
	Primary *uint32

	// Specifies the reselection policy for the primary slave
	PrimaryReselect *BondPrimaryReselect

	// Specifies whether active-backup mode should set all slaves to the same MAC address at enslavement, when enabled, or perform special handling
	FailOverMac *BondFailOverMac

	// Selects the transmit hash policy to use for slave selection
	XmitHashPolicy *BondXmitHashPolicy

	// Specifies the number of IGMP membership reports to be issued after a failover event
	ResendIgmp *uint32

	// Specify the number of peer notifications (gratuitous ARPs and unsolicited IPv6 Neighbor Advertisements) to be issued after a failover event
	NumPeerNotif *uint8

	// Specifies that duplicate frames (received on inactive ports) should be dropped (0) or delivered (1)
	AllSlavesActive *uint8

	// Specifies the minimum number of links that must be active before asserting carrier
	MinLinks *uint32

	// Specifies the number of seconds between instances where the bonding driver sends learning packets to each slaves peer switch
	LpInterval *uint32

	// Specify the number of packets to transmit through a slave before moving to the next one
	PacketsPerSlave *uint32

	// Option specifying whether to send LACPDU frames periodically
	AdLacpActive *BondAdLacpActive

	// Option specifying the rate in which we’ll ask our link partner to transmit LACPDU packets
	AdLacpRate *BondLacpRate

	// Specifies the 802.3ad aggregation selection logic to use
	AdSelect *BondAdSelect

	// In an AD system, this specifies the system priority
	AdActorSysPrio *uint16

	// Defines the upper 10 bits of the port key
	AdUserPortKey *uint16

	// In an AD system, this specifies the mac-address for the actor in protocol packet exchanges
	AdActorSystem net.HardwareAddr

	// Specifies if dynamic shuffling of flows is enabled in tlb or alb mode
	TlbDynamicLb *uint8

	// Specifies the number of arp_interval monitor checks that must fail in order for an interface to be marked down by the ARP monitor
	MissedMax *uint8

	// Specifies the 802.3ad aggregation information, this is read only value
	AdInfo *BondAdInfo
}

var _ rtnetlink.LinkDriver = &Bond{}

func (b *Bond) New() rtnetlink.LinkDriver {
	return &Bond{}
}

func (b *Bond) Encode(ae *netlink.AttributeEncoder) error {
	if b.Mode < BondModeUnknown {
		ae.Uint8(unix.IFLA_BOND_MODE, uint8(b.Mode))
	}
	if b.ActiveSlave != nil {
		ae.Uint32(unix.IFLA_BOND_ACTIVE_SLAVE, *b.ActiveSlave)
	}
	if b.Miimon != nil {
		ae.Uint32(unix.IFLA_BOND_MIIMON, *b.Miimon)
	}
	if b.UpDelay != nil {
		ae.Uint32(unix.IFLA_BOND_UPDELAY, *b.UpDelay)
	}
	if b.DownDelay != nil {
		ae.Uint32(unix.IFLA_BOND_DOWNDELAY, *b.DownDelay)
	}
	if b.PeerNotifyDelay != nil {
		ae.Uint32(unix.IFLA_BOND_PEER_NOTIF_DELAY, *b.PeerNotifyDelay)
	}
	if b.UseCarrier != nil {
		ae.Uint8(unix.IFLA_BOND_USE_CARRIER, *b.UseCarrier)
	}
	if b.ArpInterval != nil {
		ae.Uint32(unix.IFLA_BOND_ARP_INTERVAL, *b.ArpInterval)
	}
	if b.ArpIpTargets != nil {
		if lb := len(b.ArpIpTargets); lb > bondMaxTargets {
			return fmt.Errorf("exceeded max ArpIpTargets %d, %d", bondMaxTargets, lb)
		}
		ae.Nested(unix.IFLA_BOND_ARP_IP_TARGET, func(nae *netlink.AttributeEncoder) error {
			for i := range b.ArpIpTargets {
				ip := b.ArpIpTargets[i].To4()
				if ip == nil {
					return fmt.Errorf("%s is not an ip4 address", ip)
				}
				nae.Bytes(uint16(i), ip)
			}
			return nil
		})
	}
	if b.NsIP6Targets != nil {
		if lb := len(b.ArpIpTargets); lb > bondMaxTargets {
			return fmt.Errorf("exceeded max NsIP6Targets %d, %d", bondMaxTargets, lb)
		}
		ae.Nested(unix.IFLA_BOND_NS_IP6_TARGET, func(nae *netlink.AttributeEncoder) error {
			for i := range b.NsIP6Targets {
				ip := b.NsIP6Targets[i].To16()
				if ip == nil {
					return fmt.Errorf("%s is not an ip6 address", ip)
				}
				nae.Bytes(uint16(i), ip)
			}
			return nil
		})
	}
	if b.ArpValidate != nil {
		ae.Uint32(unix.IFLA_BOND_ARP_VALIDATE, uint32(*b.ArpValidate))
	}
	if b.ArpAllTargets != nil {
		ae.Uint32(unix.IFLA_BOND_ARP_ALL_TARGETS, uint32(*b.ArpAllTargets))
	}
	if b.Primary != nil {
		ae.Uint32(unix.IFLA_BOND_PRIMARY, *b.Primary)
	}
	if b.PrimaryReselect != nil {
		ae.Uint8(unix.IFLA_BOND_PRIMARY_RESELECT, uint8(*b.PrimaryReselect))
	}
	if b.FailOverMac != nil {
		ae.Uint8(unix.IFLA_BOND_FAIL_OVER_MAC, uint8(*b.FailOverMac))
	}
	if b.XmitHashPolicy != nil {
		ae.Uint8(unix.IFLA_BOND_XMIT_HASH_POLICY, uint8(*b.XmitHashPolicy))
	}
	if b.ResendIgmp != nil {
		ae.Uint32(unix.IFLA_BOND_RESEND_IGMP, *b.ResendIgmp)
	}
	if b.NumPeerNotif != nil {
		ae.Uint8(unix.IFLA_BOND_NUM_PEER_NOTIF, *b.NumPeerNotif)
	}
	if b.AllSlavesActive != nil {
		ae.Uint8(unix.IFLA_BOND_ALL_SLAVES_ACTIVE, *b.AllSlavesActive)
	}
	if b.MinLinks != nil {
		ae.Uint32(unix.IFLA_BOND_MIN_LINKS, *b.MinLinks)
	}
	if b.LpInterval != nil {
		ae.Uint32(unix.IFLA_BOND_LP_INTERVAL, *b.LpInterval)
	}
	if b.PacketsPerSlave != nil {
		ae.Uint32(unix.IFLA_BOND_PACKETS_PER_SLAVE, *b.PacketsPerSlave)
	}
	if b.AdLacpActive != nil {
		ae.Uint8(unix.IFLA_BOND_AD_LACP_ACTIVE, uint8(*b.AdLacpActive))
	}
	if b.AdLacpRate != nil {
		ae.Uint8(unix.IFLA_BOND_AD_LACP_RATE, uint8(*b.AdLacpRate))
	}
	if b.AdSelect != nil {
		ae.Uint8(unix.IFLA_BOND_AD_SELECT, uint8(*b.AdSelect))
	}
	if b.AdActorSysPrio != nil {
		ae.Uint16(unix.IFLA_BOND_AD_ACTOR_SYS_PRIO, *b.AdActorSysPrio)
	}
	if b.AdUserPortKey != nil {
		ae.Uint16(unix.IFLA_BOND_AD_USER_PORT_KEY, *b.AdUserPortKey)
	}
	if b.AdActorSystem != nil {
		ae.Bytes(unix.IFLA_BOND_AD_ACTOR_SYSTEM, []byte(b.AdActorSystem))
	}
	if b.TlbDynamicLb != nil {
		ae.Uint8(unix.IFLA_BOND_TLB_DYNAMIC_LB, *b.TlbDynamicLb)
	}
	if b.MissedMax != nil {
		ae.Uint8(unix.IFLA_BOND_MISSED_MAX, *b.MissedMax)
	}

	return nil
}

func (b *Bond) Decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.IFLA_BOND_MODE:
			b.Mode = BondMode(ad.Uint8())
		case unix.IFLA_BOND_ACTIVE_SLAVE:
			v := ad.Uint32()
			b.ActiveSlave = &v
		case unix.IFLA_BOND_MIIMON:
			v := ad.Uint32()
			b.Miimon = &v
		case unix.IFLA_BOND_UPDELAY:
			v := ad.Uint32()
			b.UpDelay = &v
		case unix.IFLA_BOND_DOWNDELAY:
			v := ad.Uint32()
			b.DownDelay = &v
		case unix.IFLA_BOND_PEER_NOTIF_DELAY:
			v := ad.Uint32()
			b.PeerNotifyDelay = &v
		case unix.IFLA_BOND_USE_CARRIER:
			v := ad.Uint8()
			b.UseCarrier = &v
		case unix.IFLA_BOND_ARP_INTERVAL:
			v := ad.Uint32()
			b.ArpInterval = &v
		case unix.IFLA_BOND_ARP_IP_TARGET:
			ad.Nested(func(nad *netlink.AttributeDecoder) error {
				for nad.Next() {
					b.ArpIpTargets = append(b.ArpIpTargets, nad.Bytes())
				}
				return nil
			})
		case unix.IFLA_BOND_NS_IP6_TARGET:
			ad.Nested(func(nad *netlink.AttributeDecoder) error {
				for nad.Next() {
					b.NsIP6Targets = append(b.NsIP6Targets, nad.Bytes())
				}
				return nil
			})
		case unix.IFLA_BOND_ARP_VALIDATE:
			v := BondArpValidate(ad.Uint32())
			b.ArpValidate = &v
		case unix.IFLA_BOND_ARP_ALL_TARGETS:
			v := BondArpAllTargets(ad.Uint32())
			b.ArpAllTargets = &v
		case unix.IFLA_BOND_PRIMARY:
			v := ad.Uint32()
			b.Primary = &v
		case unix.IFLA_BOND_PRIMARY_RESELECT:
			v := BondPrimaryReselect(ad.Uint8())
			b.PrimaryReselect = &v
		case unix.IFLA_BOND_FAIL_OVER_MAC:
			v := BondFailOverMac(ad.Uint8())
			b.FailOverMac = &v
		case unix.IFLA_BOND_XMIT_HASH_POLICY:
			v := BondXmitHashPolicy(ad.Uint8())
			b.XmitHashPolicy = &v
		case unix.IFLA_BOND_RESEND_IGMP:
			v := ad.Uint32()
			b.ResendIgmp = &v
		case unix.IFLA_BOND_NUM_PEER_NOTIF:
			v := ad.Uint8()
			b.NumPeerNotif = &v
		case unix.IFLA_BOND_ALL_SLAVES_ACTIVE:
			v := ad.Uint8()
			b.AllSlavesActive = &v
		case unix.IFLA_BOND_MIN_LINKS:
			v := ad.Uint32()
			b.MinLinks = &v
		case unix.IFLA_BOND_LP_INTERVAL:
			v := ad.Uint32()
			b.LpInterval = &v
		case unix.IFLA_BOND_PACKETS_PER_SLAVE:
			v := ad.Uint32()
			b.PacketsPerSlave = &v
		case unix.IFLA_BOND_AD_LACP_ACTIVE:
			v := BondAdLacpActive(ad.Uint8())
			b.AdLacpActive = &v
		case unix.IFLA_BOND_AD_LACP_RATE:
			v := BondLacpRate(ad.Uint8())
			b.AdLacpRate = &v
		case unix.IFLA_BOND_AD_SELECT:
			v := BondAdSelect(ad.Uint8())
			b.AdSelect = &v
		case unix.IFLA_BOND_AD_ACTOR_SYS_PRIO:
			v := ad.Uint16()
			b.AdActorSysPrio = &v
		case unix.IFLA_BOND_AD_USER_PORT_KEY:
			v := ad.Uint16()
			b.AdUserPortKey = &v
		case unix.IFLA_BOND_AD_ACTOR_SYSTEM:
			b.AdActorSystem = ad.Bytes()
		case unix.IFLA_BOND_TLB_DYNAMIC_LB:
			v := ad.Uint8()
			b.TlbDynamicLb = &v
		case unix.IFLA_BOND_MISSED_MAX:
			v := ad.Uint8()
			b.MissedMax = &v
		case unix.IFLA_BOND_AD_INFO:
			ad.Nested(func(nad *netlink.AttributeDecoder) error {
				b.AdInfo = &BondAdInfo{}
				for nad.Next() {
					switch nad.Type() {
					case unix.IFLA_BOND_AD_INFO_AGGREGATOR:
						b.AdInfo.AggregatorId = nad.Uint16()
					case unix.IFLA_BOND_AD_INFO_NUM_PORTS:
						b.AdInfo.NumPorts = nad.Uint16()
					case unix.IFLA_BOND_AD_INFO_ACTOR_KEY:
						b.AdInfo.ActorKey = nad.Uint16()
					case unix.IFLA_BOND_AD_INFO_PARTNER_KEY:
						b.AdInfo.PartnerKey = nad.Uint16()
					case unix.IFLA_BOND_AD_INFO_PARTNER_MAC:
						b.AdInfo.PartnerMac = nad.Bytes()
					}
				}
				return nil
			})
		}
	}
	return nil
}

func (*Bond) Kind() string {
	return "bond"
}

// BondSlaveState specifies bond slave state
type BondSlaveState uint8

const (
	BondStateActive BondSlaveState = iota
	BondStateBackup
)

func (b BondSlaveState) String() string {
	switch b {
	case BondStateActive:
		return "ACTIVE"
	case BondStateBackup:
		return "BACKUP"
	default:
		return fmt.Sprintf("unknown BondSlaveState value %d", b)
	}
}

// BondSlaveMiiStatus MII link monitoring frequency status
type BondSlaveMiiStatus uint8

const (
	BondLinkUp BondSlaveMiiStatus = iota
	BondLinkFail
	BondLinkDown
	BondLinkBack
)

func (b BondSlaveMiiStatus) String() string {
	switch b {
	case BondLinkUp:
		return "UP"
	case BondLinkFail:
		return "GOING_DOWN"
	case BondLinkDown:
		return "DOWN"
	case BondLinkBack:
		return "GOING_BACK"
	default:
		return fmt.Sprintf("unknown BondSlaveMiiStatus value %d", b)
	}
}

// BondSlave implements LinkSlaveDriver interface for bond driver
type BondSlave struct {
	State                  *BondSlaveState
	MiiStatus              *BondSlaveMiiStatus
	LinkFailureCount       *uint32
	PermHardwareAddr       net.HardwareAddr
	QueueId                *uint16
	Priority               *int32
	AggregatorId           *uint16
	AdActorOperPortState   *uint8
	AdPartnerOperPortState *uint16
}

var _ rtnetlink.LinkSlaveDriver = &BondSlave{}

func (b *BondSlave) New() rtnetlink.LinkDriver {
	return &BondSlave{}
}

func (b *BondSlave) Slave() {}

func (b *BondSlave) Encode(ae *netlink.AttributeEncoder) error {
	if b.QueueId != nil {
		ae.Uint16(unix.IFLA_BOND_SLAVE_QUEUE_ID, *b.QueueId)
	}
	if b.Priority != nil {
		ae.Int32(unix.IFLA_BOND_SLAVE_PRIO, *b.Priority)
	}
	return nil
}

func (b *BondSlave) Decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.IFLA_BOND_SLAVE_STATE:
			v := BondSlaveState(ad.Uint8())
			b.State = &v
		case unix.IFLA_BOND_SLAVE_MII_STATUS:
			v := BondSlaveMiiStatus(ad.Uint8())
			b.MiiStatus = &v
		case unix.IFLA_BOND_SLAVE_LINK_FAILURE_COUNT:
			v := ad.Uint32()
			b.LinkFailureCount = &v
		case unix.IFLA_BOND_SLAVE_PERM_HWADDR:
			b.PermHardwareAddr = net.HardwareAddr(ad.Bytes())
		case unix.IFLA_BOND_SLAVE_QUEUE_ID:
			v := ad.Uint16()
			b.QueueId = &v
		case unix.IFLA_BOND_SLAVE_PRIO:
			v := ad.Int32()
			b.Priority = &v
		case unix.IFLA_BOND_SLAVE_AD_AGGREGATOR_ID:
			v := ad.Uint16()
			b.AggregatorId = &v
		case unix.IFLA_BOND_SLAVE_AD_ACTOR_OPER_PORT_STATE:
			v := ad.Uint8()
			b.AdActorOperPortState = &v
		case unix.IFLA_BOND_SLAVE_AD_PARTNER_OPER_PORT_STATE:
			v := ad.Uint16()
			b.AdPartnerOperPortState = &v
		}
	}
	return nil
}

func (*BondSlave) Kind() string {
	return "bond"
}
