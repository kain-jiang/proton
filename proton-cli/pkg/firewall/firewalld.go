// Package firewall 提供防火墙管理功能
package firewall

import (
	"net/netip"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"

	firewalld "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/firewalld/v1alpha1"
	node "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	systemd "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/systemd/v1alpha1"
)

// 防火墙相关常量定义
const (
	// systemd unit: firewalld.service
	systemdUnitFirewalld = "firewalld.service"

	firewalldSourcePrefixIPSet = "ipset:"

	// firewalldZoneProtonCS 定义 Proton 集群服务使用的防火墙区域名称
	firewalldZoneProtonCS = "proton-cs"
	// firewalldIPSetProtonCSHost 定义 IPv4 主机内部 IP 集合名称
	firewalldIPSetProtonCSHost = "proton-cs-host"
	// firewalldIPSetProtonCSHost6 定义 IPv6 主机内部 IP 集合名称
	firewalldIPSetProtonCSHost6 = "proton-cs-host6"

	// firewalldInterfaceTunl0 定义 tunnel 接口名称
	firewalldInterfaceTunl0 = "tunl0"
	// firewalldInterfaceCaliPlus 定义 calico 网络接口名称模式
	firewalldInterfaceCaliPlus = "cali+"
)

// firewalldSpec 定义firewalld配置规格
// 包含IP集合和区域配置

type firewalldSpec struct {
	// IPSets 定义 IP 集合列表
	IPSets []firewalldIPSet
	// Zones 定义防火墙区域列表
	Zones []firewalldZone
}

// firewalldZone 定义防火墙区域配置

type firewalldZone struct {
	// Name 区域名称
	Name string
	// Target 区域默认目标策略
	Target firewalld.Target
	// Interfaces 关联的网络接口列表
	Interfaces []string
	// Sources 关联的源列表
	Sources []string
}

// firewalldIPSet 定义IP集合配置

type firewalldIPSet struct {
	// Name IPSet 名称
	Name string
	// Family IPSet 地址族类型(IPv4/IPv6)
	Family firewalld.IPSetFamily
	// Type IPSet 类型
	Type firewalld.IPSetType
	// Entries IPSet 的条目
	Entries []string
}

// moduleFirewalld 实现了 firewall.Interface 接口的 firewalld 模块

type moduleFirewalld struct {
	// 节点客户端接口列表
	nodes []node.Interface
	// 允许来自这些地址的请求
	addresses []netip.Addr
	// 允许来自这些网络的请求
	nets []string
	// logger 日志记录器
	logger *logrus.Logger
}

// Apply 实现Interface.Apply方法，应用firewalld配置
// 通过并发方式为所有客户端应用相同的防火墙配置
// 返回执行过程中的错误，如果全部成功则返回nil
func (m *moduleFirewalld) Apply() error {
	// 定义 IPSet 配置
	//  1. proton-cs-host: 用于存储 IPv4 地址
	//  2. proton-cs-host6: 用于存储 IPv6 地址
	var ipSets = []firewalldIPSet{
		{
			Name:    firewalldIPSetProtonCSHost,
			Family:  firewalld.IPSetFamilyINet, // IPv4
			Type:    firewalld.IPSetType{Method: firewalld.IPSetMethodHash, DataTypes: []firewalld.IPSetDataType{firewalld.IPSetDataTypeIP}},
			Entries: lo.FilterMap(m.addresses, func(ip netip.Addr, _ int) (string, bool) { return ip.String(), ip.Is4() }), // 过滤出 IPv4 地址
		},
		{
			Name:    firewalldIPSetProtonCSHost6,
			Family:  firewalld.IPSetFamilyINet6, // IPv6
			Type:    firewalld.IPSetType{Method: firewalld.IPSetMethodHash, DataTypes: []firewalld.IPSetDataType{firewalld.IPSetDataTypeIP}},
			Entries: lo.FilterMap(m.addresses, func(ip netip.Addr, _ int) (string, bool) { return ip.String(), ip.Is6() }), // 过滤出 IPv6 地址
		},
	}

	m.logger.WithField("addresses", m.addresses).Debug("DEBUGGING")
	var sources []string
	sources = append(sources, firewalldSourcePrefixIPSet+firewalldIPSetProtonCSHost)
	sources = append(sources, firewalldSourcePrefixIPSet+firewalldIPSetProtonCSHost6)
	sources = append(sources, m.nets...)
	// 定义区域配置
	//  1. proton-cs: Proton 集群服务专用区域
	var zones = []firewalldZone{
		{
			Name:       firewalldZoneProtonCS,
			Target:     firewalld.TargetAccept,
			Interfaces: []string{firewalldInterfaceCaliPlus, firewalldInterfaceTunl0},
			Sources:    sources,
		},
	}

	g := new(errgroup.Group)

	// 并发修改所以节点的 firewalld 配置
	for _, n := range m.nodes {
		g.Go(func() error { return ensureNodeFirewalld(n, &firewalldSpec{IPSets: ipSets, Zones: zones}) })
	}

	return g.Wait() // 等待所有并发任务完成
}

