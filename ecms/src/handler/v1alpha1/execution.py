import json
import time
import logging
import subprocess
import os
import distutils.util

from tornado.concurrent import run_on_executor

from src.handler.ecms_handler import BaseHandler
from src.handler.ecms_handler import handler_try_except
from src.handler.ecms_handler import simple_auth

logger = logging.getLogger(__name__)

DEVNULL = os.open(os.devnull, os.O_RDWR)


class ExecutionHandler(BaseHandler):

    @run_on_executor
    @handler_try_except
    @simple_auth
    def post(self):
        command = self.get_query_arguments("command")

        # parse query parameters
        query_stdin = distutils.util.strtobool(
            str(self.get_query_argument("stdin", "false")))
        query_stdout = distutils.util.strtobool(
            str(self.get_query_argument("stdout", "true")))
        query_stderr = distutils.util.strtobool(
            str(self.get_query_argument("stderr", "false")))

        stdin = b""
        if query_stdin:
            stdin = self.request.body

        stdout = DEVNULL
        if query_stdout:
            stdout = subprocess.PIPE

        stderr = subprocess.PIPE
        if query_stdout and query_stderr:
            stderr = subprocess.STDOUT

        # logging
        logger.debug("execute command: %s, stdin: %s, stdout: %s, stderr: %s",
                     " ".join(command), query_stdin, query_stdout,
                     query_stderr)

        try:
            p = subprocess.Popen(command,
                                 stdin=subprocess.PIPE,
                                 stdout=stdout,
                                 stderr=stderr)
        except OSError as ex:
            if ex.errno is not 2:
                raise ex
            msg = {
                "status": 404000000,
                "message": ex.strerror,
            }
            self.set_status(404)
            json.dump(msg, self)
            return

        bytes_stdout, bytes_stderr = p.communicate(input=stdin)

        self.set_header("x-exit-code", str(p.returncode))

        if p.returncode is not 0:
            if query_stdout and query_stderr:
                logger.warning(
                    "execute command: %s, return code: %d, combined output:\n%s",
                    " ".join(command), p.returncode, bytes_stdout)
            else:
                logger.warning(
                    "execute command: %s, return code: %d, stderr:\n%s",
                    " ".join(command), p.returncode, bytes_stderr)

        if query_stdout:
            self.write(bytes_stdout)
            return
        if query_stderr:
            self.write(bytes_stderr)
            return
