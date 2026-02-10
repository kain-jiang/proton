package eceph

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	node "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	slb "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func (m *Manager) reconcileKeepalivedHAInstances() error {
	if m.Spec.Keepalived == nil || (m.Spec.Keepalived.External == "" && m.Spec.Keepalived.Internal == "") {
		m.Logger.Debug("eceph.keepalived undefined")
		return nil
	}

	err := m.reconcileFirewalldVirtualIPs()
	if err != nil {
		return err
	}

	var internalPeers, externalPeers []net.IP
	for _, n := range m.Nodes {
		internalPeers = append(internalPeers, n.InternalIP())
		externalPeers = append(externalPeers, n.IP())
	}

	for _, n := range m.Nodes {
		log := m.Logger.WithField("node", n.Name())

		internalName, externalName := generateKeepalivedHAInstanceNames()
		internalVirtualRouterID, externalVirtualRouterID := generateKeepalivedHAInstanceIDs()
		internalLabel, externalLabel := generateKeepalivedLabels()
		internalNotifyMaster, externalNotifyMaster := generateKeepalivedNotifyMasters()
		internalNotifyBackup, externalNotifyBackup := generateKeepalivedNotifyBackups()

		networkInterfaces, err := n.NetworkInterfaces()
		if err != nil {
			return err
		}

		if len(m.Spec.Keepalived.Internal) > 0 {
			_, _, err = net.ParseCIDR(m.Spec.Keepalived.Internal)
			if err != nil {
				return err
			}
		}

		for _, s := range []struct {
			name            string
			virtualRouterID int
			vip             string
			dev             string
			label           string
			src             net.IP
			peers           []net.IP
			notifyMaster    string
			notifyBackup    string
		}{
			{
				name:            internalName,
				virtualRouterID: internalVirtualRouterID,
				vip:             m.Spec.Keepalived.Internal,
				dev:             getNetworkInterfaceNameByIP(n.InternalIP(), networkInterfaces),
				label:           internalLabel,
				src:             n.InternalIP(),
				peers:           internalPeers,
				notifyMaster:    internalNotifyMaster,
				notifyBackup:    internalNotifyBackup,
			},
			{
				name:            externalName,
				virtualRouterID: externalVirtualRouterID,
				vip:             m.Spec.Keepalived.External,
				dev:             getNetworkInterfaceNameByIP(n.IP(), networkInterfaces),
				label:           externalLabel,
				src:             n.IP(),
				peers:           externalPeers,
				notifyMaster:    externalNotifyMaster,
				notifyBackup:    externalNotifyBackup,
			},
		} {
			if len(s.vip) == 0 {
				continue
			}
			instance := generateKeepalivedHAInstance(s.name, s.virtualRouterID, s.vip, s.dev, s.label, s.src, s.peers, s.notifyMaster, s.notifyBackup, m.Nodes[0].IPVersion())
			if err := reconcileNodeKeepalivedHAInstance(n.SLB_V2().KeepalivedHAs(), s.name, instance, log); err != nil {
				// retry once without notifyMaster and notifyBackup if it somehow exists
				errStr := fmt.Sprintf("%v", err)
				log.Debugf("error string: %s", errStr)
				flagErrDuetoExistingNotify := false
				if strings.Contains(errStr, fmt.Sprintf("(%s) notify_master script already specified - ignoring %s", s.name, instance.NotifyMaster)) {
					log.Debugln("Notify master script already exists on this HA Instance, will retry without setting it")
					instance.NotifyMaster = ""
					flagErrDuetoExistingNotify = true
				}
				if strings.Contains(errStr, fmt.Sprintf("(%s) notify_backup script already specified - ignoring %s", s.name, instance.NotifyBackup)) {
					log.Debugln("Notify backup script already exists on this HA Instance, will retry without setting it")
					instance.NotifyBackup = ""
					flagErrDuetoExistingNotify = true
				}
				if flagErrDuetoExistingNotify {
					err1 := reconcileNodeKeepalivedHAInstance(n.SLB_V2().KeepalivedHAs(), s.name, instance, log)
					if err1 != nil {
						return err1
					}
				} else {
					return err
				}
			}
		}
	}
	return nil
}

func (m *Manager) resetKeepalivedHAInstances() error {
	// The reset process would not automatically remove ECeph related Proton SLB Keepalived HA instances
	// If you encounter any problem regarding K8S ETCD not starting or kubeadm waiting indefinitely during the next proton-cli apply
	// Please try deleting the following Proton SLB Keepalived HA instance manually:", KeepalivedNameECephExternalVIP, KeepalivedNameECephInternalVIP, KeepalivedNameECephSLB_HA
	return nil
}

