#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/12 11:18
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    : 
# @File    : test_firewalld_agent.py
# @Software: PyCharm
import os
import sys
import unittest

import mock
from mock import patch

CURR_SCRIPT_PATH = os.path.dirname(os.path.realpath(__file__))
SRC_PATH = os.path.dirname(os.path.dirname(CURR_SCRIPT_PATH))
sys.path.append(SRC_PATH)

from src.modules.ecms_agent.net_agent import NetAgent, ncTIfAddr
from src.modules.pydeps import netlib
from src.modules.pydeps import cmdprocess


nics_info = [
    {'nic_dev_name': 'lo', 'state_info': '<LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1',
     'ifaddrs': [{'nic_dev_name': 'lo', 'netmask': '255.0.0.0', 'ipaddr': '127.0.0.1', 'gateway': '', 'label': '', 'prefix': '8'}],
     'is_up': False, 'hw_info': 'link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00'},
    {'nic_dev_name': 'ens192', 'state_info': '<BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP qlen 1000',
     'ifaddrs': [{'nic_dev_name': 'ens192', 'netmask': '255.255.255.0', 'ipaddr': '192.0.2.90', 'gateway':
                  '192.0.2.254', 'label': '', 'prefix': '24'}], 'is_up': True, 'hw_info': 'link/ether 00:50:56:82:63:55 brd ff:ff:ff:ff:ff:ff'},
    {'nic_dev_name': 'docker0', 'state_info': '<BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP ',
     'ifaddrs': [{'nic_dev_name': 'docker0', 'netmask': '255.255.0.0', 'ipaddr': '172.17.0.1', 'gateway': '',
                  'label': '', 'prefix': '16'}], 'is_up': True, 'hw_info': 'link/ether 02:42:94:c7:cb:95 brd ff:ff:ff:ff:ff:ff'},
    {'nic_dev_name': 'tunl0', 'state_info': '<NOARP,UP,LOWER_UP> mtu 1440 qdisc noqueue state UNKNOWN qlen 1',
     'ifaddrs': [{'nic_dev_name': 'tunl0', 'netmask': '255.255.255.255', 'ipaddr': '192.169.219.64', 'gateway': '',
                  'label': '', 'prefix': '32'}], 'is_up': False, 'hw_info': 'link/ipip 0.0.0.0 brd 0.0.0.0'},
    {'nic_dev_name': 'vethd0e2003', 'state_info': '<BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc '
                                                  'noqueue master docker0 state UP ', 'ifaddrs': [], 'is_up': True,
     'hw_info': 'link/ether 0a:c8:19:67:42:27 brd ff:ff:ff:ff:ff:ff link-netnsid 0'},
    {'nic_dev_name': 'cali0bc1e7c680b', 'state_info': '<BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1440 qdisc '
                                                      'noqueue state UP ', 'ifaddrs': [], 'is_up': True, 'hw_info':
        'link/ether ee:ee:ee:ee:ee:ee brd ff:ff:ff:ff:ff:ff link-netnsid 1'},
    {'nic_dev_name': 'cali6f16cbd0132', 'state_info': '<BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1440 '
                                                      'qdisc noqueue state UP ', 'ifaddrs': [], 'is_up': True,
     'hw_info': 'link/ether ee:ee:ee:ee:ee:ee brd ff:ff:ff:ff:ff:ff link-netnsid 4'}]


# net_agent 模块相关接口单元测试类
class TestNetAgent(unittest.TestCase):
    """
    net_agent ut 测试类
    """

    def setUp(self):
        """
        每个用例前执行
        """
        pass

    def tearDown(self):
        """
        每个用例执行后执行
        """
        pass