// 确保moduleFirewalld实现了Interface接口
var _ Interface = &moduleFirewalld{}

// ensureNodeFirewalld 确保节点的firewalld配置符合预期
// 参数:
//
//	c: firewalld 客户端接口
//	spec: 期望的 firewalld 配置规格
//
// 返回执行过程中的错误
func ensureNodeFirewalld(c node.Interface, spec *firewalldSpec) error {
	// start and enable firewalld.service
	if err := ensureNodeSystemdUnitEnabledAndActive(c.Systemd(), systemdUnitFirewalld); err != nil {
		return err
	}

	// 先应用持久化配置：true表示持久化
	if err := ensureNodeFirewalldIPSets(c.Firewalld(), true, spec.IPSets); err != nil {
		return err
	}
	if err := ensureNodeFirewalldZones(c.Firewalld(), true, spec.Zones); err != nil {
		return err
	}

	// firewalld 运行中，同时应用运行时配置：false 表示运行时
	if err := ensureNodeFirewalldIPSets(c.Firewalld(), false, spec.IPSets); err != nil {
		return err
	}
	if err := ensureNodeFirewalldZones(c.Firewalld(), false, spec.Zones); err != nil {
		return err
	}

	return nil
}

// ensureNodeSystemdUnitEnabledAndActive ensure node systemd unit is enabled and active
func ensureNodeSystemdUnitEnabledAndActive(c systemd.Interface, u string) error {
	active, err := c.IsActive(u)
	if err != nil {
		return err
	}

	enabled, err := c.IsEnabled(u)
	if err != nil {
		return err
	}

	// 只要不是 enabled 就可以 systemctl enable --now 即使已经是 active 也没问题
	if !enabled {
		return c.Enabled(u, true)
	}

	if !active {
		return c.Start(u)
	}

	return nil
}

// ensureNodeFirewalldIPSets 确保 IPSet 配置符合预期
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	ipsets: 期望的 IP 集合列表
//
// 返回执行过程中的错误
func ensureNodeFirewalldIPSets(c firewalld.Interface, permanent bool, ipsets []firewalldIPSet) error {
	// 确保所有需要的 IPSet 都存在，如果不存在则创建
	if err := ensureNodeFirewalldIPSetsOnly(c, permanent, ipsets); err != nil {
		return err
	}

	// 更新每个 IPSet 的条目为期望值
	for _, s := range ipsets {
		if err := ensureNodeFirewalldIPSetEntries(c, permanent, s.Name, s.Entries); err != nil {
			return err
		}
	}

	return nil
}

// ensureNodeFirewalldIPSetsOnly 确保指定的 IPSet 存在，不处理 IPSet 的条目内容
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	ipsets: 期望存在的 IPSet 列表
//
// 返回执行过程中的错误
func ensureNodeFirewalldIPSetsOnly(c firewalld.Interface, permanent bool, ipsets []firewalldIPSet) error {
	// 获取已经存在的 IPSet
	actual, err := c.GetIPSets(permanent)
	if err != nil {
		return err
	}

	// 找出缺失的 IPSet
	missing := lo.WithoutBy(ipsets, func(z firewalldIPSet) string { return z.Name }, actual...)

	// 如果没有缺失的 IPSet，直接返回
	if len(missing) == 0 {
		return nil
	}

	// 对于运行时配置，如果有缺失，直接重新加载
	if !permanent {
		return c.Reload()
	}

	// 创建缺失的 IPSet
	for _, s := range missing {
		if err := c.NewIPSet(s.Name, s.Type, s.Family); err != nil {
			return err
		}
	}

	return nil
}

// ensureNodeFirewalldIPSetEntries 确保 IPSet 的条目符合预期
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	ipset: IPSet 名称
//	entries: 期望的 IPSet 条目列表
//
// 返回执行过程中的错误
func ensureNodeFirewalldIPSetEntries(c firewalld.Interface, permanent bool, ipset string, entries []string) error {
	// 获取已经存在的条目
	actual, err := c.GetIPSetEntries(permanent, ipset)
	if err != nil {
		return err
	}

	// 添加缺失的条目
	if missing := sets.New(entries...).Delete(actual...); missing.Len() != 0 {
		if err := c.AddIPSetEntries(permanent, ipset, sets.List(missing)); err != nil {
			return err
		}
	}

	// 移除多余的条目
	if redundant := sets.New(actual...).Delete(entries...); redundant.Len() != 0 {
		if err := c.RemoveIPSetEntries(permanent, ipset, sets.List(redundant)); err != nil {
			return err
		}
	}

	return nil
}

// ensureNodeFirewalldZones 确保防火墙区域配置符合预期
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	zones: 期望的防火墙区域列表
//
// 返回执行过程中的错误
func ensureNodeFirewalldZones(c firewalld.Interface, permanent bool, zones []firewalldZone) error {
	// 确保所有需要的区域都存在，如果不存在则创建
	if err := ensureNodeFirewalldZonesOnly(c, permanent, zones); err != nil {
		return err
	}

	// 更新每个区域的详细配置为期望值
	for _, z := range zones {
		if err := ensureNodeFirewalldZone(c, permanent, &z); err != nil {
			return err
		}
	}

	return nil
}

