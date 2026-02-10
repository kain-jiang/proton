#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""This is firewall manager"""
import re
import yaml

import tracer, netlib, tclients, nodeconf
from logger import syslog, syslog_alert, syslog_debug

from modules.ecmsdb import TFirewallDBManager
from modules.ecmsdb.t_firewall_status import TFirewallStatusDBManager
from modules.ecmsdb.t_node import TNodeDBManager

import common_lib
from common_lib import raise_t_exception

from ECMSAgent.ttypes import ncTServiceStatus
from ECMSManager.ttypes import ncTECMSManagerError
from ECMSManager.ttypes import ncTFirewallInfo, ncTHaSys

# 默认防火墙信息 yaml 配置
DEFAULT_FIREWALL_YAML = '/sysvol/apphome/conf/default_firewall.yaml'
# 支持的防火墙规则归属子系统
SUPPORTED_SYS_ROLE = ['basic', 'ecms', 'db', 'app', 'storage', 'asu']

MODULE_NAME = 'FirewallManager'


class FirewallManager(object):
    """
    firewall manager class
    """

    @classmethod
    @tracer.trace_func
    def get_firewall_status(cls):
        """获取集群防火墙状态"""
        return TFirewallStatusDBManager.get_status('cluster_firewall')

    @classmethod
    @tracer.trace_func
    def get_firewall_rule(cls, sys_role):
        """
        获取指定子系统的所有规则
        @return list<ncTFirewallInfo>
        """
        if sys_role not in SUPPORTED_SYS_ROLE:
            raise Exception('Param error, sys_role is not supported')

        rule_list = list()
        rule_list.extend(TFirewallDBManager.get_firewall_rule_by_role(sys_role))

        return rule_list

    @classmethod
    @tracer.trace_func
    def get_sys_service_status(cls, sys_role):
        """
        获取子系统对外服务状态
        @return True:  子系统允许对外服务
                False: 子系统不允许对外服务
        """
        if sys_role not in SUPPORTED_SYS_ROLE:
            raise Exception('Param error, sys_role is not supported')

        # 兼容asu
        if sys_role == 'asu':
            return True
        return TFirewallStatusDBManager.get_status(sys_role)

    @classmethod
    @tracer.trace_func
    def get_container_platform_node_ips(cls, debug=True):
        """获取容器平台ip，包括ssh ip和内部ip

        :return([type]): [description]
        """
        cp_infos = common_lib.get_container_platform_info(debug=debug)
        ips = []
        for host in cp_infos["hosts"].values():
            ip_keys = ["internal_ip", "ssh_ip"]
            for key in ip_keys:
                ip = host.get(key)
                if ip:
                    ips.append(ip)
        return ips

    @classmethod
    @tracer.trace_func
    def enable_firewall(cls):
        """启用防火墙"""
        syslog(MODULE_NAME, 'Enable firewall begin')

        # 检查防火墙状态,不允许重复开启
        if cls.get_firewall_status():
            raise Exception("Firewall is enable, can't open again.")

        # 获取所有节点信息
        node_info_list = TNodeDBManager.get_all_node_info()

        # 获取所有节点uuid,IP
        uuid_list = list()
        ip_list = list()
        for node_info in node_info_list:
            uuid_list.append(node_info.node_uuid)
            ip_list.append(node_info.node_ip)
        # 列表中加入127.0.0.1地址
        ip_list.append('127.0.0.1')

        # 获取容器平台ip
        container_platform_ips = cls.get_container_platform_node_ips()
        ip_list.extend(container_platform_ips)

        # 去重
        ip_list = list(set(ip_list))

        for node_info in node_info_list:
            node_ip = node_info.node_ip
            # 兼容asu节点,判断是asu节点,增加配置以下rich-rule
            asu_firewall_list = cls.get_asu_firewall_rule()

            syslog(MODULE_NAME, 'Enable firewall on node: %s' % node_ip)

            if node_info.is_online is False:
                syslog_alert(MODULE_NAME, "Node %s is offline, skip." % node_ip)
                continue

            with tclients.TClient("ECMSAgent", node_ip) as client:
                # 确保firewalld服务启动
                if client.get_service_status('firewalld') != ncTServiceStatus.SS_STARTED:
                    client.start_service('firewalld')
                client.init_firewall_xml()

                # 集群节点IP加入trusted信任区域
                for each_ip in ip_list:
                    client.add_source(each_ip, 'trusted', is_permanent=True)
                # 如果是asu节点,增加yaml文件中asu节点ip
                if client.is_asu_node() is True:
                    for each_ip in cls.get_asu_node_ip():
                        client.add_source(each_ip, 'trusted', is_permanent=True)

                # need_remove_ssh = False
                # 清理接口中已经处理了22端口,不再单独处理

                # 配置public区域
                rich_rule_list = list()
                for firewall_info in cls._get_firewall_info_by_uuid(node_info.node_uuid):
                    rich_rule = cls._encoding_rich_rule(firewall_info)
                    rich_rule_list.append(rich_rule)
                client.add_rich_rule(rich_rule_list, zone='public', is_permanent=True)

                # 处理asu节点
                is_asu = client.is_asu_node()
                if is_asu is True and asu_firewall_list != []:
                    client.add_rich_rule(asu_firewall_list, 'public', is_permanent=True)

                # 设置默认区域
                client.set_target('default', 'public')
                client.set_default_zone('public')
                # 重载防火墙规则
                # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                # client.reload_firewall(is_complete=False)

        # 数据库记录防火墙状态
        TFirewallStatusDBManager.update_status('cluster_firewall', True)

        syslog(MODULE_NAME, 'Enable firewall end')

    @classmethod
    @tracer.trace_func
    def disable_firewall(cls):
        """禁用防火墙"""
        syslog(MODULE_NAME, 'Disable firewall begin')

        # 检查防火墙状态,不允许重复开启
        if not cls.get_firewall_status():
            raise Exception("Firewall is disable, can't disable again.")

        node_info_list = TNodeDBManager.get_all_node_info()

        for node_info in node_info_list:
            node_ip = node_info.node_ip
            syslog(MODULE_NAME, 'Disable firewall on node: %s' % node_ip)

            if node_info.is_online is False:
                syslog_alert(MODULE_NAME, "Node %s is offline, skip." % node_ip)
                continue

            with tclients.TClient("ECMSAgent", node_ip) as client:
                # 确保firewalld服务启动
                if client.get_service_status('firewalld') != ncTServiceStatus.SS_STARTED:
                    client.start_service('firewalld')

                # 清理防火墙规则成默认状态, clear采用重命名配置文件方式,需要单独reload一次
                client.init_firewall_xml()
                # 设置默认区域
                client.set_target('ACCEPT', 'public')
                client.set_default_zone('trusted')

                # 载入规则
                client.reload_firewall(is_complete=False)

        # 数据库记录防火墙状态
        TFirewallStatusDBManager.update_status('cluster_firewall', False)

        syslog(MODULE_NAME, 'Disable firewall end')

    @classmethod
    @tracer.trace_func
    def add_firewall_rule(cls, firewall_info):
        """添加防火墙规则"""
        syslog(MODULE_NAME, 'Add firewall rule %r begin' % firewall_info)

        # 参数检查, 统一转换子网格式为xxx.xxx.xxx.xxx
        firewall_info = cls._verify_firewall_info(firewall_info)

        # 检查规则是否存在
        if TFirewallDBManager.exists_firewall_rule(firewall_info):
            raise Exception("Parameter error, firewall rule already exists")

        # 查询需要配置规则的节点
        ip_list = cls._get_ip_list_by_firewall_role_sys(firewall_info.role_sys)

        # 防火墙开启状态立即生效,否则只写入数据库
        if cls.get_firewall_status() and cls.get_sys_service_status(firewall_info.role_sys):
            # 添加规则
            for each_ip in ip_list:
                with tclients.TClient("ECMSAgent", each_ip) as client:
                    cmd_str = cls._encoding_rich_rule(firewall_info)
                    client.add_rich_rule([cmd_str], 'public', is_permanent=True)
                    # 重载防火墙规则
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(is_complete=False)

        # 写入数据库
        TFirewallDBManager.add_firewall_rule(firewall_info)

        syslog(MODULE_NAME, 'Add firewall rule %r end' % firewall_info)

    @classmethod
    @tracer.trace_func
    def del_firewall_rule(cls, firewall_info):
        """删除防火墙规则"""
        syslog(MODULE_NAME, 'Del firewall rule %r begin' % firewall_info)

        firewall_info = cls._verify_firewall_info(firewall_info)

        if not TFirewallDBManager.exists_firewall_rule(firewall_info):
            raise Exception("Parameter error, firewall rule does not exist")

        # 防火墙开启状态立即生效,否则只写入数据库
        if cls.get_firewall_status() and cls.get_sys_service_status(firewall_info.role_sys):
            # 查询需要配置规则的节点
            ip_list = cls._get_ip_list_by_firewall_role_sys(firewall_info.role_sys)
            # 删除一条规则
            for each_ip in ip_list:
                with tclients.TClient("ECMSAgent", each_ip) as client:
                    cmd_str = cls._encoding_rich_rule(firewall_info)
                    client.remove_rich_rule([cmd_str], 'public', is_permanent=True)
                    # 重载防火墙规则
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(is_complete=False)

        # 更新数据库
        TFirewallDBManager.del_firewall_rule(firewall_info)

        syslog(MODULE_NAME, 'Del firewall rule %r end' % firewall_info)

    @classmethod
    @tracer.trace_func
    def update_firewall_rule(cls, old_info, new_info):
        """更新防火墙规则"""
        syslog(MODULE_NAME, 'Update firewall rule %r to %r begin' % (old_info, new_info))

        # 参数检查
        old_info = cls._verify_firewall_info(old_info)
        new_info = cls._verify_firewall_info(new_info)

        # 旧规则不存在则抛错
        if not TFirewallDBManager.exists_firewall_rule(old_info):
            raise Exception("Parameter error, old firewall rule does not exist")

        # 检查新规则是否存在
        if (old_info.port != new_info.port or old_info.protocol != new_info.protocol or\
            old_info.source_net != new_info.source_net or old_info.dest_net != new_info.dest_net) \
            and TFirewallDBManager.exists_firewall_rule(new_info):
            raise Exception("Parameter error, new firewall rule already exists")

        # 查询需要配置规则的节点
        new_ip_list = cls._get_ip_list_by_firewall_role_sys(new_info.role_sys)

        # 取消限制,注释
        # # 不能向没有节点的子系统更新
        # if len(new_ip_list) == 0:
        #     raise Exception(
        #         "Can't update rule, sys %s not exists nodes" % new_info.role_sys)

        # 防火墙开启状态立即生效,否则只写入数据库
        if cls.get_firewall_status() and cls.get_sys_service_status(new_info.role_sys):

            # 获取需要移除的节点
            old_ip_list = cls._get_ip_list_by_firewall_role_sys(old_info.role_sys)
            # 移除
            for each_ip in old_ip_list:
                with tclients.TClient("ECMSAgent", each_ip) as client:
                    cmd_str = cls._encoding_rich_rule(old_info)
                    client.remove_rich_rule([cmd_str], 'public', is_permanent=True)
                    # 重载防火墙规则
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(is_complete=False)

            # 配置新规则
            for each_ip in new_ip_list:
                with tclients.TClient("ECMSAgent", each_ip) as client:
                    cmd_str = cls._encoding_rich_rule(new_info)
                    client.add_rich_rule([cmd_str], 'public', is_permanent=True)
                    # 重载防火墙规则
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(is_complete=False)

        # 更新数据库
        TFirewallDBManager.update_firewall_rule(old_info, new_info)

        syslog(MODULE_NAME, 'Update firewall rule %r to %r end' % (old_info, new_info))

    @classmethod
    @tracer.trace_func
    def enable_sys_service(cls, sys_role):
        """允许子系统对外服务"""
        if sys_role not in SUPPORTED_SYS_ROLE:
            raise Exception('Param error, sys_role is not supported')

        syslog(MODULE_NAME, 'Enable sys service sys_role=%s begin' % sys_role)

        # 检查状态,不允许重复开启
        if cls.get_sys_service_status(sys_role):
            raise Exception("Sys of %s is enable, can't enable again." % sys_role)

        # 查询该子系统的规则
        rule_info_list = cls.get_firewall_rule(sys_role)
        # 启用防火墙规则
        rich_rule_list = list()
        for rule_info in rule_info_list:
            rich_rule_list.append(cls._encoding_rich_rule(rule_info))

        ip_list = cls._get_ip_list_by_firewall_role_sys(sys_role)
        for each_ip in ip_list:
            syslog(MODULE_NAME, "Enable on: %s" % each_ip)
            with tclients.TClient("ECMSAgent", each_ip) as client:
                client.add_rich_rule(rich_rule_list, 'public', is_permanent=True)
                # 重载防火墙规则
                # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                # client.reload_firewall(is_complete=False)

        # 修改子系统状态
        TFirewallStatusDBManager.update_status(sys_role, True)

        syslog(MODULE_NAME, 'Enable sys service sys_role=%s end' % sys_role)

    @classmethod
    @tracer.trace_func
    def disable_sys_service(cls, sys_role):
        """禁止子系统对外服务"""
        if sys_role not in SUPPORTED_SYS_ROLE:
            raise Exception('Param error, sys_role is not supported')

        syslog(MODULE_NAME, 'Disable sys service sys_role=%s begin' % sys_role)

        # 检查状态,不允许重复禁用
        if not cls.get_sys_service_status(sys_role):
            raise Exception("Sys of %s is disable, can't disable again." % sys_role)

        syslog(MODULE_NAME, 'Disable sys service begin.')

        # 查询该子系统的规则
        rule_info_list = cls.get_firewall_rule(sys_role)

        # 移除防火墙规则
        rich_rule_list = list()
        for rule_info in rule_info_list:
            rich_rule_list.append(cls._encoding_rich_rule(rule_info))

        ip_list = cls._get_ip_list_by_firewall_role_sys(sys_role)
        for each_ip in ip_list:
            syslog(MODULE_NAME, "Disable on: %s" % each_ip)
            with tclients.TClient("ECMSAgent", each_ip) as client:
                client.remove_rich_rule(rich_rule_list, 'public', is_permanent=True)
                # 重载防火墙规则
                # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                # client.reload_firewall(is_complete=False)

        # 修改子系统状态
        TFirewallStatusDBManager.update_status(sys_role, False)

        syslog(MODULE_NAME, 'Disable sys service sys_role=%s end' % sys_role)

    @classmethod
    @tracer.trace_func
    def add_trusted_ip(cls, ip_list):
        """添加集群内部的信任ip"""
        syslog(MODULE_NAME, 'Add trusted ip begin.')

        # 检查节点内部 IP 与集群内部 IP 是否处于同一网段
        master_uuid = TNodeDBManager.get_role_ecms_master_uuid()
        master_ip = TNodeDBManager.get_node_info(master_uuid).node_ip

        with tclients.TClient('ECMSAgent', master_ip) as client:
            master_ifaddr = client.get_ifaddr_by_ipaddr(ipaddr=master_ip)

        for each_ip in ip_list:
            if not netlib.is_same_network(master_ifaddr.ipaddr, each_ip, master_ifaddr.netmask):
                raise_t_exception(
                    exp_msg="node ip {0} and node ip {1} are not in the same subnet."
                    .format(master_ifaddr.ipaddr, each_ip),
                    exp_id=ncTECMSManagerError.NCT_INVALID_ARGUMENT)

        # 获取节点信息
        node_info_list = TNodeDBManager.get_all_node_info()

        for node_info in node_info_list:
            # 存在离线，添加信任 ip 时跳过，记录日志
            if node_info.is_online is False:
                syslog_alert(MODULE_NAME, "Node %s is offline, skip." % node_info.node_uuid)
                continue
            with tclients.TClient("ECMSAgent", node_info.node_ip) as client:
                for each_ip in ip_list:
                    client.add_source(each_ip, 'trusted', True)
                # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                # client.reload_firewall(False)

        syslog(MODULE_NAME, 'Add trusted ip end.')

    @classmethod
    @tracer.trace_func
    def set_storage_trusted_ip(cls, firewall_list):
        """设置本地存储（Ceph）的信任ip，覆盖原有规则"""
        syslog(MODULE_NAME, 'Add storage trusted ip begin')

        role_sys = 'asu'

        new_firewall_list = list()
        new_rich_rule_list = list()
        for firewall_info in firewall_list:
            # 判断是否为 asu 类型规则，且则不指定端口
            if firewall_info.role_sys != role_sys or firewall_info.port != 0:
                raise Exception("(%r) is not a storage trusted ip rule" % firewall_info)

            # 参数检查, 统一转换子网格式为xxx.xxx.xxx.xxx
            firewall_info = cls._verify_firewall_info(firewall_info)
            new_firewall_list.append(firewall_info)

            rich_rule = cls._encoding_rich_rule(firewall_info)
            new_rich_rule_list.append(rich_rule)

        # 获取待删除的规则
        old_firewall_list = list()
        old_rich_rule_list = list()
        asu_firewall_rules = TFirewallDBManager.get_firewall_rule_by_role(role_sys)
        for firewall_info in asu_firewall_rules:
            if firewall_info.role_sys != role_sys or firewall_info.port != 0:
                continue
            old_firewall_list.append(firewall_info)
            rich_rule = cls._encoding_rich_rule(firewall_info)
            old_rich_rule_list.append(rich_rule)

        # 查询需要配置规则的节点
        ip_list = cls._get_ip_list_by_firewall_role_sys(role_sys)

        # 防火墙开启状态立即生效,否则只写入数据库
        if cls.get_firewall_status() and cls.get_sys_service_status(role_sys):
            for each_ip in ip_list:
                with tclients.TClient("ECMSAgent", each_ip) as client:
                    # 清理原有规则
                    if len(old_rich_rule_list) != 0:
                        client.remove_rich_rule(old_rich_rule_list, 'public', is_permanent=True)
                    # 添加新规则
                    if len(new_rich_rule_list) != 0:
                        client.add_rich_rule(new_rich_rule_list, 'public', is_permanent=True)
                    # 重载防火墙规则
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(is_complete=False)

        # 删除数据库中原有规则
        for firewall_info in old_firewall_list:
            TFirewallDBManager.del_firewall_rule(firewall_info)
        # 新规则写入数据库
        for firewall_info in new_firewall_list:
            TFirewallDBManager.add_firewall_rule(firewall_info)

        syslog(MODULE_NAME, 'Add storage trusted ip end')

    @classmethod
    @tracer.trace_func
    def update_public_zone(cls, node_uuid):
        """更新指定节点public区域的rich rule规则"""
        syslog(MODULE_NAME, 'Update rich rule begin.')

        node_info = TNodeDBManager.get_node_info(node_uuid)
        node_ip = node_info.node_ip
        # 兼容asu节点,判断是asu节点,增加配置以下rich-rule
        asu_firewall_list = cls.get_asu_firewall_rule()

        with tclients.TClient("ECMSAgent", node_ip) as client:

            # 配置节点rich rule规则
            firewall_info_list = list()

            # 获取节点配置规则
            if TFirewallStatusDBManager.get_status('cluster_firewall') is True:

                # ========================================================================
                # 各子系统配置规则
                if node_info.role_ecms == 1:
                    # 查询ecms子系统状态
                    if TFirewallStatusDBManager.get_status('ecms'):
                        firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('ecms'))
                if node_info.role_db != 0:
                    # 查询db子系统状态
                    if TFirewallStatusDBManager.get_status('db'):
                        firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('db'))
                if node_info.role_app != 0 or \
                   node_info.is_ha == ncTHaSys.BASIC or \
                   node_info.is_ha == ncTHaSys.APP:
                    # 查询应用子系统状态
                    if TFirewallStatusDBManager.get_status('app'):
                        firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('app'))
                if node_info.role_storage != 0 or \
                   node_info.is_ha == ncTHaSys.BASIC or \
                   node_info.is_ha == ncTHaSys.STORAGE:
                    # 查询存储子系统状态
                    if TFirewallStatusDBManager.get_status('storage'):
                        firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('storage'))
                # 查询基础规则状态
                if TFirewallStatusDBManager.get_status('basic'):
                    # 添加基础规则
                    firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('basic'))
                # ========================================================================

                client.set_default_zone('public')
                client.set_target('default', 'public')
            else:
                # 防火墙关闭状态,asu规则列表置空
                asu_firewall_list = []

                client.set_default_zone('trusted')
                client.set_target('ACCEPT', 'public')

            # 清理rich rule规则
            rule_list = client.get_firewall_info('rich-rule', 'public', is_permanent=True)
            client.remove_rich_rule(rule_list, 'public', is_permanent=True)

            rich_rule_list = list()
            for firewall_info in firewall_info_list:
                rich_rule_list.append(cls._encoding_rich_rule(firewall_info))

                # 判断是否移除firewall中ssh服务
                if firewall_info.port == 22:
                    service_list = client.get_firewall_info('service', 'public', is_permanent=True)
                    if 'ssh' in service_list:
                        client.remove_service('ssh', 'public', True)

            client.add_rich_rule(rich_rule_list, 'public', is_permanent=True)

            # ========================================================================
            # 配置asu节点
            # ceph容器化后需要根据集群决定是否配asu防火墙
            with tclients.TClient("ECMSAgent") as local_client:
                is_asu = local_client.is_asu_node()
            if is_asu is True and asu_firewall_list != []:
                client.add_rich_rule(asu_firewall_list, 'public', is_permanent=True)
            # ========================================================================

            # 载入规则
            # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
            # client.reload_firewall(False)

        syslog(MODULE_NAME, 'Update rich rule end.')

    @classmethod
    @tracer.trace_func
    def update_trusted_zone(cls, node_uuid, need_reload_firewall=True):
        """更新信任区域"""
        syslog(MODULE_NAME, 'Update trusted zone on node[%s] begin' % node_uuid)

        # 获取节点信息
        curr_node = TNodeDBManager.get_node_info(node_uuid)
        node_ip = curr_node.node_ip

        # 获取所有节点信息
        node_info_list = TNodeDBManager.get_all_node_info()
        # 获取所有节点IP
        ip_list = list()
        for node_info in node_info_list:
            ip_list.append(node_info.node_ip)
        # 列表中加入127.0.0.1地址
        ip_list.append('127.0.0.1')

        container_platform_ips = cls.get_container_platform_node_ips()
        ip_list.extend(container_platform_ips)

        with tclients.TClient("ECMSAgent", node_ip) as client:
            # 移除原trusted区域所有source
            source_list = client.get_firewall_info('source', 'trusted', is_permanent=True)
            for source in source_list:
                client.remove_source(source, 'trusted', True)

            if TFirewallStatusDBManager.get_status('cluster_firewall') is True:

                # 如果是asu节点,增加yaml文件中asu节点ip
                if client.is_asu_node() is True:
                    ip_list.extend(cls.get_asu_node_ip())

                # 集群节点IP加入trusted信任区域
                for each_ip in list(set(ip_list)):
                    client.add_source(each_ip, 'trusted', True)
                client.set_default_zone('public')
                client.set_target('default', 'public')
            else:
                client.set_default_zone('trusted')
                client.set_target('ACCEPT', 'public')

            # 载入规则
            # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
            # if need_reload_firewall:
            #     client.reload_firewall(False)

        syslog(MODULE_NAME, 'Update trusted zone on node[%s] end' % node_uuid)

