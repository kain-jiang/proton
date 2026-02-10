package v1alpha1

import (
	"bytes"
)

const (
	ZoneProtonCS = "proton-cs"
)

type IPSetFamily string

const (
	IPSetFamilyINet  IPSetFamily = "inet"
	IPSetFamilyINet6 IPSetFamily = "inet6"
)

type Target string

const (
	TargetDefault Target = "default"
	TargetAccept  Target = "ACCEPT"
	TargetDrop    Target = "DROP"
	TargetReject  Target = "REJECT"
)

type IPSetType struct {
	Method    IPSetMethod
	DataTypes []IPSetDataType
}

func (t IPSetType) String() string {
	var buf bytes.Buffer
	buf.WriteString(string(t.Method))
	buf.WriteRune(':')
	for i, tt := range t.DataTypes {
		if i != 0 {
			buf.WriteRune(',')
		}
		buf.WriteString(string(tt))
	}
	return buf.String()
}

type IPSetMethod string

const (
	IPSetMethodBitmap IPSetMethod = "bitmap"
	IPSetMethodHash   IPSetMethod = "hash"
	IPSetMethodList   IPSetMethod = "list"
)

type IPSetDataType string

const (
	IPSetDataTypeIP        IPSetDataType = "ip"
	IPSetDataTypeNet       IPSetDataType = "net"
	IPSetDataTypeMAC       IPSetDataType = "mac"
	IPSetDataTypePort      IPSetDataType = "port"
	IPSetDataTypeInterface IPSetDataType = "iface"
)

type Interface interface {
	IsRunning() (bool, error)
	//  Add the port for a zone [P] [Z] [T]
	AddPort(port string, permanent bool, zone string, timeout string) error
	// List ports added for a zone [P] [Z]
	ListPorts(permanent bool, zone string) ([]string, error)
	// Add an IP to firewalld trusted source
	RemoveAndAddFirewallSource(ip string, oldIP string, permanent bool) error

	// Check whether the firewalld daemon is active.
	State() (bool, error)

	// Reload firewall rules and keep state information. Current permanent
	// configuration will become new runtime configuration, i.e. all runtime
	// only changes done until reload are lost with reload if they have not been
	// also in permanent configuration.
	Reload() error

	// Add a new permanent and empty zone.
	NewZone(zone string) error
	// Delete an existing permanent zone.
	DeleteZone(zone string) error
	// Get predefined zones.
	GetZones(permanent bool) ([]string, error)

	// Set the zone's target.
	SetZoneTarget(zone string, target Target) error
	// Get the zone's target.
	GetZoneTarget(zone string) (Target, error)

	// Add a new permanent and empty ipset with specifying the type and optional
	// the family.
	NewIPSet(ipset string, ipsetType IPSetType, family IPSetFamily) error
	// Delete an existing permanent ipset.
	DeleteIPSet(ipset string) error
	// Get ipset list.
	GetIPSets(permanent bool) ([]string, error)

	// Bind the source to zone. If zone is omitted, default zone will be used.
	AddZoneSources(permanent bool, zone string, sources []string) error
	// Remove binding of interface from zone it was previously added to.
	RemoveZoneSources(permanent bool, zone string, sources []string) error
	// List sources that are bound to zone. If zone is omitted, default zone
	// will be used.
	ListZoneSources(permanent bool, zone string) ([]string, error)

	// Bind interface interface to zone. If zone is omitted, default zone will
	// be used.
	AddZoneInterfaces(permanent bool, zone string, interfaces []string) error
	// Change zone the interface is bound to.
	ChangeZoneInterfaces(permanent bool, zone string, interfaces []string) error
	// Remove binding of interface from zone it was previously added to.
	RemoveZoneInterfaces(permanent bool, zone string, interfaces []string) error
	//List interfaces that are bound to zone as a space separated list. If zone
	//is omitted, default zone will be used.
	ListZoneInterfaces(permanent bool, zone string) ([]string, error)

	// Add new entries to the ipset.
	AddIPSetEntries(permanent bool, ipset string, entries []string) error
	// Remove entries from the ipset.
	RemoveIPSetEntries(permanent bool, ipset string, entries []string) error
	// List all entries of the ipset.
	GetIPSetEntries(permanent bool, ipset string) ([]string, error)
}
