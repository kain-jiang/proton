# -*- coding: utf-8 -*-

import json
import os
import logging
import distutils.util
import shutil
import mimetypes

from tornado.concurrent import run_on_executor

from src.handler.ecms_handler import BaseHandler
from src.handler.ecms_handler import handler_try_except
from src.handler.ecms_handler import simple_auth

logger = logging.getLogger(__name__)

S_IFMT = 0o170000  # bit mask for the file type bit field

S_IFSOCK = 0o140000  # socket
S_IFLNK = 0o120000  # symbolic link
S_IFREG = 0o100000  # regular file
S_IFBLK = 0o060000  # block device
S_IFDIR = 0o040000  # directory
S_IFCHR = 0o020000  # character device
S_IFIFO = 0o010000  # FIFO


class UnimplementedError(Exception):
    pass


class FileHandler(BaseHandler):

    @run_on_executor
    @handler_try_except
    @simple_auth
    def post(self, path):
        content_type = self.request.headers.get("content-type")

        # create directory
        if content_type == "application/json":
            content = json.loads(self.request.body)
            # validate request
            if not isinstance(content, dict):
                msg = {
                    "status": 400000000,
                    "message": "invalid request",
                }
                self.set_status(400)
                json.dump(msg, self)
                return
            if "type" not in content:
                msg = {
                    "status": 400000000,
                    "message": "required field type is missing",
                }
                self.set_status(400)
                json.dump(msg, self)
                return
            if content["type"] != "directory":
                msg = {
                    "status": 400000000,
                    "message": "field type should be directory",
                }
                self.set_status(400)
                json.dump(msg, self)
                return

            if os.path.isdir(path):
                return

            logger.info("create directory: %s", path)
            os.makedirs(path)
            return

        # check parent
        parent = os.path.dirname(path)
        if parent != "":
            if not os.path.exists(parent):
                msg = {
                    "status": 400000000,
                    "message": "parent is not found",
                    "detail": {
                        "path": parent,
                    }
                }
                self.set_status(400)
                json.dump(msg, self)
                return
            if not os.path.isdir(parent):
                msg = {
                    "status": 400000000,
                    "message": "parent is not a directory",
                    "detail": {
                        "path": parent,
                    }
                }
                self.set_status(400)
                json.dump(msg, self)
                return
            if not os.access(parent, os.W_OK):
                msg = {
                    "status": 400000000,
                    "message": "parent is not writable",
                    "detail": {
                        "path": parent,
                    }
                }
                self.set_status(400)
                json.dump(msg, self)
                return

        # create file
        logger.info("create file: %s", path)
        with open(path, mode="wb") as f:
            f.write(self.request.body)

    @run_on_executor
    @handler_try_except
    @simple_auth
    def delete(self, path):
        # use os.lstat rather than os.stat to support removing symbolic link
        try:
            s = os.lstat(path)
        except OSError as ex:
            if ex.errno is not 2:
                raise ex
            msg = {
                "status": 404000000,
                "message": "not found",
                "detail": {
                    "path": path,
                }
            }
            self.set_status(404)
            json.dump(msg, self)
            return

        # check parent is writable
        parent = os.path.dirname(path)
        if parent == "":
            parent = "."
        if not os.access(parent, os.W_OK):
            msg = {
                "status": 400000000,
                "message": "parent is not writable",
                "detail": {
                    "parent": parent,
                }
            }
            self.set_status(400)
            json.dump(msg, self)
            return

        # check target is symbolic link, regular file or directory
        if s.st_mode & S_IFMT not in (S_IFLNK, S_IFREG, S_IFDIR):
            msg = {
                "status": 400000000,
                "message": "unsupported file type",
                "detail": {
                    "mode": oct(s.st_mode),
                }
            }
            self.set_status(400)
            json.dump(msg, self)
            return

        if s.st_mode & S_IFMT == S_IFLNK:
            logger.info("remove symbolic link: %s", path)
            os.remove(path)
            return
        if s.st_mode & S_IFMT == S_IFREG:
            logger.info("remove regular file: %s", path)
            os.remove(path)
            return
        if s.st_mode & S_IFMT == S_IFDIR:
            logger.info("remove directory: %s", path)
            shutil.rmtree(path)
            return

    @run_on_executor
    @handler_try_except
    @simple_auth
    def put(self, path):
        try:
            s = os.stat(path)
        except OSError as ex:
            if ex.errno is not 2:
                raise ex
            msg = {
                "status": 404000000,
                "message": "not found",
                "detail": {
                    "path": path,
                }
            }
            self.set_status(404)
            json.dump(msg, self)
            return

        if not os.access(path, os.W_OK):
            msg = {
                "status": 400000000,
                "message": "target is not writable",
                "detail": {
                    "path": path,
                }
            }
            self.set_status(400)
            json.dump(msg, self)
            return

        with open(path, "wb") as fp:
            fp.write(self.request.body)

    @run_on_executor
    @handler_try_except
    @simple_auth
    def head(self, path):
        """获取文件、目录的元数据"""
        # 检查目标是否存在
        if not os.path.exists(path):
            msg = {
                "status": 404000000,
                "message": "not found",
                "detail": {
                    "path": path,
                }
            }
            self.set_status(404)
            json.dump(msg, self)
            return

        # 如果目标是文件
        if not os.path.isdir:
            mime_type, encoding = mimetypes.guess_type(path)
            if mime_type is None:
                mime_type = "application/octet-stream"
            self.set_header(name="content-type", value=mime_type)
            if encoding is not None:
                self.set_header(name="content-encoding", value="encoding")

        follow = self.get_query_argument("follow", default="true")
        stat = os.stat if distutils.util.strtobool(str(follow)) else os.lstat
        self._set_header_x_st(stat(path))

    @run_on_executor
    @handler_try_except
    @simple_auth
    def get(self, path):
        # 检查目标是否存在
        if not os.path.exists(path):
            msg = {
                "status": 404000000,
                "message": "not found",
                "detail": {
                    "path": path,
                }
            }
            self.set_status(404)
            json.dump(msg, self)
            return

        # 获取目标元数据
        self._set_header_x_st(os.stat(path))

        # 目标是目录，返回目录内容的元数据列表
        if os.path.isdir(path):
            results = []
            for n in os.listdir(path):
                r = os.stat(os.path.join(path, n))
                results.append({
                    "name": n,
                    "mode": oct(r.st_mode),
                    "uid": r.st_uid,
                    "gid": r.st_gid,
                    "size": r.st_size,
                })
            json.dump(results, self)
            return

        # 目标是文件
        mime_type, encoding = mimetypes.guess_type(path)
        if mime_type is None:
            mime_type = "application/octet-stream"
        self.set_header(name="content-type", value=mime_type)
        if encoding is not None:
            self.set_header(name="content-encoding", value="encoding")

        # 复制文件内容到 response body
        shutil.copyfileobj(open(path), self)

    def _set_header_x_st(self, r):
        """使用 stat_result 设置 response header"""
        self.set_header("x-st-mode", oct(r.st_mode))
        self.set_header("x-st-uid", r.st_uid)
        self.set_header("x-st-gid", r.st_gid)
        self.set_header("x-st-SIZE", r.st_size)


class FileMovementHandler(BaseHandler):

    @run_on_executor
    @handler_try_except
    def post(self, path):
        # check path parameter
        try:
            os.lstat(path)
        except OSError as ex:
            if ex.errno is not 2:
                raise ex
            msg = {
                "status": 404000000,
                "message": "not found",
                "detail": {
                    "path": path,
                }
            }
            self.set_status(404)
            json.dump(msg, self)
            return
        src = path
        # check body parameter
        content = json.loads(self.request.body)
        if "destination" not in content:
            msg = {
                "status": 400000000,
                "message": "invalid request",
                "detail": content
            }
            self.set_status(404)
            json.dump(msg, self)
            return
        dst = content["destination"]
        logger.info("move %s to %s", src, dst)
        os.rename(src, dst)