func generateKeepalivedHAInstanceNames() (internal, external string) {
	internal, external = KeepalivedNameECephInternalVIP, KeepalivedNameECephExternalVIP
	return
}

func generateKeepalivedHAInstanceIDs() (internal, external int) {
	internal, external = KeepalivedVirtualRouterIDECephInternalVIP, KeepalivedVirtualRouterIDECephExternalVIP
	return
}

func generateKeepalivedLabels() (internal, external string) {
	internal, external = KeepalivedVirtualAddressLabelECephInternalVIP, KeepalivedVirtualAddressLabelECephExternalVIP
	return
}

func generateKeepalivedNotifyMasters() (internal, external string) {
	internal, external = "", KeepalivedNotifyMasterECephExternalVIP
	return
}

func generateKeepalivedNotifyBackups() (internal, external string) {
	internal, external = "", KeepalivedNotifyBackupECephExternalVIP
	return
}

func generateKeepalivedHAInstance(name string, virtualRouterID int, vip string, dev, label string, unicastSRC_IP net.IP, unicastPeers []net.IP, notifyMaster, notifyBackup string, ipVersion string) *slb.KeepalivedHA {
	var peers []string
	for _, p := range unicastPeers {
		peers = append(peers, p.String())
	}
	var nm, nb string
	if a := strings.Count(notifyMaster, "%s"); a == 0 {
		nm = notifyMaster
	} else if a == 1 {
		nm = fmt.Sprintf(notifyMaster, name)
	} else {
		panic("notifyMaster contains more than 1 format directives, this should never happen")
	}
	if a := strings.Count(notifyBackup, "%s"); a == 0 {
		nb = notifyBackup
	} else if a == 1 {
		nb = fmt.Sprintf(notifyBackup, name)
	} else {
		panic("notifyBackup contains more than 1 format directives, this should never happen")
	}

	var vipaddress map[string]string
	if ipVersion == configuration.IPVersionIPV4 {
		vipaddress = generateVirtualIPAddress(vip, dev, label)
	} else if ipVersion == configuration.IPVersionIPV6 {
		vipaddress = generateVirtualIPAddressV6(vip, dev)
	} else {
		panic("invalid ipVersion, this should never happen")
	}

	return &slb.KeepalivedHA{
		Interface:        dev,
		UnicastSRC_IP:    unicastSRC_IP.String(),
		UnicastPeer:      peers,
		VirtualRouterID:  strconv.Itoa(virtualRouterID),
		Priority:         generateKeepalivedPriorityFromIP(unicastSRC_IP),
		VirtualIPAddress: vipaddress,
		NotifyMaster:     nm,
		NotifyBackup:     nb,
	}
}

// Returns the last 8 bits of the IP address as a decimal string.
// will return another random number if the priority is 30, 50 or 100 due to how nodemanagement works
func generateKeepalivedPriorityFromIP(ip net.IP) string {
	// 先取当前输入IP的最后一位
	result := int(ip[len(ip)-1])
	return strconv.Itoa(result)
}

func generateVirtualIPAddress(vip string, dev, label string) map[string]string {
	return map[string]string{
		vip: fmt.Sprintf("label %s:%s dev %s", dev, label, dev),
	}
}

func generateVirtualIPAddressV6(vip string, dev string) map[string]string {
	return map[string]string{
		vip: fmt.Sprintf("dev %s", dev),
	}
}

func getNetworkInterfaceNameByIP(ip net.IP, interfaces []node.NetworkInterface) string {
	for _, ifi := range interfaces {
		for _, a := range ifi.Addresses {
			if a.Contains(ip) {
				return ifi.Name
			}
		}
	}
	return "unknown"
}

func reconcileNodeKeepalivedHAInstance(c slb.KeepalivedHAInterface, name string, instance *slb.KeepalivedHA, log logrus.FieldLogger) error {
	names, err := c.List(context.TODO())
	if err != nil {
		return err
	}
	if !slices.Contains(names, name) {
		log.WithFields(logrus.Fields{"name": name, "instance": instance}).Info("create proton slb keepalived ha instance")
		return c.Create(context.TODO(), name, instance)
	}

	log.WithFields(logrus.Fields{"name": name, "instance": instance}).Info("update proton slb keepalived ha instance")
	return c.Update(context.TODO(), name, instance)
}