class TestGet(TestNetAgent):
    """
    测试获取网卡
    """
    # 测试获取IP地址
    def test_get_ips(self):
        ips = NetAgent.get_ip_addrs()
        self.assertTrue(len(ips) > 0)

    def test_get_interface_name_for_vip(self):
        inf_names = NetAgent.get_interface_name_for_vip()
        self.assertTrue(len(inf_names) > 0)

    def test_get_nics_info(self):
        nics = NetAgent.get_nics()
        self.assertTrue(len(nics) > 0)

    # 测试通过网卡标签获取IP地址
    def test_get_ip_by_label(self):
        ip = NetAgent.get_ifaddr('iv')
        self.assertIs(ip, ncTIfAddr)

    # 测试通过ip获取网卡信息
    def test_get_ifaddr_by_ipaddr(self):
        nic_info = NetAgent.get_ifaddr_by_ipaddr('1.1.1.1')
        self.assertIs(nic_info, ncTIfAddr)

    def test_exists_arp(self):
        flag = NetAgent.exists_arp("1.1.1.1")
        self.assertFalse(flag)

    # 测试获取指定网卡网关地址
    def test_get_gw_by_nic(self):
        lines = NetAgent.get_all_gateway("inet")
        gw = NetAgent.get_gateway({"nic_dev_name": "xxx"}, lines)
        self.assertIs(gw, '')
        gw = NetAgent.get_gateway({"nic_dev_name": "lo"}, lines)
        self.assertIs(gw, '')


class TestSet(TestNetAgent):
    """
    测试设置网卡
    """
    @patch("src.modules.ecms_agent.net_agent.NetAgent._save_ifcfg")
    @patch("src.modules.ecms_agent.net_agent.NetAgent._set_ifaddr")
    @patch("src.modules.ecms_agent.net_agent.NetAgent.del_ifaddr")
    def test_set_ifaddr(self, mock_del_ifaddr, mock__set_ifaddr, mock__save_ifcfg):
        ifaddr = dict()
        ifaddr["nic_dev_name"] = "lo"
        ifaddr["netmask"] = "255.255.255.255"
        ifaddr["ipaddr"] = "192.168.139.139"
        ifaddr["gateway"] = ""
        ifaddr["label"] = "vip"
        NetAgent.set_ifaddr(ifaddr)
        mock_del_ifaddr.assert_called_once()
        mock__set_ifaddr.assert_called_once()
        mock__save_ifcfg.assert_called_once()

    # @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=('', ''))
    # @patch(target="src.modules.pydeps.filelib.write_file")
    # def test_bind_nics(self, mock_write_file, mock_shell_cmd):
    #     nic_name_list = ["test01", "test02"]
    #     NetAgent.bind_nics(nic_name_list)
    #     self.assertEqual(mock_write_file.call_count, 3)
    #     mock_shell_cmd.assert_called()

    # @patch(target="src.modules.ecms_agent.net_agent.NetAgent._get_bond_slaves", return_value=['test'])
    # def test_unbind_nic(self, mock_get_bond):
    #     NetAgent.unbind_nic("bond0")
    #     mock_get_bond.assert_called()


class TestDel(TestNetAgent):
    """
    测试删除网卡
    """

    @patch(target="os.remove")
    @patch(target="os.listdir", return_value=["a:test", "b:test"])
    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", side_effect=[('', ''), ('', ''), ('', ''), ('', '')])
    @patch(target="src.modules.ecms_agent.net_agent.NetAgent.get_ifaddr", side_effect=(nics_info[1]["ifaddrs"][0],
                                                                                       ncTIfAddr))
    def test_del_ifaddr(self, mock_get_ifaddr, mock_shell_cmd, mock_listdir, mock_remove):
        NetAgent.del_ifaddr("test")
        mock_get_ifaddr.assert_called()
        mock_listdir.assert_called_once()
        self.assertEqual(2, mock_remove.call_count)
        mock_shell_cmd.assert_called()

    def test_del_arp(self):
        self.assertIsNone(NetAgent.del_arp("1.1.1.1"))


class TestNetLib(TestNetAgent):
    @mock.patch.object(cmdprocess, "shell_cmd_not_raise", create=True, return_value=(0, "", ""))
    def test_netlib_arping(self, mock0):
        self.assertIsNone(netlib.arping("eth0", 4, "192.168.139.11", "192.168.139.1"))


if __name__ == '__main__':
    unittest.main()
