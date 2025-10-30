package driver

import (
	"fmt"
	"net"

	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"github.com/mdlayher/netlink"
)

// BridgeStpState represents the Spanning Tree Protocol state.
type BridgeStpState uint32

// STP states.
const (
	BridgeStpStateDisabled BridgeStpState = 0
	BridgeStpStateEnabled  BridgeStpState = 1
)

// String returns a string representation of the BridgeStpState.
func (s BridgeStpState) String() string {
	switch s {
	case BridgeStpStateDisabled:
		return "disabled"
	case BridgeStpStateEnabled:
		return "enabled"
	default:
		return fmt.Sprintf("unknown BridgeStpState value (%d)", uint32(s))
	}
}

// BridgeEnable represents an enable/disable flag for various bridge features.
type BridgeEnable uint8

// Enable/disable states.
const (
	BridgeEnableDisabled BridgeEnable = 0
	BridgeEnableEnabled  BridgeEnable = 1
)

// String returns a string representation of the BridgeEnable.
func (e BridgeEnable) String() string {
	switch e {
	case BridgeEnableDisabled:
		return "disabled"
	case BridgeEnableEnabled:
		return "enabled"
	default:
		return fmt.Sprintf("unknown BridgeEnable value (%d)", uint8(e))
	}
}

// Bridge implements LinkDriver for the bridge driver
type Bridge struct {
	// For more detailed information see https://www.kernel.org/doc/html/latest/networking/bridge.html

	// Forward delay time in centiseconds (default: 1500, which is 15 seconds)
	ForwardDelay *uint32

	// Hello time in centiseconds (default: 200, which is 2 seconds)
	HelloTime *uint32

	// Maximum message age in centiseconds (default: 2000, which is 20 seconds)
	MaxAge *uint32

	// Ageing time in centiseconds (default: 30000, which is 300 seconds/5 minutes)
	AgeingTime *uint32

	// STP state (disabled/enabled)
	StpState *BridgeStpState

	// Bridge priority (default: 32768)
	Priority *uint16

	// VLAN filtering
	VlanFiltering *BridgeEnable

	// VLAN protocol (e.g., VlanProtocol8021Q for 802.1Q, VlanProtocol8021AD for 802.1ad)
	VlanProtocol *VlanProtocol

	// Group forward mask
	GroupFwdMask *uint16

	// Group address (MAC address)
	GroupAddr net.HardwareAddr

	// FDB flush - triggers a flush of the forwarding database
	FdbFlush *uint8

	// Multicast router setting (0=disabled, 1=enabled, 2=auto)
	McastRouter *uint8

	// Multicast snooping
	McastSnooping *BridgeEnable

	// Multicast query use IFADDR
	McastQueryUseIfaddr *BridgeEnable

	// Multicast querier
	McastQuerier *BridgeEnable

	// Multicast hash elasticity
	McastHashElasticity *uint32

	// Multicast hash max
	McastHashMax *uint32

	// Multicast last member count
	McastLastMemberCnt *uint32

	// Multicast startup query count
	McastStartupQueryCnt *uint32

	// Multicast last member interval in centiseconds
	McastLastMemberIntvl *uint64

	// Multicast membership interval in centiseconds
	McastMembershipIntvl *uint64

	// Multicast querier interval in centiseconds
	McastQuerierIntvl *uint64

	// Multicast query interval in centiseconds
	McastQueryIntvl *uint64

	// Multicast query response interval in centiseconds
	McastQueryResponseIntvl *uint64

	// Multicast startup query interval in centiseconds
	McastStartupQueryIntvl *uint64

	// NF call iptables
	NfCallIptables *BridgeEnable

	// NF call ip6tables
	NfCallIp6tables *BridgeEnable

	// NF call arptables
	NfCallArptables *BridgeEnable

	// VLAN default PVID
	VlanDefaultPvid *uint16

	// VLAN stats enabled
	VlanStatsEnabled *BridgeEnable

	// Multicast stats enabled
	McastStatsEnabled *BridgeEnable

	// Multicast IGMP version (2 or 3)
	McastIgmpVersion *uint8

	// Multicast MLD version (1 or 2)
	McastMldVersion *uint8

	// VLAN stats per port
	VlanStatsPerPort *BridgeEnable

	// FDB max learned entries (0=unlimited)
	FdbMaxLearned *uint32
}

var _ rtnetlink.LinkDriver = &Bridge{}

func (b *Bridge) New() rtnetlink.LinkDriver {
	return &Bridge{}
}

