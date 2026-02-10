#!/usr/bin/env python
# -*- encoding: utf-8 -*-
"""
@Description:
@Date: 2020/12/24
@Author: ruan.yulin
"""

from mock import MagicMock


def mock_requests_response(*args, **kwargs):
    response = MagicMock()
    response.status_code = kwargs.get("status_code", 200)
    response.text = kwargs.get("text", "ut_text")
    response.json.return_value = kwargs.get("json", {"ut_key": "ut_value"})
    return response


def mock_ssh_exec_command_return(*args, **kwargs):
    stdout = MagicMock()
    stdout.read.return_value = kwargs.get("stdout", "ut_stdout").encode()
    stdout.channel.recv_exit_status.return_value = kwargs.get("exit_status", 0)
    stderr = MagicMock()
    stderr.read.return_value = kwargs.get("stderr", "ut_stderr").encode()
    stderr.channel.recv_exit_status.return_value = kwargs.get("exit_status", 0)

    return None, stdout, stderr
