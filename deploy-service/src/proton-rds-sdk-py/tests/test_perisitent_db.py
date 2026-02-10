import re
import allure
import pytest
import pymysql

from dbutilsx.persistent_db import PersistentDB, PersistentDBInfo

op = None

@allure.feature("proton-rds-sdk-py/dbutilsx")
class Test_PersistentDB:
    @allure.title("init")
    def test_0(self,host,port,user,password,database1):
        global op
        w = PersistentDBInfo(creator=pymysql, host=host, port=int(port), user=user, password=password, database=database1)
        op = PersistentDB(
            master = w,
            backup = w,
        )

    @allure.title("queryAndFetchOne")
    def test_queryAndFetchOne(self):
        op.execute("create table if not exists t1(id int)")
        op.execute("insert into t1 values(%s)", (1,))
        r = op.queryAndFetchOne("select * from t1")
        assert r == (1, )
        op.execute("delete from t1")

    @allure.title("queryAndFetchMany")
    def test_queryAndFetchMany(self):
        op.execute("create table if not exists t1(id int)")
        op.executemany("insert into t1 values(%s);", ((1, ), (2, ), (3, )))
        r = op.queryAndFetchMany("select * from t1", size=1)
        assert r == ((1, ),)
        r = op.queryAndFetchMany("select * from t1", size=5)
        assert len(r) == 3
        op.execute("delete from t1")

    @allure.title("queryAndFetchAll")
    def test_queryAndFetchAll(self):
        op.execute("create table if not exists t1(id int)")
        op.executemany("insert into t1 values(%s);", ((1, ), (2, ), (3, )))
        r = op.queryAndFetchAll("select * from t1")
        assert len(r) == 3
        op.execute("delete from t1")

    @allure.title("connection")
    def test_connection(self):
        with op.connection() as conn:
            with conn.cursor() as cur:
                cur.execute("create table if not exists t1(id int)")
                cur.execute("insert into t1 values(%s)", (1,))
                cur.execute("select * from t1")
                r = cur.fetchone()
                assert r == (1, )
                op.execute("delete from t1")

    @allure.title("read write split")
    def test_readWriteSplit(self,host,port,user,password,database1,database2):
        w = PersistentDBInfo(creator=pymysql, host=host, port=int(port), user=user, password=password, database=database1)
        r = PersistentDBInfo(creator=pymysql, host=host, port=int(port), user=user, password=password, database=database2)
        op = PersistentDB(
            master = w,
            backup = r,
        )
        op.execute("create table if not exists t1(id int)")
        op.executemany("insert into t1 values(%s);", ((1, ), (2, ), (3, )))
        try:
            r = op.queryAndFetchAll("select * from t1")
        except Exception as e:
            assert True
        else:
            assert False
        op.execute("delete from t1")


