package v2

import (
	"encoding/json"
)

type KeepalivedHA struct {

	//  interface for inside_network, bound by vrrp.
	//
	// Note: if using unicasting, the interface can be omitted as long as the
	// unicast addresses are not IPv6 link local addresses (this is necessary,
	// for example, if using asymmetric routing).
	//
	// If the interface is omitted, then all VIPs and eVIPs should specify the
	// interface they are to be configured on, otherwise they will be added to
	// the default interface.
	Interface string `json:"interface,omitempty"`

	// interface for inside_network, bound by vrrp.
	//
	// Note: if using unicasting, the interface can be omitted as long as the
	// unicast addresses are not IPv6 link local addresses (this is necessary,
	// for example, if using asymmetric routing).
	//
	// If the interface is omitted, then all VIPs and eVIPs should specify the
	// interface they are to be configured on, otherwise they will be added to
	// the default interface.
	UnicastSRC_IP string `json:"unicast_src_ip,omitempty"`

	// Do not send VRRP adverts over a VRRP multicast group.
	//
	// Instead it sends adverts to the following list of ip addresses using
	// unicast. It can be cool to use the VRRP FSM and features in networking
	// environment where multicast is not supported!
	//
	// IP addresses specified can be IPv4 as well as IPv6.
	//
	// If min_ttl and/or max_ttl are specified, the TTL/hop limit of any
	// received packet is checked against the specified TTL range, and is
	// discarded if it is outside the range.
	//
	// Specifying min_ttl or max_ttl turns on check_unicast_src.
	UnicastPeer unicastPeer `json:"unicast_peer,omitempty"`

	// arbitrary unique number from 1 to 255
	//
	// used to differentiate multiple instances of vrrpd running on the same
	// network interface and address family and multicast/unicast (and hence
	// same socket).
	//
	// Note: using the same virtual_router_id with the same address family on
	// different interfaces has been known to cause problems with some network
	// switches; if you are experiencing problems with using the same
	// virtual_router_id on different interfaces, but the problems are resolved
	// by not duplicating virtual_router_ids, your network switches are probably
	// not functioning correctly.
	//
	// Whilst in general it is important not to duplicate a virtual_router_id on
	// the same network interface, there is a special case when using unicasting
	// if the unicast peers for the vrrp instances with duplicated
	// virtual_router_ids on the network interface do not overlap, in which case
	// virtual_router_ids can be duplicated.
	//
	// It is also possible to duplicate virtual_router_ids on an interface with
	// multicasting if different multicast addresses are used (see
	// mcast_dst_ip).
	VirtualRouterID string `json:"virtual_router_id,omitempty"`

	// for electing MASTER, highest priority wins.
	//
	// The valid range of values for priority is [1-255], with priority255
	// meaning "address owner".
	//
	// To be MASTER, it is recommended to make this 50 more than on other
	// machines. All systems should have different priorities in order to make
	// behaviour deterministic. If you want to stop a higher priority instance
	// taking over as master when it starts, configure no_preempt rather than
	// using equal priorities.
	//
	// If no_accept is configured (or vrrp_strict # which also sets no_accept
	// mode), then unless the vrrp_instance has priority 255, the system will
	// not receive packets addressed to the # VIPs/eVIPs, and the VIPs/eVIPs can
	// only be used for routeing purposes.
	//
	// Further, if an instance has priority 255 configured, the priority cannot
	// be reduced by track_scripts, track_process etc, and likewise
	// track_scripts etc cannot increase the priority to 255 if the configured
	// priority is not 255.
	Priority string `json:"priority,omitempty"`

	// addresses add|del on change to MASTER, to BACKUP.
	//
	// With the same entries on other machines, the opposite transition will be
	// occurring.
	//
	// For virtual_ipaddress, virtual_ipaddress_excluded, virtual_routes and
	// virtual_rules most of the options match the options of the command ip
	// address/route/rule add.  The track_group option only applies to static
	// addresses/routes/rules.  no_track is specific to keepalived and means
	// that the vrrp_instance will not transition out of master state if the
	// address/route/rule is deleted and the address/route/rule will not be
	// reinstated until the vrrp instance next transitions to master.
	//
	//  <LABEL>: is optional and creates a name for the alias.
	//           For compatibility with "ifconfig", it should
	//           be of the form <realdev>:<anytext>, for example
	//           eth0:1 for an alias on eth0.
	//  <SCOPE>: ("site"|"link"|"host"|"nowhere"|"global")
	//
	// preferred_lft is set to 0 to deprecate IPv6 addresses (this is the
	// default if the address mask is /128). Use "preferred_lft forever" to
	// specify that a /128 address should not be deprecated.
	//
	// NOTE: care needs to be taken if dev is specified for an address and your
	// network uses MAC learning switches. The VRRP protocol ensures that the
	// source MAC address of the interface sending adverts is maintained in the
	// MAC cache of switches; however by default this will not work for the MACs
	// of any VIPs/eVIPs that are configured on different interfaces from the
	// interface on which the VRRP instance is configured, since the interface,
	// especially if it is a VMAC interface, will only send using the MAC
	// address of the interface in response to ARP requests. This may mean that
	// the interface MAC addresses may time out in the MAC caches of switches.
	// In order to avoid this, use the garp_extra_if or garp_extra_if_vmac
	// options to send periodic GARP/ND messages on those interfaces.
	VirtualIPAddress map[string]string `json:"virtual_ipaddress,omitempty"`

	// notify scripts, alert as above
	NotifyMaster string `json:"notify_master,omitempty"`

	// notify scripts, alert as above
	NotifyBackup string `json:"notify_backup,omitempty"`
}

// The definition of the data structure returned by the interface is []string,
// but it is actually map[string]string. So compatible
type unicastPeer []string

func (up *unicastPeer) UnmarshalJSON(b []byte) error {
	var s []string
	if json.Unmarshal(b, &s) != nil {
		var m map[string]string
		if err := json.Unmarshal(b, &m); err != nil {
			return err
		}
		for k := range m {
			s = append(s, k)
		}
	}

	*up = make(unicastPeer, len(s))
	copy(*up, s)

	return nil
}

const (
	KeepalivedHAStateMaster = "MASTER"
	KeepalivedHAStateBackup = "BACKUP"
)