# ============================================================================
# 管理asu防火墙模块函数
# ============================================================================
    @classmethod
    @tracer.trace_func
    def get_asu_firewall_rule(cls):
        """
        获取asu放行端口
        @ return list<string> 防火墙rich-rule规则列表
        """
        rich_rule_list = list()
        for each_info in TFirewallDBManager.get_firewall_rule_by_role('asu'):
            rich_rule_list.append(cls._encoding_rich_rule(each_info))
        return rich_rule_list

    @classmethod
    @tracer.trace_func
    def get_asu_node_ip(cls):
        """获取asu节点ip信息"""
        asu_node_ip_list = list()
        for each_info in TNodeDBManager.get_all_node_info():
            try:
                node_ip = each_info.node_ip
                with tclients.TClient("ECMSAgent", node_ip) as client:
                    if client.is_asu_node() is True:
                        asu_node_ip_list.append(node_ip)
            except Exception as ex:
                syslog_alert(MODULE_NAME, "Can't connect node %s:%s" % (node_ip, str(ex)))
        return asu_node_ip_list

# ============================================================================
# 内部函数
# ============================================================================
    @classmethod
    @tracer.trace_func
    def _encoding_rich_rule(cls, firewall_info, family='ipv4'):
        """
        将数据库防火墙信息构造成富规则字符串,accept规则
        目标规则:
        rule family="ipv4" source address="0.0.0.0" \
        destination address="192.168.136.190" port port="3306" protocol="tcp" accept
        @return str
        """
        if firewall_info:
            rich_rule = 'rule family="%s"' % family
            if firewall_info.source_net:
                rich_rule = rich_rule + ' source address="%s"' % firewall_info.source_net
            if firewall_info.dest_net:
                rich_rule = rich_rule + ' destination address="%s"' % firewall_info.dest_net
            if firewall_info.port:
                rich_rule = rich_rule + ' port port="%d" protocol="%s"' % (firewall_info.port,
                                                                            firewall_info.protocol)
            rich_rule = rich_rule + ' accept'
        return rich_rule

    @classmethod
    @tracer.trace_func
    def _decoding_rich_rule(cls, rich_rule):
        """根据防火墙rich rule规则生成防火墙结构"""
        firewall_info = ncTFirewallInfo()

        re_str = '^rule family=".+?"( source address="(?P<sip>.+?)")?( destination address="(?P<dip>.+?)")?( port port="(?P<port>.+?)" protocol="(?P<ptl>.+?)")? accept$'
        reobj = re.match(re_str, rich_rule)

        if reobj:
            firewall_dict = reobj.groupdict()
            firewall_info.role_sys = ""
            firewall_info.service_desc = ""

            if firewall_dict['port']:
                firewall_info.port = int(firewall_dict['port'])
                firewall_info.protocol = firewall_dict['ptl']

            if firewall_dict['sip']:
                firewall_info.source_net = firewall_dict['sip']

            if firewall_dict['dip']:
                firewall_info.dest_net = firewall_dict['dip']

        return firewall_info

    @classmethod
    @tracer.trace_func
    def _get_firewall_info_by_uuid(cls, node_uuid):
        """根据节点id获取节点防火墙规则"""
        firewall_info_list = list()

        node_info = TNodeDBManager.get_node_info(node_uuid)

        if cls.get_sys_service_status('basic'):
            firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('basic'))

        # ecms主节点 ecms子系统应用访问规则
        if node_info.role_ecms == 1 and cls.get_sys_service_status('ecms'):
            firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('ecms'))
        # 数据库节点获取规则
        if node_info.role_db != 0 and cls.get_sys_service_status('db'):
            firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('db'))
        # 应用节点获取规则
        if (node_info.role_app != 0 or \
            node_info.is_ha == ncTHaSys.BASIC or \
            node_info.is_ha == ncTHaSys.APP) and \
            cls.get_sys_service_status('app'):
            firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('app'))
        # 存储节点获取规则
        if (node_info.role_storage != 0 or \
            node_info.is_ha == ncTHaSys.BASIC or \
            node_info.is_ha == ncTHaSys.STORAGE) and \
            cls.get_sys_service_status('storage'):
            firewall_info_list.extend(TFirewallDBManager.get_firewall_rule_by_role('storage'))

        # aus节点获取指定规则
        rule_list = cls.get_asu_firewall_rule()
        with tclients.TClient("ECMSAgent", node_info.node_ip) as client:
            if client.is_asu_node() is True:
                for each_rule in rule_list:
                    firewall_info_list.append(cls._decoding_rich_rule(each_rule))

        return firewall_info_list

    @classmethod
    @tracer.trace_func
    def _verify_firewall_info(cls, firewall_info):
        """
        验证防火墙信息结构的参数, 并转换子网为xxx.xxx.xxx.xxx形式
        """
        def _verify_ip(param, ip_str):
            """验证ip及掩码"""
            if ip_str.find('/') == -1:
                raise Exception("ncTFirewallInfo.%s error, not found netmask" % param)
            ip = ip_str.split('/')[0]
            mask = ip_str.split('/')[1]
            # 如果掩码是cidr
            if mask.isalnum() is True:
                if int(mask) not in [num for num in xrange(0, 33)]:
                    raise Exception("ncTFirewallInfo.%s error, CIDR error" % (param))
                mask = netlib.exchange_int_to_mask(int(mask))
            else:
                if not netlib.is_valid_mask(mask):
                    raise Exception(
                        "ncTFirewallInfo.%s error, mask is not available" % (param))

            # 验证ip和掩码
            if not netlib.is_valid_ip(ip):
                raise Exception("ncTFirewallInfo.%s error, ip is not available" % (param))
            # 转换为统一形式返回
            return ip + '/' + mask

        if firewall_info.protocol != '':
            if firewall_info.protocol not in ['tcp', 'udp']:
                raise Exception('ncTFirewallInfo.protocol error')
        if firewall_info.source_net != '':
            firewall_info.source_net = _verify_ip('source_net', firewall_info.source_net)

        if firewall_info.dest_net != '':
            firewall_info.dest_net = _verify_ip('dest_net', firewall_info.dest_net)

        if firewall_info.role_sys not in SUPPORTED_SYS_ROLE:
            raise Exception('ncTFirewallInfo.role_sys error')
        return firewall_info

    @classmethod
    @tracer.trace_func
    def _is_valid_firewall_info(cls, firewall_info):
        """
        判断防火墙信息结构是否有效
        """
        if firewall_info.protocol != '':
            if firewall_info.protocol not in ['tcp', 'udp']:
                return False
        if firewall_info.source_net != '':
            if not netlib.is_valid_ip(firewall_info.source_net):
                return False
        if firewall_info.dest_net != '':
            if not netlib.is_valid_ip(firewall_info.dest_net):
                return False
        if firewall_info.role_sys not in SUPPORTED_SYS_ROLE:
            return False
        return True

    @classmethod
    @tracer.trace_func
    def _get_ip_list_by_firewall_role_sys(cls, role_sys):
        """根据防火墙规则,获取需要配置的节点ip"""
        if role_sys == 'ecms':
            uuid_list = [TNodeDBManager.get_role_ecms_master_uuid()]
        if role_sys == 'db':
            uuid_list = TNodeDBManager.get_role_db_uuid()
        if role_sys == 'app':
            uuid_list = TNodeDBManager.get_role_app_uuid()
        if role_sys == 'storage':
            uuid_list = TNodeDBManager.get_role_storage_uuid()

        if role_sys == 'basic':
            node_info_list = TNodeDBManager.get_all_node_info()
            uuid_list = list(node_info.node_uuid for node_info in node_info_list)

        # 兼容asu,直接返回ip
        if role_sys == 'asu':
            return cls.get_asu_node_ip()

        ip_list = list()
        for each_uuid in uuid_list:
            node_info = TNodeDBManager.get_node_info(each_uuid)
            # 只获取在线节点
            if node_info.is_online is True:
                ip_list.append(node_info.node_ip)
        return ip_list