func (b *Bridge) Encode(ae *netlink.AttributeEncoder) error {
	if b.ForwardDelay != nil {
		ae.Uint32(unix.IFLA_BR_FORWARD_DELAY, *b.ForwardDelay)
	}
	if b.HelloTime != nil {
		ae.Uint32(unix.IFLA_BR_HELLO_TIME, *b.HelloTime)
	}
	if b.MaxAge != nil {
		ae.Uint32(unix.IFLA_BR_MAX_AGE, *b.MaxAge)
	}
	if b.AgeingTime != nil {
		ae.Uint32(unix.IFLA_BR_AGEING_TIME, *b.AgeingTime)
	}
	if b.StpState != nil {
		ae.Uint32(unix.IFLA_BR_STP_STATE, uint32(*b.StpState))
	}
	if b.Priority != nil {
		ae.Uint16(unix.IFLA_BR_PRIORITY, *b.Priority)
	}
	if b.VlanFiltering != nil {
		ae.Uint8(unix.IFLA_BR_VLAN_FILTERING, uint8(*b.VlanFiltering))
	}
	if b.VlanProtocol != nil {
		ae.Uint16(unix.IFLA_BR_VLAN_PROTOCOL, uint16(*b.VlanProtocol))
	}
	if b.GroupFwdMask != nil {
		ae.Uint16(unix.IFLA_BR_GROUP_FWD_MASK, *b.GroupFwdMask)
	}
	if b.GroupAddr != nil {
		ae.Bytes(unix.IFLA_BR_GROUP_ADDR, []byte(b.GroupAddr))
	}
	if b.FdbFlush != nil {
		ae.Uint8(unix.IFLA_BR_FDB_FLUSH, *b.FdbFlush)
	}
	if b.McastRouter != nil {
		ae.Uint8(unix.IFLA_BR_MCAST_ROUTER, *b.McastRouter)
	}
	if b.McastSnooping != nil {
		ae.Uint8(unix.IFLA_BR_MCAST_SNOOPING, uint8(*b.McastSnooping))
	}
	if b.McastQueryUseIfaddr != nil {
		ae.Uint8(unix.IFLA_BR_MCAST_QUERY_USE_IFADDR, uint8(*b.McastQueryUseIfaddr))
	}
	if b.McastQuerier != nil {
		ae.Uint8(unix.IFLA_BR_MCAST_QUERIER, uint8(*b.McastQuerier))
	}
	if b.McastHashElasticity != nil {
		ae.Uint32(unix.IFLA_BR_MCAST_HASH_ELASTICITY, *b.McastHashElasticity)
	}
	if b.McastHashMax != nil {
		ae.Uint32(unix.IFLA_BR_MCAST_HASH_MAX, *b.McastHashMax)
	}
	if b.McastLastMemberCnt != nil {
		ae.Uint32(unix.IFLA_BR_MCAST_LAST_MEMBER_CNT, *b.McastLastMemberCnt)
	}
	if b.McastStartupQueryCnt != nil {
		ae.Uint32(unix.IFLA_BR_MCAST_STARTUP_QUERY_CNT, *b.McastStartupQueryCnt)
	}
	if b.McastLastMemberIntvl != nil {
		ae.Uint64(unix.IFLA_BR_MCAST_LAST_MEMBER_INTVL, *b.McastLastMemberIntvl)
	}
	if b.McastMembershipIntvl != nil {
		ae.Uint64(unix.IFLA_BR_MCAST_MEMBERSHIP_INTVL, *b.McastMembershipIntvl)
	}
	if b.McastQuerierIntvl != nil {
		ae.Uint64(unix.IFLA_BR_MCAST_QUERIER_INTVL, *b.McastQuerierIntvl)
	}
	if b.McastQueryIntvl != nil {
		ae.Uint64(unix.IFLA_BR_MCAST_QUERY_INTVL, *b.McastQueryIntvl)
	}
	if b.McastQueryResponseIntvl != nil {
		ae.Uint64(unix.IFLA_BR_MCAST_QUERY_RESPONSE_INTVL, *b.McastQueryResponseIntvl)
	}
	if b.McastStartupQueryIntvl != nil {
		ae.Uint64(unix.IFLA_BR_MCAST_STARTUP_QUERY_INTVL, *b.McastStartupQueryIntvl)
	}
	if b.NfCallIptables != nil {
		ae.Uint8(unix.IFLA_BR_NF_CALL_IPTABLES, uint8(*b.NfCallIptables))
	}
	if b.NfCallIp6tables != nil {
		ae.Uint8(unix.IFLA_BR_NF_CALL_IP6TABLES, uint8(*b.NfCallIp6tables))
	}
	if b.NfCallArptables != nil {
		ae.Uint8(unix.IFLA_BR_NF_CALL_ARPTABLES, uint8(*b.NfCallArptables))
	}
	if b.VlanDefaultPvid != nil {
		ae.Uint16(unix.IFLA_BR_VLAN_DEFAULT_PVID, *b.VlanDefaultPvid)
	}
	if b.VlanStatsEnabled != nil {
		ae.Uint8(unix.IFLA_BR_VLAN_STATS_ENABLED, uint8(*b.VlanStatsEnabled))
	}
	if b.McastStatsEnabled != nil {
		ae.Uint8(unix.IFLA_BR_MCAST_STATS_ENABLED, uint8(*b.McastStatsEnabled))
	}
	if b.McastIgmpVersion != nil {
		ae.Uint8(unix.IFLA_BR_MCAST_IGMP_VERSION, *b.McastIgmpVersion)
	}
	if b.McastMldVersion != nil {
		ae.Uint8(unix.IFLA_BR_MCAST_MLD_VERSION, *b.McastMldVersion)
	}
	if b.VlanStatsPerPort != nil {
		ae.Uint8(unix.IFLA_BR_VLAN_STATS_PER_PORT, uint8(*b.VlanStatsPerPort))
	}
	if b.FdbMaxLearned != nil {
		ae.Uint32(unix.IFLA_BR_FDB_MAX_LEARNED, *b.FdbMaxLearned)
	}

	return nil
}

