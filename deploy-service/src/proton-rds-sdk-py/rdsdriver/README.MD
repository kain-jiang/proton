# Introduction
rdsdriver is used to access databases with Python Database API Specification v2.0.
rdsdriver selects the corresponding  driver to connect to the database based on different environmental variables.
Pymysql is default driver, When set DB_TYPE=DM8, it will use dmPython.
# Installation
```bash
$ arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/)
$ curl ftp://ftp-ict.aishu.cn/proton/dm/${arch}/dpi_8.1.2.tar.gz -o dpi.tar.gz
$ tar -zxvf dpi.tar.gz -C /usr/lib #set LD_LIBRARY_PATH
$ python3 setup.py install
```
# Usage
```
rdsdriver.connect():
    """
    Representation of a socket with a mysql server.
    :param host: Host where the database server is located.
    :param port: Database port to use
    :param user: Username to log in as.
    :param password: Password to use.
    :param database: Database to use, None to not use a particular one.
    :param autocommit: Autocommit mode.
    :param cursorclass: Custom cursor class to use.[rdsdriver.DictCursor/rdsdriver.TupleCursor]
    :
    """


```
# Example
Examples are available in example directory.
```python

import os
os.environ['DB_TYPE']='DM8'
import rdsdriver

conn = rdsdriver.connect(host='127.0.0.1',port=5236, user='SYSDBA', password='SYSDBA',database='test',cursorclass=rdsdriver.DictCursor)
cursor = conn.cursor()
cursor.execute("drop table t1")
cursor.execute("create table t1(id int)")
cursor.execute("insert into t1 values(%s)", (1,))
cursor.execute("select * from t1 where `id`=%s",(1,))
print(cursor.fetchall())
cursor.close()
conn.close()
```