# ==================================================================================
# 防火墙管理公共模块
# ==================================================================================
    @classmethod
    @tracer.trace_func
    def set_sys_role_firewall(cls, node_ip, sys_role):
        """
        在指定节点上配置指定角色的防火墙规则
        @param str node_ip 指定节点的IP
        """
        # 查询防火墙状态
        f_status = TFirewallStatusDBManager.get_status('cluster_firewall')
        # 查询子系统防火墙状态
        s_status = TFirewallStatusDBManager.get_status(sys_role)
        if f_status is True and s_status is True:
            # 获取存储子系统防火墙规则
            firewall_list = list()
            firewall_list.extend(TFirewallDBManager.get_firewall_rule_by_role(sys_role))

            # 配置到节点
            if firewall_list:
                rich_rule_list = list()
                for firewall_info in firewall_list:
                    cmd_str = cls._encoding_rich_rule(firewall_info)
                    rich_rule_list.append(cmd_str)
                with tclients.TClient("ECMSAgent", node_ip) as client:
                    client.add_rich_rule(rich_rule_list, 'public', is_permanent=True)
                    # 重载防火墙规则
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(is_complete=False)

    @classmethod
    @tracer.trace_func
    def remove_sys_role_firewall(cls, node_ip, sys_role):
        """移除指定节点的存储子系统规则"""
        # 查询防火墙状态
        f_status = TFirewallStatusDBManager.get_status('cluster_firewall')
        # 查询子系统防火墙状态
        s_status = TFirewallStatusDBManager.get_status(sys_role)

        if f_status is True and s_status is True:
            # 状态开启表明节点有子系统规则存储
            firewall_list = list()
            firewall_list.extend(TFirewallDBManager.get_firewall_rule_by_role(sys_role))
            # 配置到节点
            if firewall_list:
                with tclients.TClient("ECMSAgent", node_ip) as client:

                    # 获取 asu 规则
                    asu_list = list()
                    if client.is_asu_node() is True:
                        asu_info = TFirewallDBManager.get_firewall_rule_by_role('asu')
                        asu_list = [cls._encoding_rich_rule(info) for info in asu_info]

                    rich_rule_list = list()
                    for firewall_info in firewall_list:
                        cmd_str = cls._encoding_rich_rule(firewall_info)
                        # asu节点上, 不移除asu规则
                        if cmd_str in asu_list:
                            continue
                        rich_rule_list.append(cmd_str)

                    client.remove_rich_rule(rich_rule_list, 'public', is_permanent=True)
                    # 重载防火墙规则
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(is_complete=False)


