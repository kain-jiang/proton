#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""
net agent 集群网络管理代理模块
"""
import re
import os
import shutil
import subprocess
from copy import deepcopy

from src.modules.pydeps import logger, netlib, tracer, filelib, cmdprocess

BOND_NAME = "bond%s"
ETHTOOL_OPTS = "-K %s gro off lro off"
IFCFG_DIR_PATH = "/etc/sysconfig/network-scripts"
NIC_CFG_FILE_PATH = "/etc/sysconfig/network-scripts/ifcfg-%s"

MODULE_NAME = 'NetAgent'

ncTIfAddr = {
    "nic_dev_name": "",
    "label": "",
    "ipaddr": "",
    "netmask": "",
    "prefix": "",
    "gateway": ""
}

ncTNic = {
    "nic_dev_name": "",        # 设备名，如 em1,bond0,bond0.32等
    "is_up": False,            # 设备状态是 UP or DOWN
    "state_info": "",          # 设备状态详情
    "hw_info": "",             # 设备物理硬件信息
    "ifaddrs": [],              # 该设备上的协议地址列表
}


class NetAgent(object):
    """
    This is network agent class
    """

    ########################################################################################
    # 以下函数为接口功能实现
    @classmethod
    @tracer.trace_func
    def get_ip_addrs(cls):
        """
        获取当前节点 IP 列表
        注意：不包括lo网卡
        """
        nic_list = cls.get_nics()
        ip_set = set()
        for each_nic in nic_list:
            if each_nic["nic_dev_name"] == 'lo':
                continue
            for each_ifaddr in each_nic["ifaddrs"]:
                ip_set.add(each_ifaddr["ipaddr"])
        return list(ip_set)

    @classmethod
    @tracer.trace_func
    def get_interface_name_for_vip(cls):
        """
        获取指定节点上可用于配置vip的网卡名(存在 and 启用 and 非slave)
        """
        cmdstr = "ip addr list"
        (ret, outdata, errdata) = cmdprocess.shell_cmd_not_raise(cmdstr)
        if ret != 0:
            raise Exception('Execute %s failed' % cmdstr)

        lines = cmdprocess.output_to_lines(outdata)

        interface_list = list()
        for each_line in lines:
            reobj = re.search(r"^\d+:\s*(.*?):\s*(.*?)$", each_line)
            if reobj is not None:
                interface_info = reobj.group(2)
                # 过滤DOWN的网卡
                if interface_info.find('state DOWN') != -1:
                    continue
                # 过滤slave网卡
                elif interface_info.find('SLAVE') != -1:
                    continue
                # 过滤lo网卡
                elif interface_info.find('LOOPBACK') != -1:
                    continue
                else:
                    interface_list.append(reobj.group(1).split('@')[0])
        return interface_list

    @classmethod
    @tracer.trace_func
    def get_interface_name_for_bond(cls):
        """
        获取指定节点上可用于配置vip的网卡名(存在 and 启用 and 非slave)
        """
        cmdstr = "ip addr list"
        (ret, outdata, errdata) = cmdprocess.shell_cmd_not_raise(cmdstr)
        if ret != 0:
            raise Exception('Execute %s failed' % cmdstr)

        lines = cmdprocess.output_to_lines(outdata)

        interface_list = list()
        for each_line in lines:
            reobj = re.search(r"^\d+:\s*(.*?):\s*(.*?)$", each_line)
            if reobj is not None:
                interface_info = reobj.group(2)
                # 过滤slave网卡
                if interface_info.find('SLAVE') != -1:
                    continue
                # 过滤lo网卡
                elif interface_info.find('LOOPBACK') != -1:
                    continue
                else:
                    interface_list.append(reobj.group(1).split('@')[0])
        return interface_list

    @classmethod
    @tracer.trace_func
    def get_nics(cls):
        # import pdb
        """
        /**
         * 获取当前系统中的所有活动的网络接口信息
         *
         * @return list<ncTNic>: 网络接口设备列表
         */
        """
        # 查询当前系统的网络设备及IP列表
        cmd = "ip addr list"
        cmd_ret = cmdprocess.shell_cmd_dict(cmd)

        # 解析命令输出
        nics = []
        cur_nic = None
        lines = cmdprocess.output_to_lines(cmd_ret["outmsg"])

        # 获取网关信息
        gateway_infos_inte6 = cls.get_all_gateway("inet6")
        gateway_infos_inte = cls.get_all_gateway("inet")
        for line in lines:
            # pdb.set_trace()
            # 6: bond0: <BROADCAST,MULTICAST,MASTER,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP
            #     link/ether 00:e0:ed:22:ac:48 brd ff:ff:ff:ff:ff:ff
            #     inet 192.168.77.30/24 brd 192.168.77.255 scope global bond0:99
            #        valid_lft forever preferred_lft forever

            # 匹配网卡信息第一行:
            # 6: bond0: <BROADCAST,MULTICAST,MASTER,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP
            # vlan 设备：
            # 6: bond0.100@bond0: <BROADCAST,MULTICAST,MASTER,UP,LOWER_UP>
            #    mtu 1500 qdisc noqueue state UP
            reobj = re.search(r"^\d+:\s*(.*?):\s*(.*?)$", line)
            if reobj is not None:
                if cur_nic is not None:
                    # print cur_nic
                    nics.append(cur_nic)

                cur_nic = deepcopy(ncTNic)
                # pdb.set_trace()
                cur_nic["nic_dev_name"] = reobj.group(1).split('@')[0]
                cur_nic["state_info"] = reobj.group(2)
                if cur_nic["state_info"].find("state UP") != -1:
                    cur_nic["is_up"] = True

            # 匹配网卡信息第二行:
            #     link/ether 00:e0:ed:22:ac:48 brd ff:ff:ff:ff:ff:ff
            reobj = re.search(r"^\s*(link.*?)$", line)
            if reobj is not None:
                cur_nic["hw_info"] = reobj.group(1)

            # 匹配网卡IP地址信息第一行:
            #     inet 192.168.77.30/24 brd 192.168.77.255 scope global bond0:99
            reobj = re.search(r"^\s*inet6?\s+(\S+)\s+(.*?)$", line)
            if reobj is not None:
                ifaddr_subs = reobj.group(1).split('/')
                if len(ifaddr_subs) == 2:
                    cur_ifaddr = deepcopy(ncTIfAddr)
                    label_subs = reobj.group(2).split()[-1].split(":")
                    cur_ifaddr["nic_dev_name"] = cur_nic["nic_dev_name"]
                    if len(label_subs) == 2:
                        cur_ifaddr["label"] = label_subs[1]
                    cur_ifaddr["ipaddr"] = ifaddr_subs[0]
                    ipfamily = "inet6"
                    cur_ifaddr["prefix"] = ifaddr_subs[1]
                    if ":" not in ifaddr_subs[0]:
                        ipfamily = "inet"
                        cur_ifaddr["netmask"] = netlib.exchange_int_to_mask(int(ifaddr_subs[1]))
                    if ipfamily == "inet6":
                        lines = gateway_infos_inte6
                    elif ipfamily == "inet":
                        lines = gateway_infos_inte
                    cur_ifaddr["gateway"] = cls.get_gateway(cur_ifaddr, lines)
                    cur_nic["ifaddrs"].append(cur_ifaddr)

        if cur_nic is not None:
            nics.append(cur_nic)
        return nics

    @classmethod
    @tracer.trace_func
    def get_ifaddr(cls, label):
        """
        /**
         * 获取当前系统中指定标签的协议地址信息
         * 若不存在，则返回 ncTIfAddr.nic_dev_name 为空
         *
         * @param string label: 协议地址的标签，如bond0:1中的‘1’，bond0:inner_vip中的‘inner_vip’
         */
        """
        nics = cls.get_nics()
        for nic in nics:
            for ifaddr in nic["ifaddrs"]:
                if ifaddr["label"] == label:
                    return ifaddr
        return ncTIfAddr

    @classmethod
    @tracer.trace_func
    def get_all_gateway(cls, ipfamily="inet"):
        """
        获取指定网络地址的默认网关配置
        @return list gateway""
        """
        cmd = "ip -family {} route list".format(ipfamily)
        cmd_ret = cmdprocess.shell_cmd_dict(cmd)

        # 解析命令输出
        lines = cmdprocess.output_to_lines(cmd_ret["outmsg"])
        return lines

    @classmethod
    @tracer.trace_func
    def get_gateway(cls, ifaddr, lines):
        """
        获取指定网络地址的默认网关配置
        @return str gateway     若未找到，则返回""
        """
        # cmd = "ip -family {} route list".format(ipfamily)
        # cmd_ret = cmdprocess.shell_cmd_dict(cmd)

        # 解析命令输出
        # lines = cmdprocess.output_to_lines(cmd_ret["outmsg"])
        for line in lines:
            # 匹配： default via 192.168.77.1 dev bond0
            reobj = re.search(r"default\s+via\s+(\S+)\s+dev\s+(\S+)\s+", line)
            if reobj is not None:
                gateway = reobj.group(1)
                nic_dev_name = reobj.group(2)
                # ipaddr = reobj.group(3)
                if nic_dev_name == ifaddr["nic_dev_name"]:
                    return gateway
        # 未找到
        return ""

    @classmethod
    @tracer.trace_func
    def get_ifaddr_by_ipaddr(cls, ipaddr):
        """
        /**
         * 获取当前系统中指定 IP 的协议地址信息
         * 若不存在，则返回 ncTIfAddr.nic_dev_name 为空
         *
         * @param string ipaddr: 协议地址的 IP
         */
        """
        nics = cls.get_nics()
        for nic in nics:
            for ifaddr in nic["ifaddrs"]:
                if ifaddr["ipaddr"] == ipaddr:
                    return ifaddr
        return ncTIfAddr

    @classmethod
    @tracer.trace_func
    def del_ifaddr(cls, label):
        """
        /**
         * 删除指定标签的协议地址(持久化的)
         *
         * @param string label:    协议地址的标签，如bond0:1中的‘1’，bond0:inner_vip中的‘inner_vip’
         */
        """
        logger.syslog(MODULE_NAME, "Delete ifaddr {0} begin.".format(label))

        # 删除指定协议地址及其默认路由
        while True:
            ifaddr = cls.get_ifaddr(label)
            if ifaddr["nic_dev_name"] == "":
                break

            # down掉所属某个子网的primary ip的时候，所有相关的secondary ip也会down掉。
            # 这个可以通过设置一个内核参数，当primary ip宕掉时可以将secondary ip提升为primary ip。
            cmd = 'echo "1" > /proc/sys/net/ipv4/conf/{0}/promote_secondaries'.format(
                ifaddr["nic_dev_name"])
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

            if ifaddr["gateway"] != "":
                # 删除默认网关
                cmd = "ip route del default via {0} dev {1} src {2}".format(
                    ifaddr["gateway"], ifaddr["nic_dev_name"], ifaddr["ipaddr"])
                (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
                logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

            cmd = "ip addr del {1}/{2} dev {0}".format(
                ifaddr["nic_dev_name"], ifaddr["ipaddr"], ifaddr["prefix"])
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

        # 删除指定label的网络配置文件
        for name in os.listdir(IFCFG_DIR_PATH):
            if name.find(":{0}".format(label)) != -1:
                ifcfg_path = os.path.join(IFCFG_DIR_PATH, name)
                os.remove(ifcfg_path)
                logger.syslog(MODULE_NAME, "Deleted {0}.".format(ifcfg_path))

        logger.syslog(MODULE_NAME, "Delete ifaddr {0} end.".format(label))

    @classmethod
    @tracer.trace_func
    def set_ifaddr(cls, ifaddr):
        """
        /**
         * 在指定接口设备上配置协议地址(持久化的)
         *
         * @param ncTIfAddr ifaddr:  协议地址配置
         */
        """
        logger.syslog(MODULE_NAME, "Set ifaddr {0} begin.".format(ifaddr))

        # 先删除
        cls.del_ifaddr(ifaddr["label"])

        # 再添加 IP
        cls._set_ifaddr(ifaddr)

        # 保存配置文件
        ifcfg_path = cls._join_ifcfg_path(ifaddr["nic_dev_name"], ifaddr["label"])
        confs_new = {}
        confs_new["DEVICE"] = cls._join_ifaddr_label(ifaddr["nic_dev_name"], ifaddr["label"])
        ip_ver = netlib.get_ip_version(ifaddr["ipaddr"])
        if ip_ver == -1:
            raise Exception("invalid ip address: {0}".format(ifaddr["ipaddr"]))
        elif ip_ver == 4:   # ipv4 网卡配置
            confs_new["IPADDR"] = ifaddr["ipaddr"]
            confs_new["NETMASK"] = ifaddr["netmask"]
            if ifaddr["gateway"] != "":
                confs_new["GATEWAY"] = ifaddr["gateway"]
        else:   # ipv6 网卡配置
            confs_new["IPV6ADDR"] = "/".join(ifaddr["ipaddr"], ifaddr["prefix"])
            if ifaddr["gateway"] != "":
                confs_new["IPV6_DEFAULTGW"] = ifaddr["gateway"]
            confs_new["IPV6INIT"] = 'yes'
            confs_new["IPV6_AUTOCONF"] = 'yes'
            confs_new["IPV6_DEFROUTE"] = 'yes'
            confs_new["IPV6_FAILURE_FATAL"] = 'no'
            confs_new["IPV6_ADDR_GEN_MODE"] = 'stable-privacy'
        confs_new["ONBOOT"] = 'yes'
        confs_new["BOOTPROTO"] = 'static'
        cls._save_ifcfg(ifcfg_path, confs_new)

        logger.syslog(MODULE_NAME, "Set ifaddr {0} end.".format(ifaddr))

    @classmethod
    @tracer.trace_func
    def bind_nics(cls, nic_name_list):
        """
        根据参数进行物理绑定
        绑定网卡名按照: bond0, bond1...依次递增
        """
        logger.syslog(MODULE_NAME, "bind nics begin.")

        # 初始值为-1
        bond_num = [-1]
        # 计算网卡名
        for each_nic in cls.get_nics():
            reobj = re.match(r'bond(\d)', each_nic["nic_dev_name"])
            if reobj:
                bond_num.append(int(reobj.group(1)))
        bond_name = BOND_NAME % (cls._find_first_missing_position(bond_num))

        # 如果是 8/9 系列的系统已不再支持network服务，需要使用nmcli来配置bond，这里简单检查network服务是否启动，如果未启动则使用nmcli来配置
        if not cls.is_services_started("network"):
            logger.syslog(MODULE_NAME, "Set master {0} begin.".format(bond_name))
            cmd = 'nmcli connection add type bond con-name {0} ifname {0} bond.options "mode=balance-alb,miimon=100"'.format(bond_name)
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)
            cmd = 'nmcli connection modify {0} connection.autoconnect-slaves 1'.format(bond_name)
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)
            cmd = 'nmcli connection up {0}'.format(bond_name)
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)
            logger.syslog(MODULE_NAME, "Set slave {1} begin.".format(bond_name, nic_name_list))
            for nic in nic_name_list:
                cmd = 'nmcli c modify {0} master {1}'.format(nic, bond_name)
                (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
                logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

                cmd = 'nmcli c up {0}'.format(nic)
                (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
                logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)
            logger.syslog(MODULE_NAME, "Set slave {1} end.".format(bond_name, nic_name_list))
        else:
            # 配置 bond master
            logger.syslog(MODULE_NAME, "Set master {0} begin.".format(bond_name))
            nic_cfg = []
            nic_cfg.append("DEVICE=\"%s\"" % bond_name)
            nic_cfg.append("ONBOOT=\"yes\"")
            nic_cfg.append("BONDING_OPTS=\"mode=6 miimon=100\"")
            nic_cfg.append("BOOTPROTO=\"static\"")
            path = NIC_CFG_FILE_PATH % bond_name
            filelib.write_file(path, nic_cfg)
            logger.syslog(MODULE_NAME, "Set master {0} end.".format(bond_name))

            # 配置 bond slave
            logger.syslog(MODULE_NAME, "Set slave {1} begin.".format(bond_name, nic_name_list))
            for nic in nic_name_list:
                # 备份网卡设置,解绑时进行恢复
                nic_cfg = NIC_CFG_FILE_PATH % nic
                nic_cfg_bak = NIC_CFG_FILE_PATH % nic + ".bak"
                if os.path.exists(nic_cfg):
                    shutil.copy(nic_cfg, nic_cfg_bak)
                    logger.syslog(MODULE_NAME, "copy {0} {1}.".format(nic_cfg, nic_cfg_bak))
                nic_cfg = []
                nic_cfg.append("DEVICE=\"%s\"" % nic)
                nic_cfg.append("ONBOOT=\"yes\"")
                nic_cfg.append("BOOTPROTO=\"static\"")
                nic_cfg.append("MASTER=\"%s\"" % bond_name)
                nic_cfg.append("SLAVE=\"yes\"")
                nic_cfg.append("ETHTOOL_OPTS=\"%s\"" % ETHTOOL_OPTS % nic)
                filelib.write_file(NIC_CFG_FILE_PATH % nic, nic_cfg)
            logger.syslog(MODULE_NAME, "Set slave {1} end.".format(bond_name, nic_name_list))

            # 重启网络
            cmd = "systemctl restart network"
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

        logger.syslog(MODULE_NAME, "bind nics end.")

    @classmethod
    @tracer.trace_func
    def unbind_nic(cls, bond_dev_name):
        """
        解除绑定的逻辑网卡,还原物理网卡配置
        """
        logger.syslog(MODULE_NAME, "unbind nic begin.")
        # 获取绑定的物理网卡列表
        nic_list = cls._get_bond_slaves(bond_dev_name)

        # 如果是 8/9 系列的系统已不再支持network服务，需要使用nmcli来配置bond，这里简单检查network服务是否启动，如果未启动则使用nmcli来配置
        if not cls.is_services_started("network"):
            for nic in nic_list:
                cmd = "nmcli c delete {0}".format(nic)
                (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
                logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)
            
            cmd = "nmcli c delete {0}".format(bond_dev_name)
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)
        else:
            # 删除逻辑网卡配置文件
            bond_file = NIC_CFG_FILE_PATH % bond_dev_name
            if os.path.exists(bond_file):
                os.remove(bond_file)
                logger.syslog(MODULE_NAME, "remove bonding file {0}.".format(bond_file))

            
            logger.syslog(MODULE_NAME, "bonding slaves: {0}.".format(nic_list))
            # 还原物理网卡配置文件
            for nic in nic_list:
                nic_cfg_bak = "".join([NIC_CFG_FILE_PATH % nic, ".bak"])
                nic_cfg = NIC_CFG_FILE_PATH % nic

                # 若备份文件存在则恢复
                if os.path.exists(nic_cfg_bak):
                    shutil.move(nic_cfg_bak, nic_cfg)
                    logger.syslog(MODULE_NAME, "move {0} {1}.".format(nic_cfg_bak, nic_cfg))

            # 重启网络
            # cmd = "systemctl restart network"
            # (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            # logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

        logger.syslog(MODULE_NAME, "unbind nic end.")

    @classmethod
    def reload_network(self):
        # 重启网络
        cmd = "systemctl restart network"
        (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
        logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

    @classmethod
    @cmdprocess.check_param_security_for_shell
    @tracer.trace_func
    def exists_arp(cls, ipaddr):
        """参见 API 说明"""
        # 使用subprocess参数列表避免命令注入
        # 首先获取所有ARP条目
        cmd_args = ['arp', '-an']
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        outmsg_str = outmsg.decode('utf-8', errors='ignore')
        
        # 在Python中进行过滤，而不是使用shell管道
        outlines = outmsg_str.splitlines()
        matched_lines = [line for line in outlines if ipaddr in line and 'incomplete' not in line]
        
        if len(matched_lines) <= 0:
            return False
        else:
            return True

    @classmethod
    @cmdprocess.check_param_security_for_shell
    @tracer.trace_func
    def del_arp(cls, ipaddr):
        """参见 API 说明"""
        logger.syslog(MODULE_NAME, "Delete arp {0} begin.".format(ipaddr))
        
        # 使用subprocess参数列表避免命令注入
        # 首先获取所有ARP条目
        cmd_args = ['arp', '-an']
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        outmsg_str = outmsg.decode('utf-8', errors='ignore')
        
        # 在Python中进行过滤，而不是使用shell管道
        outlines = outmsg_str.splitlines()
        matched_lines = [line for line in outlines if ipaddr in line]
        
        if len(matched_lines) <= 0:
            logger.syslog(MODULE_NAME, "Not found arp host {0}.".format(ipaddr))
            return

        # 尝试删除指定主机 arp 映射信息,出现异常则忽略
        cmd_args = ['arp', '-d', ipaddr]
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        returncode = process.returncode
        
        # 为了保持日志格式一致，构造命令字符串用于日志记录
        cmd = "arp -d '{0}'".format(ipaddr)
        logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg, returncode)

        logger.syslog(MODULE_NAME, "Delete arp {0} end.".format(ipaddr))

    ########################################################################################
    # 以下为内部公共函数

    @classmethod
    @cmdprocess.check_param_security_for_shell
    @tracer.trace_func
    def _set_ifaddr(cls, ifaddr):
        """使用 ip 命令配置 IP"""
        # 添加 IP
        prefix = ifaddr.get("prefix", "")
        if netlib.get_ip_version(ifaddr["ipaddr"]) == 4:
            prefix = netlib.exchange_mask_to_int(ifaddr["netmask"])
        cmd = "ip addr add {2}/{3} brd + label {1} dev {0}".format(
            ifaddr["nic_dev_name"],
            cls._join_ifaddr_label(ifaddr["nic_dev_name"], ifaddr["label"]),
            ifaddr["ipaddr"],
            prefix)
        (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
        logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

        # 添加默认网关
        if ifaddr["gateway"]:
            cmd = "ip route add default via {0} dev {1} src {2}".format(
                ifaddr["gateway"], ifaddr["nic_dev_name"], ifaddr["ipaddr"])
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

    @classmethod
    @tracer.trace_func
    def _save_ifcfg(cls, ifcfg_path, confs):
        """
        保存网卡配置
        @ifcfg_path  网卡配置文件路径
        @confs      网卡具体配置，如： {"DEVICE": "bond0:99",
                                        "IPADDR": "192.168.77.31",
                                        "NETMASK": "255.255.255.0",
                                        ...
                                        }
        """
        ifcfg_buf = ""
        for key, value in confs.iteritems():
            ifcfg_buf += '{0}="{1}"\n'.format(key, value)
        filelib.write_file(path=ifcfg_path, content=ifcfg_buf)
        logger.syslog(MODULE_NAME, "Saved {0} with {1}.".format(ifcfg_path, confs))

    @classmethod
    @tracer.trace_func
    def _join_ifaddr_label(cls, nic_dev_name, label):
        """组合标签设备名"""
        if label != "":
            return "{0}:{1}".format(nic_dev_name, label)
        else:
            return nic_dev_name

    @classmethod
    @tracer.trace_func
    def _join_ifcfg_path(cls, nic_dev_name, label):
        """组合网卡配置文件路径"""
        ifaddr_label = cls._join_ifaddr_label(nic_dev_name, label)
        ifcfg_name = "ifcfg-{0}".format(ifaddr_label)
        ifcfg_path = os.path.join(IFCFG_DIR_PATH, ifcfg_name)
        return ifcfg_path

    @classmethod
    @tracer.trace_func
    def _get_bond_slaves(cls, bond_name):
        """
        获取指定绑定设备物理网卡
        """
        bond_file = "/proc/net/bonding/%s" % bond_name

        slave_list = []
        if os.path.exists(bond_file):
            # 读取文件,获取bond slave 网卡列表
            cmd = "cat %s | grep 'Slave Interface:'" % (bond_file)
            (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
            logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

            # 将返回转换为列表
            lines = cmdprocess.output_to_lines(outmsg)
            for line in lines:
                reobj = re.search(r"^Slave Interface:\s*(\w+)", line)
                if reobj is not None:
                    slave_list.append(reobj.group(1))
        return slave_list

    @classmethod
    @cmdprocess.check_param_security_for_shell
    @tracer.trace_func
    def _stop_nic(cls, nic_dev_name, label):
        """
        停止指定网卡
        """
        ifaddr_label = cls._join_ifaddr_label(nic_dev_name, label)
        cmd = "ip link set dev {0} down".format(ifaddr_label)
        (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
        logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

    @classmethod
    @cmdprocess.check_param_security_for_shell
    @tracer.trace_func
    def _start_nic(cls, nic_dev_name, label):
        """
        启动指定网卡
        """
        ifaddr_label = cls._join_ifaddr_label(nic_dev_name, label)
        cmd = "ip link set dev {0} up".format(ifaddr_label)
        (outmsg, errmsg) = cmdprocess.shell_cmd(cmd)
        logger.syslog_cmd(MODULE_NAME, cmd, outmsg, errmsg)

    @classmethod
    @tracer.trace_func
    def _find_first_missing_position(cls, num_list):
        """查找绑定网卡第一个缺失的编号"""
        num_list.sort()
        n = len(num_list)
        for i in range(n):
            if i != num_list[i] + 1:
                return i - 1
        return n - 1

    @classmethod
    def is_services_started(cls, service_name):
        # 使用subprocess参数列表避免命令注入
        cmd_args = ['systemctl', 'status', service_name]
        process = subprocess.Popen(cmd_args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        outmsg, errmsg = process.communicate()
        returncode = process.returncode
        if returncode:
            return False
        else:
            return True