func (b *Bridge) Decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.IFLA_BR_FORWARD_DELAY:
			v := ad.Uint32()
			b.ForwardDelay = &v
		case unix.IFLA_BR_HELLO_TIME:
			v := ad.Uint32()
			b.HelloTime = &v
		case unix.IFLA_BR_MAX_AGE:
			v := ad.Uint32()
			b.MaxAge = &v
		case unix.IFLA_BR_AGEING_TIME:
			v := ad.Uint32()
			b.AgeingTime = &v
		case unix.IFLA_BR_STP_STATE:
			v := BridgeStpState(ad.Uint32())
			b.StpState = &v
		case unix.IFLA_BR_PRIORITY:
			v := ad.Uint16()
			b.Priority = &v
		case unix.IFLA_BR_VLAN_FILTERING:
			v := BridgeEnable(ad.Uint8())
			b.VlanFiltering = &v
		case unix.IFLA_BR_VLAN_PROTOCOL:
			v := VlanProtocol(ad.Uint16())
			b.VlanProtocol = &v
		case unix.IFLA_BR_GROUP_FWD_MASK:
			v := ad.Uint16()
			b.GroupFwdMask = &v
		case unix.IFLA_BR_GROUP_ADDR:
			b.GroupAddr = net.HardwareAddr(ad.Bytes())
		case unix.IFLA_BR_FDB_FLUSH:
			v := ad.Uint8()
			b.FdbFlush = &v
		case unix.IFLA_BR_MCAST_ROUTER:
			v := ad.Uint8()
			b.McastRouter = &v
		case unix.IFLA_BR_MCAST_SNOOPING:
			v := BridgeEnable(ad.Uint8())
			b.McastSnooping = &v
		case unix.IFLA_BR_MCAST_QUERY_USE_IFADDR:
			v := BridgeEnable(ad.Uint8())
			b.McastQueryUseIfaddr = &v
		case unix.IFLA_BR_MCAST_QUERIER:
			v := BridgeEnable(ad.Uint8())
			b.McastQuerier = &v
		case unix.IFLA_BR_MCAST_HASH_ELASTICITY:
			v := ad.Uint32()
			b.McastHashElasticity = &v
		case unix.IFLA_BR_MCAST_HASH_MAX:
			v := ad.Uint32()
			b.McastHashMax = &v
		case unix.IFLA_BR_MCAST_LAST_MEMBER_CNT:
			v := ad.Uint32()
			b.McastLastMemberCnt = &v
		case unix.IFLA_BR_MCAST_STARTUP_QUERY_CNT:
			v := ad.Uint32()
			b.McastStartupQueryCnt = &v
		case unix.IFLA_BR_MCAST_LAST_MEMBER_INTVL:
			v := ad.Uint64()
			b.McastLastMemberIntvl = &v
		case unix.IFLA_BR_MCAST_MEMBERSHIP_INTVL:
			v := ad.Uint64()
			b.McastMembershipIntvl = &v
		case unix.IFLA_BR_MCAST_QUERIER_INTVL:
			v := ad.Uint64()
			b.McastQuerierIntvl = &v
		case unix.IFLA_BR_MCAST_QUERY_INTVL:
			v := ad.Uint64()
			b.McastQueryIntvl = &v
		case unix.IFLA_BR_MCAST_QUERY_RESPONSE_INTVL:
			v := ad.Uint64()
			b.McastQueryResponseIntvl = &v
		case unix.IFLA_BR_MCAST_STARTUP_QUERY_INTVL:
			v := ad.Uint64()
			b.McastStartupQueryIntvl = &v
		case unix.IFLA_BR_NF_CALL_IPTABLES:
			v := BridgeEnable(ad.Uint8())
			b.NfCallIptables = &v
		case unix.IFLA_BR_NF_CALL_IP6TABLES:
			v := BridgeEnable(ad.Uint8())
			b.NfCallIp6tables = &v
		case unix.IFLA_BR_NF_CALL_ARPTABLES:
			v := BridgeEnable(ad.Uint8())
			b.NfCallArptables = &v
		case unix.IFLA_BR_VLAN_DEFAULT_PVID:
			v := ad.Uint16()
			b.VlanDefaultPvid = &v
		case unix.IFLA_BR_VLAN_STATS_ENABLED:
			v := BridgeEnable(ad.Uint8())
			b.VlanStatsEnabled = &v
		case unix.IFLA_BR_MCAST_STATS_ENABLED:
			v := BridgeEnable(ad.Uint8())
			b.McastStatsEnabled = &v
		case unix.IFLA_BR_MCAST_IGMP_VERSION:
			v := ad.Uint8()
			b.McastIgmpVersion = &v
		case unix.IFLA_BR_MCAST_MLD_VERSION:
			v := ad.Uint8()
			b.McastMldVersion = &v
		case unix.IFLA_BR_VLAN_STATS_PER_PORT:
			v := BridgeEnable(ad.Uint8())
			b.VlanStatsPerPort = &v
		case unix.IFLA_BR_FDB_MAX_LEARNED:
			v := ad.Uint32()
			b.FdbMaxLearned = &v
		}
	}
	return nil
}