class FirewallProcessor(common_lib.BaseProcessor):
    """
    firewall processor class
    """
    @classmethod
    @tracer.trace_func
    def on_is_env_dirty(cls, config):
        """检查节点环境"""
        with tclients.TClient("ECMSAgent", config['node_ipaddr']) as client:
            public_rich_rule = client.get_firewall_info('rich-rule', 'public', is_permanent=True)

            # 初始存在的规则,不检查
            default_rule = ['rule family="ipv4" port port="22" protocol="tcp" accept',
                            'rule family="ipv4" port port="9202" protocol="tcp" accept']
            for each_rule in default_rule:
                if each_rule in public_rich_rule:
                    public_rich_rule.remove(each_rule)

            if public_rich_rule != []:
                syslog_debug(
                    MODULE_NAME, "NODE DIRTY: public zone exists rule %r" % public_rich_rule)
                return True
            trusted_source = client.get_firewall_info('source', 'trusted', is_permanent=True)
            if trusted_source != []:
                syslog_debug(MODULE_NAME, "NODE DIRTY: trusted zone exists %r" % trusted_source)
                return True
            return False

    @classmethod
    @tracer.trace_func
    def on_clear_node(cls, config):
        """
        清理防火墙规则
        """
        syslog(MODULE_NAME, 'on clear node begin.')

        # 清理防火墙规则: 全部放行
        with tclients.TClient("ECMSAgent", config['node_ipaddr']) as client:
            try:
                # 确保firewalld服务启动
                if client.get_service_status('firewalld') != ncTServiceStatus.SS_STARTED:
                    client.start_service('firewalld')
                client.init_firewall_xml()
                # 设置默认区域
                client.set_target('ACCEPT', 'public')
                client.set_default_zone('trusted')

                # 载入规则
                client.reload_firewall(is_complete=True)
            except Exception:
                syslog_debug(MODULE_NAME, "Clear firewall has Exception, skip")

        syslog(MODULE_NAME, 'on clear node end.')

    @classmethod
    @tracer.trace_func
    def on_active_cluster(cls, config):
        """
        激活集群操作
        """
        syslog(MODULE_NAME, 'On active cluster begin')
        # 默认的防火墙规则
        syslog(MODULE_NAME, 'Adding default firewall rule to database...')
        for firewall_rule in yaml.load(open(DEFAULT_FIREWALL_YAML)):
            info = ncTFirewallInfo()
            info.port = int(firewall_rule['port'])
            info.role_sys = firewall_rule['sys_role']
            info.service_desc = firewall_rule['description']
            info.protocol = firewall_rule['protocol']
            if FirewallManager._is_valid_firewall_info(info) \
                    and (not TFirewallDBManager.exists_firewall_rule(info)):
                TFirewallDBManager.add_firewall_rule(info)
        syslog(MODULE_NAME, 'On active cluster end')

    @classmethod
    @tracer.trace_func
    def on_add_node_into_cluster(cls, config):
        """
        添加节点
        """
        if FirewallManager.get_firewall_status() is False:
            return

        syslog(MODULE_NAME, 'On add node info cluster begin')

        master_ip = config['ecms_ip']
        curr_ip = config['node_info'].node_ip

        # 在ecms节点临时放行被添加节点的IP
        with tclients.TClient("ECMSAgent", master_ip) as client:
            client.add_source(curr_ip, 'trusted', is_permanent=False)

        # 在被添加节点临时放行ecms节点IP
        # 初始化防火墙规则
        with tclients.TClient("ECMSAgent", curr_ip) as client:
            client.add_source(master_ip, 'trusted', is_permanent=False)
            client.init_firewall_xml()

        syslog(MODULE_NAME, 'On add node info cluster end')

    @classmethod
    @tracer.trace_func
    def on_node_join_cluster(cls, config):
        """
        节点加入集群操作
        """
        if FirewallManager.get_firewall_status() is False:
            return

        syslog(MODULE_NAME, 'on node join cluster begin.')

        new_node_uuid = config['node_info'].node_uuid

        node_info_list = TNodeDBManager.get_all_node_info()
        # 所有节点配置trusted区域
        for node_info in node_info_list:

            # 跳过节点离线
            if node_info.is_online is False:
                syslog_alert(MODULE_NAME, "Node %s is offline, skip." % node_info.node_uuid)
                continue
            if node_info.node_uuid == new_node_uuid:
                continue

            FirewallManager.update_trusted_zone(node_info.node_uuid)

        # 被添加节点，配置trusted区域
        FirewallManager.update_trusted_zone(new_node_uuid, need_reload_firewall=False)
        # 被添加节点，配置public区域
        FirewallManager.update_public_zone(new_node_uuid)

        syslog(MODULE_NAME, 'on node join cluster end.')

    @classmethod
    @tracer.trace_func
    def on_remove_node_from_cluster(cls, config):
        """移除节点"""
        if FirewallManager.get_firewall_status() is False:
            return

        syslog(MODULE_NAME, 'On remove node from cluster begin')
        # 集群节点移除
        del_node = config['node_info']
        all_node_list = config['all_node_info']
        firewall_status = TFirewallStatusDBManager.get_status('cluster_firewall')

        for node_info in all_node_list:
            # 跳过节点离线
            if node_info.is_online is False:
                syslog_alert(MODULE_NAME, "Node %s is offline, skip." % node_info.node_uuid)
                continue

            # 处理被移除节点
            if node_info.node_uuid == del_node.node_uuid:
                with tclients.TClient("ECMSAgent", del_node.node_ip) as client:
                    # 如果被移除节点是asu节点,则配置asu防火墙规则
                    if client.is_asu_node() is True and firewall_status is True:
                        # 移除asu节点public规则
                        rule_list = client.get_firewall_info(
                            'rich-rule', 'public', is_permanent=True)
                        client.remove_rich_rule(rule_list, 'public', is_permanent=True)
                        # asu 节点单独配置asu规则
                        asu_firewall_list = FirewallManager.get_asu_firewall_rule()

                        # 加入22端口规则, 9202端口
                        for port in [22, 9202]:
                            firewall_info = ncTFirewallInfo()
                            firewall_info.port = port
                            firewall_info.protocol = 'tcp'
                            asu_firewall_list.append(FirewallManager._encoding_rich_rule(firewall_info))

                        client.add_rich_rule(asu_firewall_list, 'public', is_permanent=True)
                        # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                        # client.reload_firewall(False)
                # 被移除节点不再处理下面逻辑, 跳出
                continue

            with tclients.TClient("ECMSAgent", node_info.node_ip) as client:
                client.remove_source(del_node.node_ip, 'trusted', True)
                # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                # client.reload_firewall(False)

        syslog(MODULE_NAME, 'On remove node from cluster end')

    @classmethod
    @tracer.trace_func
    def on_troubleshoot(cls):
        """参见基类描述"""
        # 未启用防火墙模块, 跳过检查
        if FirewallManager.get_firewall_status() is False:
            return

        # 查看防火墙服务是否被停止
        with tclients.TClient("ECMSAgent") as client:
            if client.get_service_status('firewalld') != ncTServiceStatus.SS_STARTED:
                client.start_service('firewalld')

    @classmethod
    @tracer.trace_func
    def on_consistency_repair(cls):
        """参见基类描述"""
        syslog(MODULE_NAME, "on_consistency_repair start.")

        if cls.on_consistency_check() is True:
            syslog(MODULE_NAME, "Skip firewall consistency repair")
            return

        node_uuid = nodeconf.NodeConfig.get_node_uuid()
        FirewallManager.update_trusted_zone(node_uuid, need_reload_firewall=False)
        FirewallManager.update_public_zone(node_uuid)

        syslog(MODULE_NAME, "on_consistency_repair end.")

    @classmethod
    @tracer.trace_func
    def on_consistency_check(cls):
        """参见基类描述"""
        # 未启用防火墙模块, 跳过检查
        if FirewallManager.get_firewall_status() is False:
            return True

        node_uuid = nodeconf.NodeConfig.get_node_uuid()
        node_info = TNodeDBManager.get_node_info(node_uuid)
        all_node_list = TNodeDBManager.get_all_node_info()

        with tclients.TClient("ECMSAgent", node_info.node_ip) as client:
            # 判断默认区域
            default_zone = client.get_default_zone()
            target = client.get_target("public")
            cluster_status = TFirewallStatusDBManager.get_status('cluster_firewall')

            trusted_source = client.get_firewall_info("source", "trusted", is_permanent=True)

            public_rich_rule = client.get_firewall_info("rich-rule", "public", is_permanent=True)

            # 防火墙关闭情况
            if cluster_status is False:

                if default_zone != "trusted":
                    syslog_debug(MODULE_NAME, "CHECK RET: default zone is %s" % default_zone)
                    return False
                if target != "ACCEPT":
                    syslog_debug(MODULE_NAME, "CHECK RET: public target is %s" % target)
                    return False

            else:
                # 防火墙开启状态
                if default_zone != "public":
                    syslog_debug(MODULE_NAME, "CHECK RET: default zone is %s" % default_zone)
                    return False
                if target != "default":
                    syslog_debug(MODULE_NAME, "CHECK RET: target is %s" % target)
                    return False

                # 判断trusted区域
                db_ip_list = list(each_info.node_ip for each_info in all_node_list)
                db_ip_list.append('127.0.0.1')
                # 如果是asu节点,增加yaml文件中asu节点ip
                if client.is_asu_node() is True:
                    db_ip_list.extend(FirewallManager.get_asu_node_ip())

                container_platform_ips = FirewallManager.get_container_platform_node_ips(debug=False)
                db_ip_list.extend(container_platform_ips)

                db_ip_set = set(db_ip_list)

                get_ip_set = set(trusted_source)
                diff_set = get_ip_set.symmetric_difference(db_ip_set)
                if diff_set != set():
                    syslog_debug(
                        MODULE_NAME, "CHECK RET: trusted sources not null(%r)" % diff_set)
                    return False

                # 判断public区域
                db_firewall_list = list()
                # 特殊处理防火墙结构中的role_sys信息
                for each_info in FirewallManager._get_firewall_info_by_uuid(node_uuid):
                    db_firewall_list.append(FirewallManager._encoding_rich_rule(each_info))

                # 返回一个新的 set 包含两个set中不重复的元素
                diff_set = set(public_rich_rule).symmetric_difference(set(db_firewall_list))
                if diff_set != set():
                    syslog_debug(
                        MODULE_NAME, "CHECK RET: public rich rule not equal(%r)" % diff_set)
                    return False

        # 通过检查
        return True

    # ==============================================================
    # 节点高可用使用接口
    # ==============================================================
    @classmethod
    @tracer.trace_func
    def on_node_entering_master(cls):
        """节点进入高可用主"""
        node_uuid = nodeconf.NodeConfig.get_node_uuid()
        node_info = TNodeDBManager.get_node_info(node_uuid)

        if node_info.is_ha == ncTHaSys.BASIC or node_info.is_ha == ncTHaSys.APP:
            # 开放ecms角色端口
            rule_str_list = list()
            for each_rule in FirewallManager.get_firewall_rule('ecms'):
                rule_str_list.append(FirewallManager._encoding_rich_rule(each_rule))

            if FirewallManager.get_firewall_status() and \
            FirewallManager.get_sys_service_status('ecms'):
                # 当前节点防火墙添加规则
                with tclients.TClient("ECMSAgent") as client:
                    client.add_rich_rule(rule_str_list, 'public', is_permanent=True)
                    # 重载防火墙规则
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(is_complete=False)

                master_uuid = TNodeDBManager.get_role_ecms_master_uuid()
                if node_uuid != master_uuid:
                    master_info = TNodeDBManager.get_node_info(master_uuid)
                    # 原集群管理节点,移除规则.只修改在线节点
                    try:
                        with tclients.TClient("ECMSAgent", master_info.node_ip) as client:
                            client.remove_rich_rule(rule_str_list, 'public', is_permanent=True)
                            # 重载防火墙规则
                            # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                            # client.reload_firewall(is_complete=False)
                    except Exception:
                        syslog_alert(MODULE_NAME, "connect to node[%s] failed" % master_info.node_uuid)

    @classmethod
    @tracer.trace_func
    def on_node_entering_slave(cls):
        """节点进入高可用从"""
        pass

    @classmethod
    @tracer.trace_func
    def on_node_entering_master_check(cls):
        """pass"""
        return "continue"

    @classmethod
    @tracer.trace_func
    def on_node_entering_slave_check(cls):
        """pass"""
        return "continue"

    @classmethod
    @tracer.trace_func
    def on_change_node_ip(cls, config):
        """更改节点ip"""
        syslog(MODULE_NAME, 'On change node ip begin.')

        old_ip = config['old_ip']

        # 获取节点信息
        node_info_list = TNodeDBManager.get_all_node_info()

        # 存在离线，删除节点原 ip 时跳过，记录日志
        for node_info in node_info_list:
            if node_info.is_online is False:
                syslog_alert(MODULE_NAME, "Node %s is offline, skip." % node_info.node_uuid)
                continue
            with tclients.TClient("ECMSAgent", node_info.node_ip) as client:
                result_list = client.get_firewall_info('source', 'trusted', True)
                if old_ip in result_list:
                    client.remove_source(old_ip, 'trusted', True)
                    # 下面的重载防火墙会导致k8s创建的iptables临时规则被删除，暂时屏蔽
                    # client.reload_firewall(False)

        syslog(MODULE_NAME, 'On change node ip end.')

    ###########################################################################
    # 模块管理
    ###########################################################################

    @classmethod
    @tracer.trace_func
    def on_enable_zabbix(cls, config):
        """启用 zabbix 操作"""
        syslog(MODULE_NAME, 'on enable zabbix begin')

        # 开放防火墙端口
        syslog_debug(MODULE_NAME, 'opening firewall port')
        FirewallManager.add_firewall_rule(
            firewall_info=ncTFirewallInfo(
                port=config.get('gui_port'),
                protocol='tcp',
                source_net='',
                dest_net='',
                role_sys='ecms',
                service_desc='Zabbix Service(httpd)',
            ),
        )

        syslog(MODULE_NAME, 'on enable zabbix end')

    @classmethod
    @tracer.trace_func
    def on_disable_zabbix(cls, config):
        """禁用 zabbix 操作"""
        syslog(MODULE_NAME, 'on disable zabbix begin')

        # 关闭防火墙端口
        syslog_debug(MODULE_NAME, 'closing firewall port')
        FirewallManager.del_firewall_rule(
            firewall_info=ncTFirewallInfo(
                port=config.get('gui_port'),
                protocol='tcp',
                source_net='',
                dest_net='',
                role_sys='ecms',
                service_desc='Zabbix Service(httpd)',
            ),
        )

        syslog(MODULE_NAME, 'on disable zabbix end')