// ensureNodeFirewalldZonesOnly 确保指定的防火墙区域存在，不处理区域的具体配置
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	zones: 期望存在的防火墙区域列表
//
// 返回执行过程中的错误
func ensureNodeFirewalldZonesOnly(c firewalld.Interface, permanent bool, zones []firewalldZone) error {
	// 获取已经存在的区域
	actual, err := c.GetZones(permanent)
	if err != nil {
		return err
	}

	// 找出缺失的区域
	missing := lo.WithoutBy(zones, func(z firewalldZone) string { return z.Name }, actual...)

	// 如果没有缺失的区域，直接返回
	if len(missing) == 0 {
		return nil
	}

	// 对于运行时配置，如果有缺失，直接重新加载
	if !permanent {
		return c.Reload()
	}

	// 创建缺失的区域
	for _, z := range missing {
		if err := c.NewZone(z.Name); err != nil {
			return err
		}
	}

	return nil
}

// ensureNodeFirewalldZone 确保单个防火墙区域的完整配置符合预期
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	zone: 期望的防火墙区域配置
//
// 返回执行过程中的错误
func ensureNodeFirewalldZone(c firewalld.Interface, permanent bool, zone *firewalldZone) error {
	// 确保区域的默认目标策略正确
	if err := ensureNodeFirewalldZoneTarget(c, permanent, zone.Name, zone.Target); err != nil {
		return err
	}
	// 确保区域关联的网络接口正确
	if err := ensureNodeFirewalldZoneInterfaces(c, permanent, zone.Name, zone.Interfaces); err != nil {
		return err
	}
	// 确保区域关联的源正确
	if err := ensureNodeFirewalldZoneSources(c, permanent, zone.Name, zone.Sources); err != nil {
		return err
	}
	return nil
}

// ensureNodeFirewalldZoneTarget 确保防火墙区域的默认目标策略符合预期
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	zone: 区域名称
//	target: 期望的默认目标策略
//
// 返回执行过程中的错误
func ensureNodeFirewalldZoneTarget(c firewalld.Interface, permanent bool, zone string, target firewalld.Target) error {
	// 获取当前的默认目标策略
	got, err := c.GetZoneTarget(zone)
	if err != nil {
		return err
	}
	// 如果当前配置已经符合预期，直接返回
	if got == target {
		return nil
	}
	// 对于运行时配置，直接重新加载
	if !permanent {
		return c.Reload()
	}
	// 设置新的默认目标策略
	return c.SetZoneTarget(zone, target)
}

// ensureNodeFirewalldZoneInterfaces 确保防火墙区域关联的网络接口符合预期
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	zone: 区域名称
//	interfaces: 期望关联的网络接口列表
//
// 返回执行过程中的错误
func ensureNodeFirewalldZoneInterfaces(c firewalld.Interface, permanent bool, zone string, interfaces []string) error {
	// 获取当前关联的网络接口
	actual, err := c.ListZoneInterfaces(permanent, zone)
	if err != nil {
		return nil // 注意：这里遇到错误直接返回nil，可能需要根据实际情况修改
	}

	// 添加缺失的网络接口
	//
	// 支持 openEuler，需要允许设备 tunl0 的网络包，否则不同节点的 Pod 无法互相通信。
	if missing := sets.New(interfaces...).Delete(actual...); missing.Len() != 0 {
		if err := c.AddZoneInterfaces(permanent, zone, sets.List(missing)); err != nil {
			return err
		}
	}

	// 移除多余的网络接口
	if redundant := sets.New(actual...).Delete(interfaces...); redundant.Len() != 0 {
		if err := c.RemoveZoneInterfaces(permanent, zone, sets.List(redundant)); err != nil {
			return err
		}
	}

	return nil
}

// ensureNodeFirewalldZoneSources 确保防火墙区域关联的源IP/IP集合符合预期
// 参数:
//
//	c: firewalld 客户端接口
//	permanent: 是否为持久化配置
//	zone: 区域名称
//	sources: 期望关联的源 IP/IP 集合列表
//
// 返回执行过程中的错误
func ensureNodeFirewalldZoneSources(c firewalld.Interface, permanent bool, zone string, sources []string) error {
	// 获取当前关联的 source
	actual, err := c.ListZoneSources(permanent, zone)
	if err != nil {
		return nil // 注意：这里遇到错误直接返回nil，可能需要根据实际情况修改
	}

	// 添加缺失的 source
	if missing := sets.New(sources...).Delete(actual...); missing.Len() != 0 {
		if err := c.AddZoneSources(permanent, zone, sets.List(missing)); err != nil {
			return err
		}
	}

	// 移除多余的 source
	if redundant := sets.New(actual...).Delete(sources...); redundant.Len() != 0 {
		if err := c.RemoveZoneSources(permanent, firewalldZoneProtonCS, sets.List(redundant)); err != nil {
			return err
		}
	}

	return nil
}
