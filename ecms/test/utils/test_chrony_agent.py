#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/5/7 16:55
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    : 
# @File    : test_chrony_agent.py
# @Software: PyCharm
import os
import sys
import unittest

from mock import patch

CURR_SCRIPT_PATH = os.path.dirname(os.path.realpath(__file__))
SRC_PATH = os.path.dirname(os.path.dirname(CURR_SCRIPT_PATH))
sys.path.append(SRC_PATH)

from src.modules.ecms_agent.chrony_agent import ChronyAgent


class ChronyTest(unittest.TestCase):

    @patch(target="src.modules.pydeps.filelib.write_file")
    def test_set_chrony_server(self, mock_write_file):
        ChronyAgent.set_chrony_server()
        mock_write_file.assert_called()

    @patch(target="src.modules.pydeps.filelib.write_file")
    def test_set_chrony_client(self, mock_write_file):
        ChronyAgent.set_chrony_client("1.1.1.1")
        mock_write_file.assert_called()

    @patch(target="src.modules.pydeps.filelib.read_file", return_value="#master node")
    def test_get_chrony_role_master(self, mock_read_file):
        role = ChronyAgent.get_chrony_role()
        mock_read_file.assert_called()
        assert role == 1

    @patch(target="src.modules.pydeps.filelib.read_file", return_value="#slave node")
    def test_get_chrony_role_slave(self, mock_read_file):
        role = ChronyAgent.get_chrony_role()
        mock_read_file.assert_called()
        assert role == 2

    @patch(target="src.modules.pydeps.filelib.read_file", return_value="unknown")
    def test_get_chrony_role_unknow(self, mock_read_file):
        role = ChronyAgent.get_chrony_role()
        mock_read_file.assert_called()
        assert role == 0

    @patch(target="src.modules.pydeps.filelib.write_file")
    @patch(target="src.modules.pydeps.filelib.read_file")
    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=("", ""))
    def test_add_time_server(self, mock_shell_cmd, mock_read_file, mock_write_file):
        ChronyAgent.add_time_server("1.1.1.1")
        mock_shell_cmd.assert_called()
        mock_read_file.assert_called()
        mock_write_file.assert_called()

    @patch(target="src.modules.pydeps.filelib.write_file")
    @patch(target="src.modules.pydeps.filelib.read_file")
    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=("", ""))
    def test_del_time_server(self, mock_shell_cmd, mock_read_file, mock_write_file):
        ChronyAgent.del_time_server("1.1.1.1")
        mock_shell_cmd.assert_called()
        mock_read_file.assert_called()
        mock_write_file.assert_called()

    @patch(target="src.modules.pydeps.filelib.write_file")
    def test_clear_chrony_config(self, mock_write_file):
        ChronyAgent.clear_chrony_config()
        mock_write_file.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=("System time     : 0.000011110 "
                                                                           "seconds fast of NTP time", ""))
    def test_get_diff_from_ref_true(self, mock_shell_cmd):
        ChronyAgent.get_diff_from_ref()
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=("", ""))
    def test_get_diff_from_ref_error(self, mock_shell_cmd):
        with self.assertRaises(Exception):
            ChronyAgent.get_diff_from_ref()
        mock_shell_cmd.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd", return_value=("", ""))
    def test_makestep(self, mock_shell_cmd):
        ChronyAgent.makestep()
        mock_shell_cmd.assert_called()