func (*Bridge) Kind() string {
	return "bridge"
}

// BridgePortState specifies the port state in Spanning Tree Protocol
type BridgePortState uint8

const (
	BridgePortStateDisabled BridgePortState = iota
	BridgePortStateListening
	BridgePortStateLearning
	BridgePortStateForwarding
	BridgePortStateBlocking
)

func (s BridgePortState) String() string {
	switch s {
	case BridgePortStateDisabled:
		return "disabled"
	case BridgePortStateListening:
		return "listening"
	case BridgePortStateLearning:
		return "learning"
	case BridgePortStateForwarding:
		return "forwarding"
	case BridgePortStateBlocking:
		return "blocking"
	default:
		return fmt.Sprintf("unknown BridgePortState value (%d)", s)
	}
}

// BridgePort implements LinkSlaveDriver for bridge port/slave configuration
type BridgePort struct {
	// Port state (disabled, listening, learning, forwarding, blocking)
	State *BridgePortState

	// Port priority (default: 32)
	Priority *uint16

	// Port cost (default: automatically calculated from link speed)
	Cost *uint32

	// Hairpin mode - allow packets to be sent back out the port they were received on
	Mode *BridgeEnable

	// BPDU guard - if enabled, STP BPDUs received on the port will cause it to be disabled
	Guard *BridgeEnable

	// Root guard - if enabled, the port will not become a root port
	Protect *BridgeEnable

	// Fast leave - immediately remove the port from multicast group when a leave message is received
	FastLeave *BridgeEnable

	// Learning - controls whether the bridge will learn MAC addresses from packets on this port
	Learning *BridgeEnable

	// Unicast flood - controls flooding of unicast traffic
	UnicastFlood *BridgeEnable

	// Proxy ARP
	ProxyArp *BridgeEnable

	// Learning sync
	LearningSync *BridgeEnable

	// Proxy ARP WiFi
	ProxyArpWifi *BridgeEnable

	// Multicast router (0=disabled, 1=enabled, 2=auto)
	MulticastRouter *uint8

	// Multicast fast leave
	McastFlood *BridgeEnable

	// Multicast to unicast
	McastToUcast *BridgeEnable

	// VLAN tunnel
	VlanTunnel *BridgeEnable

	// Broadcast flood
	BcastFlood *BridgeEnable

	// Group forward mask
	GroupFwdMask *uint16

	// Neighbor suppression
	NeighSuppress *BridgeEnable

	// Isolated - prevents communication between isolated ports
	Isolated *BridgeEnable

	// Backup port index
	BackupPort *uint32

	// Locked - prevents learning of MAC addresses
	Locked *BridgeEnable

	// MAB (MAC Authentication Bypass)
	Mab *BridgeEnable

	// Neighbor VLAN suppression
	NeighVlanSuppress *BridgeEnable

	// Backup nexthop ID
	BackupNhid *uint32
}

