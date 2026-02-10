#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/5/7 16:55
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    : 
# @File    : test_systemd_agent.py
# @Software: PyCharm
import os
import sys
import unittest

from mock import patch
from mock import mock_open

CURR_SCRIPT_PATH = os.path.dirname(os.path.realpath(__file__))
SRC_PATH = os.path.dirname(os.path.dirname(CURR_SCRIPT_PATH))
sys.path.append(SRC_PATH)

from src.modules.ecms_agent.system_agent import SystemAgent


class TestSystemAgent(unittest.TestCase):

    @patch("src.modules.pydeps.filelib.read_file", return_value=[])
    @patch("src.modules.ecms_agent.system_agent.SystemAgent.list_to_dict", return_value={})
    @patch("os.listdir", return_value=["test01.conf", "test02.conf"])
    def test_get_sysctl_lasting_parameters(self, mock_listdir, mock__list_to_dict, mock_read_file):
        all_conf_dict = SystemAgent.get_sysctl_lasting_parameters()
        mock_listdir.assert_called()
        mock__list_to_dict.assert_called()
        mock_read_file.assert_called()
        self.assertEqual(all_conf_dict, {})

    @patch("src.modules.ecms_agent.system_agent.SystemAgent.list_to_dict", return_value={})
    @patch("src.modules.pydeps.cmdprocess.shell_cmd", return_value=("", ""))
    def test_get_sysctl_parameters(self, mock_shell_cmd, mock__list_to_dict):
        all_conf_dict = SystemAgent.get_sysctl_parameters()
        mock_shell_cmd.assert_called_once()
        mock__list_to_dict.assert_called_once()
        self.assertEqual(all_conf_dict, {})

    def test_create_sysctl_parameters_exception(self):
        with self.assertRaises(Exception):
            SystemAgent.create_sysctl_parameters([])

    @patch("src.modules.pydeps.cmdprocess.shell_cmd")
    @patch("src.modules.pydeps.cmdprocess.shell_cmd_not_raise", return_value=(1, "", ""))
    def test_create_sysctl_parameters_false(self, mock_shell_cmd_not_raise, mock_shell_cmd):
        with self.assertRaises(Exception):
            SystemAgent.create_sysctl_parameters({"a": 1})
        mock_shell_cmd_not_raise.assert_called_once()
        self.assertEqual(2, mock_shell_cmd.call_count)

    @patch(target="src.modules.pydeps.filelib.write_file")
    @patch("src.modules.ecms_agent.system_agent.SystemAgent.dict_to_list", return_value={})
    @patch("src.modules.ecms_agent.system_agent.SystemAgent.list_to_dict", return_value={})
    @patch(target="src.modules.pydeps.filelib.read_file", return_value=[])
    @patch("os.path.isfile", return_value=True)
    @patch("src.modules.pydeps.cmdprocess.shell_cmd_not_raise", return_value=(0, "", ""))
    def test_create_sysctl_parameters_true(self, mock_shell_cmd_not_raise, mock_isfile, mock_read_file,
                                           mock__list_to_dict, mock__dict_to_list, mock_write_file):
        SystemAgent.create_sysctl_parameters({"a": 1})
        mock_shell_cmd_not_raise.assert_called_once()
        mock_isfile.assert_called_once()
        mock_read_file.assert_called_once()
        mock__list_to_dict.assert_called_once()
        mock__dict_to_list.assert_called_once()
        mock_write_file.assert_called_once()

    @patch("src.modules.pydeps.cmdprocess.shell_cmd")
    @patch("src.modules.pydeps.cmdprocess.shell_cmd_not_raise", return_value=(1, "", ""))
    @patch("__builtin__.open", new_callable=mock_open())
    @patch("os.path.isfile", return_value=False)
    def test_update_sysctl_parameters_except(self, mock_isfile, mock_mopen, mock_shell_cmd_not_raise, mock_shell_cmd):
        with self.assertRaises(Exception):
            SystemAgent.update_sysctl_parameters({"a": 1})
        mock_isfile.assert_called_once()
        mock_mopen.assert_called_once()
        mock_shell_cmd_not_raise.assert_called_once()
        self.assertEqual(2, mock_shell_cmd.call_count)

    @patch("src.modules.pydeps.cmdprocess.shell_cmd")
    @patch("src.modules.pydeps.filelib.write_file")
    @patch("src.modules.ecms_agent.system_agent.SystemAgent.dict_to_list", side_effect=([], [], [], [], []))
    @patch("src.modules.ecms_agent.system_agent.SystemAgent.list_to_dict", side_effect=({}, {}, {}, {}, {}))
    @patch("src.modules.pydeps.filelib.read_file", side_effect=([], [], [], [], []))
    @patch("os.listdir", return_value=["test01.conf", "test02.conf"])
    @patch("src.modules.pydeps.cmdprocess.shell_cmd_not_raise", return_value=(0, "", ""))
    @patch("__builtin__.open", new_callable=mock_open())
    @patch("os.path.isfile", return_value=False)
    def test_update_sysctl_parameters_true(self, mock_isfile, mock_mopen, mock_shell_cmd_not_raise, mock_listdir,
                                           mock_read_file, mock__list_to_dict, mock__dict_to_list, mock_write_file,
                                           mock_shell_cmd):
        SystemAgent.update_sysctl_parameters({"a": 1})
        mock_isfile.assert_called_once()
        mock_mopen.assert_called_once()
        mock_shell_cmd_not_raise.assert_called_once()
        mock_listdir.assert_called_once()
        self.assertEqual(4, mock_read_file.call_count)
        mock__list_to_dict.assert_called()
        mock__dict_to_list.assert_called()
        mock_write_file.assert_called()
        self.assertEqual(2, mock_shell_cmd.call_count)

    @patch("src.modules.pydeps.cmdprocess.shell_cmd")
    @patch("src.modules.pydeps.filelib.write_file")
    @patch("src.modules.ecms_agent.system_agent.SystemAgent.dict_to_list", side_effect=([], [], [], [], []))
    @patch("src.modules.ecms_agent.system_agent.SystemAgent.list_to_dict", side_effect=({}, {}, {}, {}, {}))
    @patch("src.modules.pydeps.filelib.read_file", side_effect=([], [], [], [], []))
    @patch("os.listdir", return_value=["test01.conf", "test02.conf"])
    def test_delete_sysctl_parameters(self, mock_listdir, mock_read_file, mock__list_to_dict, mock__dict_to_list,
                                      mock_write_file, mock_shell_cmd):
        SystemAgent.delete_sysctl_parameters({"a": 1})
        mock_listdir.assert_called_once()
        mock_read_file.assert_called()
        mock__list_to_dict.assert_called()
        mock__dict_to_list.assert_called()
        mock_write_file.assert_called()
        self.assertEqual(2, mock_shell_cmd.call_count)

    def test_list_to_dict_exception(self):
        with self.assertRaises(Exception):
            SystemAgent.list_to_dict("sss")

    def test_list_to_dict_true(self):
        SystemAgent.list_to_dict(['test01', 'test02'])

    def test_dict_to_list(self):
        SystemAgent.dict_to_list({"a": 1})

    @patch("src.modules.ecms_agent.system_agent.SystemAgent.list_to_dict", return_value={"a": 1})
    @patch("src.modules.pydeps.filelib.read_file", side_effect=([], [], [], [], []))
    def test_get_conf(self, mock_read_file, mock_list_to_dict):
        mydict = SystemAgent.get_conf("ss")
        mock_read_file.assert_called()
        mock_list_to_dict.assert_called()
        self.assertEqual(mydict, {"ss": {"a": 1}})

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd_not_raise", side_effect=[(1, "", "")])
    def test_restart_services(self, mock_shell_cmd_not_raise):
        self.assertTrue(SystemAgent.restart_services("test"))
        mock_shell_cmd_not_raise.assert_called_once()

    def test_get_rpscpus(self):
        agent = SystemAgent()
        agent.get_rpscpus(12)

    @patch(target="os.listdir", return_value=(["tx-0", "tx-1"]))
    def test_get_queues_num(self, mock_listdir):
        agent = SystemAgent()
        agent.get_queues_num("test")
        mock_listdir.assert_called_once()

    @patch(target="commands.getstatusoutput", side_effect=[(0, ""), (0, "0-23")])
    @patch(target="src.modules.ecms_agent.system_agent.SystemAgent.get_list", side_effect=["3", "4", "5"])
    def test_get_priority_cpu(self, mock_get_list, mock_getstatusoutput):
        agent = SystemAgent()
        agent.get_priority_cpu("test")
        mock_get_list.assert_called()
        mock_getstatusoutput.assert_called()

    def test_get_list(self):
        agent = SystemAgent()
        agent.get_list(0, 23)

    @patch(target="src.modules.ecms_agent.system_agent.SystemAgent.get_priority_cpu", return_value=[1, 2, 3, 4])
    def test_get_cpu_bind_list(self, mock_get_priority_cpu):
        agent = SystemAgent()
        agent.get_cpu_bind_list(8, "test")
        mock_get_priority_cpu.assert_called_once()

    def test_get_eths(self):
        agent = SystemAgent()
        agent.get_eths()

    @patch(target="commands.getstatusoutput", return_value=(0, "56  57  58  59  60"))
    def test_get_interrupts(self, mock_getstatusoutput):
        agent = SystemAgent()
        agent.get_interrupts("eth")
        mock_getstatusoutput.assert_called_once()

    @patch(target="os.path.exists", side_effect=[True])
    @patch(target="src.modules.ecms_agent.system_agent.SystemAgent.get_cpu_bind_list", return_value=[1])
    @patch(target="src.modules.ecms_agent.system_agent.SystemAgent.get_queues_num", return_value=1)
    @patch(target="src.modules.ecms_agent.system_agent.SystemAgent.get_interrupts", return_value=[1])
    @patch(target="src.modules.ecms_agent.system_agent.SystemAgent.get_eths", return_value=["test"])
    @patch(target="src.modules.ecms_agent.system_agent.SystemAgent.get_rpscpus", return_value="f")
    @patch(target="commands.getstatusoutput", side_effect=[(0, ""), (0, ""), (0, ""), (0, ""), (0, "")])
    def test_bind_core(self, mock_getstatusoutput, mock_get_rpscpus, mock_get_eths, mock_get_interrupts,
                       mock_get_queues_num, mock_get_cpu_bind_list, mock_exists):
        SystemAgent.bind_core()
        mock_getstatusoutput.assert_called()
        mock_get_rpscpus.assert_called()
        mock_get_eths.assert_called()
        mock_get_interrupts.assert_called()
        mock_get_queues_num.assert_called()
        mock_get_cpu_bind_list.assert_called()
        mock_exists.assert_called()
