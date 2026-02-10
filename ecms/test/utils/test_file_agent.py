#!/usr/bin/env python
# -*- coding: utf-8 -*-
# @Time    : 2021/4/26 16:30
# @Author  : mo.kang<mo.kang@eisoo.com>
# @Site    : 
# @File    : test_file_agent.py
# @Software: PyCharm
import os
import sys
import unittest

from mock import patch
from mock import mock_open

CURR_SCRIPT_PATH = os.path.dirname(os.path.realpath(__file__))
SRC_PATH = os.path.dirname(os.path.dirname(CURR_SCRIPT_PATH))
sys.path.append(SRC_PATH)

from src.modules.ecms_agent.file_agent import FileAgent


class TestFileAgent(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        pass


class FileTest(TestFileAgent):
    @patch(target="os.path.exists", return_value=True)
    def test_directory_exists(self, mock_exists):
        self.assertTrue(FileAgent.directory_exists("test"))
        mock_exists.assert_called()

    @patch(target="os.path.exists", return_value=False)
    @patch(target="os.makedirs")
    @patch(target="os.chmod")
    def test_create_directory(self, mock_chmod, mock_makedirs, mock_exists):
        FileAgent.create_directory("/test")
        mock_chmod.assert_called()
        mock_makedirs.assert_called()
        mock_exists.assert_called()

    @patch(target="os.path.exists", return_value=True)
    @patch(target="shutil.rmtree")
    def test_delete_directory(self, mock_rmtree, mock_exists):
        FileAgent.delete_directory("test")
        mock_rmtree.assert_called()
        mock_exists.assert_called()

    @patch(target="os.path.isfile", return_value=True)
    @patch(target="os.chmod")
    def test_creat_file(self, mock_chmod, mock_isfile):
        FileAgent.creat_file("/test")
        mock_chmod.assert_called()
        mock_isfile.assert_called()

    @patch(target="os.path.isfile", return_value=False)
    @patch(target="__builtin__.open", new_callable=mock_open())
    @patch(target="os.chmod")
    def test_creat_file(self, mock_chmod, mock_mopen, mock_isfile):
        FileAgent.creat_file("/test")
        mock_chmod.assert_called()
        mock_mopen.assert_called()
        mock_isfile.assert_called()

    @patch(target="os.path.isfile", return_value=True)
    def test_file_exists(self, mock_isfile):
        self.assertTrue(FileAgent.file_exists("test"))
        mock_isfile.assert_called()

    @patch(target="os.path.isfile", return_value=True)
    @patch(target="os.remove")
    def test_delete_file_exist(self, mock_remove, mock_isfile):
        FileAgent.delete_file("test")
        mock_remove.assert_called()
        mock_isfile.assert_called()

    @patch(target="os.path.isfile", return_value=False)
    def test_delete_file_not_exist(self, mock_isfile):
        FileAgent.delete_file("test")
        mock_isfile.assert_called()

    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd_not_raise", side_effect=[(1, "", ""), (0, "", "")])
    @patch(target="src.modules.pydeps.cmdprocess.shell_cmd")
    @patch(target="src.modules.ecms_agent.file_agent.subprocess.call")
    @patch(target="src.modules.ecms_agent.file_agent.subprocess.Popen")
    @patch(target="__builtin__.open", new_callable=mock_open())
    @patch(target="os.remove")
    @patch(target="os.rename")
    @patch(target="src.modules.ecms_agent.file_agent.FileAgent.file_exists", side_effect=[False, True, True, True, True,
                                                                                          True, True, True, True])
    def test_update_tls(self, mock_file_exists, mock_rename, mock_remove, mock_mopen, mock_popen, mock_call,
                        mock_shell_cmd, mock_shell_cmd_not_raise):
        process = mock_popen.return_value
        process.communicate.return_value = (b"", b"")
        process.returncode = 0
        FileAgent.update_tls("test")
        mock_file_exists.assert_called()
        test = """PRIVATE
-----BEGIN CERTIFICATE-----
Y
-----END CERTIFICATE-----
"""
        FileAgent.update_tls(test)
        mock_rename.assert_called()
        mock_remove.assert_called()
        mock_mopen.assert_called()
        mock_shell_cmd.assert_called()
        mock_shell_cmd_not_raise.assert_called()
        FileAgent.update_tls(test)