var _ rtnetlink.LinkSlaveDriver = &BridgePort{}

func (bp *BridgePort) New() rtnetlink.LinkDriver {
	return &BridgePort{}
}

func (bp *BridgePort) Slave() {}

func (bp *BridgePort) Encode(ae *netlink.AttributeEncoder) error {
	if bp.State != nil {
		ae.Uint8(unix.IFLA_BRPORT_STATE, uint8(*bp.State))
	}
	if bp.Priority != nil {
		ae.Uint16(unix.IFLA_BRPORT_PRIORITY, *bp.Priority)
	}
	if bp.Cost != nil {
		ae.Uint32(unix.IFLA_BRPORT_COST, *bp.Cost)
	}
	if bp.Mode != nil {
		ae.Uint8(unix.IFLA_BRPORT_MODE, uint8(*bp.Mode))
	}
	if bp.Guard != nil {
		ae.Uint8(unix.IFLA_BRPORT_GUARD, uint8(*bp.Guard))
	}
	if bp.Protect != nil {
		ae.Uint8(unix.IFLA_BRPORT_PROTECT, uint8(*bp.Protect))
	}
	if bp.FastLeave != nil {
		ae.Uint8(unix.IFLA_BRPORT_FAST_LEAVE, uint8(*bp.FastLeave))
	}
	if bp.Learning != nil {
		ae.Uint8(unix.IFLA_BRPORT_LEARNING, uint8(*bp.Learning))
	}
	if bp.UnicastFlood != nil {
		ae.Uint8(unix.IFLA_BRPORT_UNICAST_FLOOD, uint8(*bp.UnicastFlood))
	}
	if bp.ProxyArp != nil {
		ae.Uint8(unix.IFLA_BRPORT_PROXYARP, uint8(*bp.ProxyArp))
	}
	if bp.LearningSync != nil {
		ae.Uint8(unix.IFLA_BRPORT_LEARNING_SYNC, uint8(*bp.LearningSync))
	}
	if bp.ProxyArpWifi != nil {
		ae.Uint8(unix.IFLA_BRPORT_PROXYARP_WIFI, uint8(*bp.ProxyArpWifi))
	}
	if bp.MulticastRouter != nil {
		ae.Uint8(unix.IFLA_BRPORT_MULTICAST_ROUTER, *bp.MulticastRouter)
	}
	if bp.McastFlood != nil {
		ae.Uint8(unix.IFLA_BRPORT_MCAST_FLOOD, uint8(*bp.McastFlood))
	}
	if bp.McastToUcast != nil {
		ae.Uint8(unix.IFLA_BRPORT_MCAST_TO_UCAST, uint8(*bp.McastToUcast))
	}
	if bp.VlanTunnel != nil {
		ae.Uint8(unix.IFLA_BRPORT_VLAN_TUNNEL, uint8(*bp.VlanTunnel))
	}
	if bp.BcastFlood != nil {
		ae.Uint8(unix.IFLA_BRPORT_BCAST_FLOOD, uint8(*bp.BcastFlood))
	}
	if bp.GroupFwdMask != nil {
		ae.Uint16(unix.IFLA_BRPORT_GROUP_FWD_MASK, *bp.GroupFwdMask)
	}
	if bp.NeighSuppress != nil {
		ae.Uint8(unix.IFLA_BRPORT_NEIGH_SUPPRESS, uint8(*bp.NeighSuppress))
	}
	if bp.Isolated != nil {
		ae.Uint8(unix.IFLA_BRPORT_ISOLATED, uint8(*bp.Isolated))
	}
	if bp.BackupPort != nil {
		ae.Uint32(unix.IFLA_BRPORT_BACKUP_PORT, *bp.BackupPort)
	}
	if bp.Locked != nil {
		ae.Uint8(unix.IFLA_BRPORT_LOCKED, uint8(*bp.Locked))
	}
	if bp.Mab != nil {
		ae.Uint8(unix.IFLA_BRPORT_MAB, uint8(*bp.Mab))
	}
	if bp.NeighVlanSuppress != nil {
		ae.Uint8(unix.IFLA_BRPORT_NEIGH_VLAN_SUPPRESS, uint8(*bp.NeighVlanSuppress))
	}
	if bp.BackupNhid != nil {
		ae.Uint32(unix.IFLA_BRPORT_BACKUP_NHID, *bp.BackupNhid)
	}

	return nil
}

