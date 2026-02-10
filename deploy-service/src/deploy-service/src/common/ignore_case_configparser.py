#!/usr/bin/env python
# coding=utf-8
from configparser import ConfigParser
from typing import Optional


class IgnoreCaseConfigParser(ConfigParser):
    def __init__(self, defaults=None):
        ConfigParser.__init__(self, defaults=defaults, allow_no_value=True)

    def optionxform(self, optionstr):
        return optionstr

    def set(self, section: str, option: str, value: Optional[str] = ...) -> None:
        if value is not None:
            value = value.replace("%", "%%")
        super().set(section, option, value)
