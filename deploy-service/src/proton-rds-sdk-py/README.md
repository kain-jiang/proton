# Introduction
Provides a set of extensions on library https://webwareforpython.github.io/DBUtils

# Installation
```bash
$ #centos-release-scl-rh repo: http://mirrors.tuna.tsinghua.edu.cn/centos/7.9.2009/sclo/x86_64/rh/
$ yum install postgresql12-devel python3-devel --nogpgcheck # apt install libpq-dev python3-dev
$ arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/)
$ curl ftp://ftp-ict.aishu.cn/proton/dm/${arch}/dpi_8.1.2.tar.gz -o dpi.tar.gz
$ tar -zxvf dpi.tar.gz -C /usr/lib #set LD_LIBRARY_PATH
$ python3 setup.py install
```
# Usage
```text
Using the Python interpreter console, you can display the documentation of the pooled_db module as follows (this works analogously for the other modules):

$ python3
Python 3.8.13 (default, Apr 26 2022, 16:57:08)
[GCC 4.8.5 20150623 (Red Hat 4.8.5-44)] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> import dbutilsx.pooled_db
>>> help(dbutilsx.pooled_db)
```
```text
The Class PersistentDB in the module dbutilsx.persistent_db implements steady, thread-affine, persistent connections to a database, using any DB-API 2 database module. "Thread-affine" and "persistent" means that the individual database connections stay assigned to the respective threads and will not be closed during the lifetime of the threads.

The class PooledDB in the module dbutilsx.pooled_db implements a pool of steady, thread-safe cached connections to a database which are transparently reused, using any DB-API 2 database module.

PersistentDB will make more sense if your application keeps a constant number of threads which frequently use the database. In this case, you will always have the same amount of open database connections. However, if your application frequently starts and ends threads, then it will be better to use PooledDB.

rdsdriver is used to access databases with Python Database API Specification v2.0
```
# Example
Examples are available in example directory.
```python

import pymysql
from dbutilsx.pooled_db import PooledDB, PooledDBInfo

if __name__ == '__main__':
    # master node
    w = PooledDBInfo(
        creator = pymysql,
        host = '192.168.166.239',
        port = 3306,
        user = 'root',
        password = 'fake_password',
        database = "testdb",
        autocommit=True,
    )

    # backup node
    r = PooledDBInfo(
        creator = pymysql,
        host = '192.168.166.239',
        port = 3307,
        user = 'root',
        password = 'fake_password',
        database = "testdb",
        autocommit=True,
    )
    op = PooledDB(
        master = w,
        backup = r,
    )
    op.execute("create table if not exists t1(id int)")
    op.executemany("replace into t1 values(%s);", ((1, ), (2, ), (3, )))
    print(op.queryAndFetchAll("select * from t1"))
```