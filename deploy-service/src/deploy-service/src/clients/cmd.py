import subprocess
from threading import Timer
from typing import Union, Tuple
from src.common.log_util import logger

import tenacity


class CmdClient(object):
    @classmethod
    def run(
        cls, cmd_str: Union[str, list], timeout: int = 600, retry_attempt: int = 3, log_debug: bool = False
    ) -> Tuple[int, str, str]:
        """运行命令，如果返回值非0，不会抛出异常

        :param cmd_str: 命令字符串或者列表
        :param timeout: 超时时间，默认600s
        :param retry_attempt: 重试次数默认3，间隔为3
        :return: (code, message, error) 返回码，标准输出和标准错误
        """

        @tenacity.retry(
            stop=tenacity.stop_after_attempt(retry_attempt),
            wait=tenacity.wait_fixed(3),
            retry=tenacity.retry_if_result(lambda rel: rel[0] != 0),
            reraise=True,
        )
        def __run(cmd_str: Union[str, list], timeout: int = 600) -> Tuple[int, str, str]:
            try:
                proc = subprocess.Popen(
                    args=cmd_str,
                    shell=isinstance(cmd_str, str),
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                    close_fds=True,
                )
            except Exception:
                raise
            tr = Timer(timeout, proc.kill)
            tr.start()
            try:
                msg, err = proc.communicate(input=None)
                if tr.is_alive():
                    if log_debug:
                        logger.debug(f"CmdClient execute return <{proc.returncode}> for command: <{cmd_str}>")
                    else:
                        logger.info(f"CmdClient execute return <{proc.returncode}> for command: <{cmd_str}>")
                    return proc.returncode, msg.decode(), err.decode()
                cmd = cmd_str if isinstance(cmd_str, str) else " ".join(cmd_str)
                raise TimeoutError(f"command: {{{cmd}}} timeout: {timeout}")
            finally:
                tr.cancel()
                if proc:
                    proc.stdin and proc.stdin.close()
                    proc.stdout and proc.stdout.close()
                    proc.stderr and proc.stderr.close()

        try:
            return __run(cmd_str=cmd_str, timeout=timeout)
        except tenacity.RetryError as re:
            return re.last_attempt.result()

    @classmethod
    def run_or_raise(
        cls, cmd_str: Union[str, list], timeout: int = 600, retry_attempt: int = 3, log_debug: bool = False
    ) -> Tuple[str, str]:
        """运行命令，如果返回值为非0时，抛出异常

        :param cmd_str: 命令字符串或者列表
        :param timeout: 超时时间，默认600s
        :param retry_attempt: 重试次数默认3，间隔为3
        :return: (message, error) 标准输出和标准错误
        """
        code, msg, err = cls.run(cmd_str, timeout, retry_attempt, log_debug=log_debug)
        if code != 0:
            cmd = cmd_str if isinstance(cmd_str, str) else " ".join(cmd_str)
            raise Exception(f"command: {{{cmd}}}, code: {code} msg: {{{msg}}}, err: {{{err}}}")
        return msg, err
