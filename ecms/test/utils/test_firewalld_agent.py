#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/27 11:38
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    : 
# @File    : test_firewalld_agent_new.py
# @Software: PyCharm
import os
import sys
import unittest

from mock import patch, MagicMock

CURR_SCRIPT_PATH = os.path.dirname(os.path.realpath(__file__))
SRC_PATH = os.path.dirname(os.path.dirname(CURR_SCRIPT_PATH))
sys.path.append(SRC_PATH)

from src.modules.ecms_agent.firewall_agent import FirewallAgent


class FirewalldTest(unittest.TestCase):

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    def test_add_rich_rule_false(self, mock_shell_cmd):
        rich_rule_list = ['rule family="ipv4" source address="10.0.0.0/25" destination '
                          'address="192.168.0.10/32" port port="8080-8090" protocol="tcp" accept']
        zone = "public"
        is_permanent = False
        FirewallAgent.add_rich_rule(rich_rule_list, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    def test_add_rich_rule_true(self, mock_shell_cmd):
        rich_rule_list = ['rule family="ipv4" source address="10.0.0.0/25" destination '
                          'address="192.168.0.10/32" port port="8080-8090" protocol="tcp" accept']
        zone = "public"
        is_permanent = True
        FirewallAgent.add_rich_rule(rich_rule_list, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    def test_remove_rich_rule_false(self, mock_shell_cmd):
        rich_rule_list = ['rule family="ipv4" source address="10.0.0.0/25" destination '
                          'address="192.168.0.10/32" port port="8080-8090" protocol="tcp" accept']
        zone = "public"
        is_permanent = False
        FirewallAgent.remove_rich_rule(rich_rule_list, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    def test_remove_rich_rule_true(self, mock_shell_cmd):
        rich_rule_list = ['rule family="ipv4" source address="10.0.0.0/25" destination '
                          'address="192.168.0.10/32" port port="8080-8090" protocol="tcp" accept']
        zone = "public"
        is_permanent = True
        FirewallAgent.remove_rich_rule(rich_rule_list, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    def test_add_source_false(self, mock_shell_cmd):
        source = "192.0.2.94"
        zone = "public"
        is_permanent = False
        FirewallAgent.add_source(source, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="xml.etree.cElementTree.ElementTree")
    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    def test_add_source_true(self, mock_shell_cmd, mock_element_tree):
        source = "192.0.2.94"
        zone = "public"
        is_permanent = True
        FirewallAgent.add_source(source, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd_not_raise", return_value=(0, 0, ""))
    def test_remove_source_false(self, mock_shell_cmd):
        source = "192.0.2.94"
        zone = "public"
        is_permanent = False
        FirewallAgent.remove_source(source, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="xml.etree.cElementTree.ElementTree")
    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd_not_raise", return_value=(0, 0, ""))
    def test_remove_source_true(self, mock_shell_cmd, mock_element_tree):
        source = "192.0.2.94"
        zone = "public"
        is_permanent = True
        FirewallAgent.remove_source(source, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    def test_remove_service(self, mock_shell_cmd):
        service_name = "ssh"
        zone = "public"
        is_permanent = False
        FirewallAgent.remove_service(service_name, zone, is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="os.path.exists", return_value=False)
    def test_get_firewall_info_false(self, mock_exists):
        option = "service"
        zone = "test"
        is_permanent = False
        FirewallAgent.get_firewall_info(option, zone, is_permanent)
        mock_exists.assert_called()

    @patch("src.modules.pydeps.cmdprocess.output_to_lines", return_value="")
    @patch("src.modules.pydeps.cmdprocess.shell_cmd", return_value=("", ""))
    @patch(target="xml.etree.cElementTree.ElementTree")
    @patch(target="os.path.exists", return_value=True)
    def test_get_firewall_info_true_service(self, mock_exists, mock_element_tree, mock_shell_cmd, mock_output_to_lines):
        option = "service"
        zone = "test"
        is_permanent = True
        FirewallAgent.get_firewall_info(option, zone, is_permanent)
        mock_exists.assert_called()
        mock_shell_cmd.assert_called_once()
        mock_output_to_lines.assert_called_once()

    @patch("src.modules.pydeps.cmdprocess.output_to_lines", return_value="")
    @patch("src.modules.pydeps.cmdprocess.shell_cmd", return_value=("", ""))
    @patch(target="xml.etree.cElementTree.ElementTree")
    @patch(target="os.path.exists", return_value=True)
    def test_get_firewall_info_true_source(self, mock_exists, mock_element_tree, mock_shell_cmd, mock_output_to_lines):
        option = "source"
        zone = "test"
        is_permanent = False
        FirewallAgent.get_firewall_info(option, zone, is_permanent)
        mock_exists.assert_called()
        mock_shell_cmd.assert_called_once()
        mock_output_to_lines.assert_called_once()

    @patch("src.modules.pydeps.cmdprocess.output_to_lines", return_value="")
    @patch("src.modules.pydeps.cmdprocess.shell_cmd", return_value=("", ""))
    @patch(target="xml.etree.cElementTree.ElementTree")
    @patch(target="os.path.exists", return_value=True)
    def test_get_firewall_info_true_rich_rule(self, mock_exists, mock_element_tree, mock_shell_cmd, mock_output_to_lines):
        option = "rich-rule"
        zone = "test"
        is_permanent = False
        FirewallAgent.get_firewall_info(option, zone, is_permanent)
        mock_exists.assert_called()
        mock_shell_cmd.assert_called_once()
        mock_output_to_lines.assert_called_once()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=('', ''))
    def test_get_target(self, mock_shell_cmd):
        zone = "public"
        FirewallAgent.get_target(zone)
        mock_shell_cmd.assert_called()

    # def test_set_target(self, ):
    #     zone = "public"
    #     option = "default"
    #     FirewallAgent.set_target(option, zone)

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=('', ''))
    def test_get_default_zone(self, mock_shell_cmd):
        FirewallAgent.get_default_zone()
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=('', ''))
    def test_set_default_zone(self, mock_shell_cmd):
        zone = "public"
        FirewallAgent.set_default_zone(zone)
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    def test_reload_firewall(self, mock_shell_cmd):
        is_permanent = False
        FirewallAgent.reload_firewall(is_permanent)
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.filelib.write_file")
    def test_init_firewall_xml(self, mock_write_file):
        FirewallAgent.init_firewall_xml()
        mock_write_file.assert_called()
