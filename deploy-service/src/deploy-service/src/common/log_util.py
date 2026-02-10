#!/usr/bin/env python
# coding=utf-8
import logging.handlers
import os
import re
import sys
import logging
import requests

class FilterFormatter(logging.Formatter):
    def format(self, record) -> str:
        return re.sub(
            r'("|\')(password|sentinelPassword)("|\'):([ ]*)("|\')([^"\']*)("|\')',
            r"\1\2\3:\4\5*****\7",
            logging.Formatter.format(self, record),
            flags=re.I,
        )


_nameToLevel = {
    "CRITICAL": logging.CRITICAL,
    "FATAL": logging.FATAL,
    "ERROR": logging.ERROR,
    "WARN": logging.WARNING,
    "WARNING": logging.WARNING,
    "INFO": logging.INFO,
    "DEBUG": logging.DEBUG,
    "NOTSET": logging.NOTSET,
}
log_level = _nameToLevel.get(os.environ.get("LOG_LEVEL", "INFO").upper())
basic_log_level = _nameToLevel.get(os.environ.get("BASIC_LOG_LEVEL", os.environ.get("LOG_LEVEL", "INFO")).upper())

logging.basicConfig(level=basic_log_level)

logger = logging.Logger("DeployService")
# 必须设置，这里如果不显示设置，默认过滤掉warning之前的所有级别的信息
logger.setLevel(log_level)

# stdout日志输出格式
stdout_formatter = FilterFormatter(
    "[%(asctime)s] %(process)d %(filename)s %(funcName)s " "line:%(lineno)d [%(levelname)s] %(message)s"
)

# 创建一个FileHandler， 向文件输出日志信息
# 创建一个StreamHandler， 向stdout输出日志信息
stdout_handler = logging.StreamHandler(sys.stdout)

# 设置日志等级
stdout_handler.setLevel(log_level)

# 设置handler的格式对象
stdout_handler.setFormatter(stdout_formatter)
# 将handler增加到logger中
logger.addHandler(stdout_handler)


def log_response(response):
    log_msg = "Request URL: {url}, method: {method}, code: {status_code}, content: {response_content}".format(
        url=response.request.url,
        method=response.request.method,
        status_code=response.status_code,
        response_content=response.text,
    )
    logger.debug(log_msg)


def log_response_info(response):
    log_msg = "Request URL: {url}, method: {method}, code: {status_code}, content: {response_content}".format(
        url=response.request.url,
        method=response.request.method,
        status_code=response.status_code,
        response_content=response.text,
    )
    logger.info(log_msg)


def log_response_with_request(response: requests.Response):
    logger.debug(
        f"Request URL: {response.request.url}, Request Body: {response.request.body}, "
        f"method: {response.request.method}, code: {response.status_code}, content: {response.text}"
    )