func (bp *BridgePort) Decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.IFLA_BRPORT_STATE:
			v := BridgePortState(ad.Uint8())
			bp.State = &v
		case unix.IFLA_BRPORT_PRIORITY:
			v := ad.Uint16()
			bp.Priority = &v
		case unix.IFLA_BRPORT_COST:
			v := ad.Uint32()
			bp.Cost = &v
		case unix.IFLA_BRPORT_MODE:
			v := BridgeEnable(ad.Uint8())
			bp.Mode = &v
		case unix.IFLA_BRPORT_GUARD:
			v := BridgeEnable(ad.Uint8())
			bp.Guard = &v
		case unix.IFLA_BRPORT_PROTECT:
			v := BridgeEnable(ad.Uint8())
			bp.Protect = &v
		case unix.IFLA_BRPORT_FAST_LEAVE:
			v := BridgeEnable(ad.Uint8())
			bp.FastLeave = &v
		case unix.IFLA_BRPORT_LEARNING:
			v := BridgeEnable(ad.Uint8())
			bp.Learning = &v
		case unix.IFLA_BRPORT_UNICAST_FLOOD:
			v := BridgeEnable(ad.Uint8())
			bp.UnicastFlood = &v
		case unix.IFLA_BRPORT_PROXYARP:
			v := BridgeEnable(ad.Uint8())
			bp.ProxyArp = &v
		case unix.IFLA_BRPORT_LEARNING_SYNC:
			v := BridgeEnable(ad.Uint8())
			bp.LearningSync = &v
		case unix.IFLA_BRPORT_PROXYARP_WIFI:
			v := BridgeEnable(ad.Uint8())
			bp.ProxyArpWifi = &v
		case unix.IFLA_BRPORT_MULTICAST_ROUTER:
			v := ad.Uint8()
			bp.MulticastRouter = &v
		case unix.IFLA_BRPORT_MCAST_FLOOD:
			v := BridgeEnable(ad.Uint8())
			bp.McastFlood = &v
		case unix.IFLA_BRPORT_MCAST_TO_UCAST:
			v := BridgeEnable(ad.Uint8())
			bp.McastToUcast = &v
		case unix.IFLA_BRPORT_VLAN_TUNNEL:
			v := BridgeEnable(ad.Uint8())
			bp.VlanTunnel = &v
		case unix.IFLA_BRPORT_BCAST_FLOOD:
			v := BridgeEnable(ad.Uint8())
			bp.BcastFlood = &v
		case unix.IFLA_BRPORT_GROUP_FWD_MASK:
			v := ad.Uint16()
			bp.GroupFwdMask = &v
		case unix.IFLA_BRPORT_NEIGH_SUPPRESS:
			v := BridgeEnable(ad.Uint8())
			bp.NeighSuppress = &v
		case unix.IFLA_BRPORT_ISOLATED:
			v := BridgeEnable(ad.Uint8())
			bp.Isolated = &v
		case unix.IFLA_BRPORT_BACKUP_PORT:
			v := ad.Uint32()
			bp.BackupPort = &v
		case unix.IFLA_BRPORT_LOCKED:
			v := BridgeEnable(ad.Uint8())
			bp.Locked = &v
		case unix.IFLA_BRPORT_MAB:
			v := BridgeEnable(ad.Uint8())
			bp.Mab = &v
		case unix.IFLA_BRPORT_NEIGH_VLAN_SUPPRESS:
			v := BridgeEnable(ad.Uint8())
			bp.NeighVlanSuppress = &v
		case unix.IFLA_BRPORT_BACKUP_NHID:
			v := ad.Uint32()
			bp.BackupNhid = &v
		}
	}
	return nil
}

func (*BridgePort) Kind() string {
	return "bridge"
}
