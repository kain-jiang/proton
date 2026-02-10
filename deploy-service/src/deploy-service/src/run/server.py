import logging
import os
import signal

import psutil
from tornado import httpserver, ioloop, log

from src.common.log_util import logger
from src.urls import application

_TORNADO_ACCESS_LOG_LEVEL = log.access_log.getEffectiveLevel()
_TORNADO_APP_LOG_LEVEL = log.app_log.getEffectiveLevel()
_TORNADO_GEN_LOG_LEVEL = log.gen_log.getEffectiveLevel()


def _set_child_process_tornado_log_level(*_):
    # 向所有子进程发送信号
    for p in psutil.Process(os.getpid()).children():
        os.kill(p.pid, signal.SIGUSR2)


def _get_tornado_log_duration_time() -> int:
    duration_file = "/tmp/tornado_log_duration"
    duration = 10800
    if os.path.exists(duration_file) and os.path.isfile(duration_file):
        with open(duration_file, "r", encoding="utf-8") as fp:
            try:
                duration = int(fp.read().strip())
            except ValueError as exp:
                logger.error("get tornado log duration time failed, err: %s", str(exp))
                logger.warning("use default duration time: %s seconds", duration)
    return duration


def _get_tornado_log_level():
    log_level_file = "/tmp/tornado_log_level"
    log_level = logging.DEBUG  # DEBUG
    if os.path.exists(log_level_file) and os.path.isfile(log_level_file):
        with open(log_level_file, "r", encoding="utf-8") as fp:
            try:
                log_level = int(fp.read().strip())
            except ValueError as exp:
                logger.error("get tornado log level failed, err: %s", str(exp))
                logger.warning("use default log level: %s", log_level)
    return log_level


def _set_tornado_log_level(*_):
    log_level = _get_tornado_log_level()

    pid = os.getpid()

    log.access_log.setLevel(log_level)
    logger.info("process[%s] set tornado access log level: %s", pid, log_level)

    log.app_log.setLevel(log_level)
    logger.info("process[%s] set tornado app log level: %s", pid, log_level)

    log.gen_log.setLevel(log_level)
    logger.info("process[%s] set tornado gen log level: %s", pid, log_level)

    duration = _get_tornado_log_duration_time()
    signal.alarm(duration)
    logger.info("process[%s] restore tornado log level after %s seconds", pid, duration)


def _restore_tornado_log_level(*_):
    pid = os.getpid()

    log.access_log.setLevel(_TORNADO_ACCESS_LOG_LEVEL)
    logger.info(
        "process[%s] restored tornado access log level: %s",
        pid,
        _TORNADO_ACCESS_LOG_LEVEL,
    )

    log.app_log.setLevel(_TORNADO_APP_LOG_LEVEL)
    logger.info("process[%s] restored tornado app log level: %s", pid, _TORNADO_ACCESS_LOG_LEVEL)

    log.gen_log.setLevel(_TORNADO_GEN_LOG_LEVEL)
    logger.info("process[%s] restored tornado gen log level: %s", pid, _TORNADO_ACCESS_LOG_LEVEL)


def registry_signal():
    signal.signal(signal.SIGUSR1, _set_child_process_tornado_log_level)
    signal.signal(signal.SIGUSR2, _set_tornado_log_level)
    signal.signal(signal.SIGALRM, _restore_tornado_log_level)


def start_deploy_service():
    registry_signal()
    server = httpserver.HTTPServer(application)
    server.listen(9703)
    server.start(int(os.environ.get("TORNADO_PROCESS", "10")))
    ioloop.IOLoop.current().start()
