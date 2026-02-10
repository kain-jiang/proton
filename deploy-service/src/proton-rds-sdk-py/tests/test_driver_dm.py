import re
import allure
import pytest
import os

os.environ['DB_TYPE']='DM8'
import rdsdriver

@allure.feature("proton-rds-sdk-py/driver")
class Test_DM:
    @allure.title("init")
    def test_0(self,host,port,user,password,database1):
        pass

    @allure.title("attr&object")
    def test_1(self,host,port,user,password,database1):
        # GLOBAL

        conn = rdsdriver.connect(host=host, port=int(port), user=user, password=password, database=database1, autocommit=True)

        assert hasattr(rdsdriver, 'apilevel')
        assert hasattr(rdsdriver, 'threadsafety')
        assert hasattr(rdsdriver, 'paramstyle')

        # EXCEPTION
        assert hasattr(rdsdriver, 'Warning')
        assert hasattr(rdsdriver, 'Error')
        assert hasattr(rdsdriver, 'InterfaceError')
        assert hasattr(rdsdriver, 'DatabaseError')
        assert hasattr(rdsdriver, 'DataError')
        assert hasattr(rdsdriver, 'OperationalError')
        assert hasattr(rdsdriver, 'IntegrityError')
        assert hasattr(rdsdriver, 'InternalError')
        assert hasattr(rdsdriver, 'ProgrammingError')
        assert hasattr(rdsdriver, 'NotSupportedError')

        # Type Objects and Constructors
        rdsdriver.Date(1,2,3)
        rdsdriver.Time(1,2,3)
        rdsdriver.Timestamp(1,2,3,4,5,6)
        rdsdriver.DateFromTicks(1)
        rdsdriver.TimeFromTicks(1)
        rdsdriver.TimestampFromTicks(1)
        assert hasattr(rdsdriver, 'STRING')
        assert hasattr(rdsdriver, 'BINARY')
        assert hasattr(rdsdriver, 'NUMBER')
        assert hasattr(rdsdriver, 'DATETIME')
        assert hasattr(rdsdriver, 'ROWID')

        # Connection
        assert hasattr(conn, 'close')
        assert hasattr(conn, 'commit')
        assert hasattr(conn, 'rollback')
        assert hasattr(conn, 'cursor')

        # Cursor
        cursor = conn.cursor()
        assert hasattr(cursor, 'description')
        assert hasattr(cursor, 'rowcount')
        assert hasattr(cursor, 'arraysize')
        assert hasattr(cursor, 'callproc')
        assert hasattr(cursor, 'close')
        assert hasattr(cursor, 'execute')
        assert hasattr(cursor, 'executemany')
        assert hasattr(cursor, 'fetchone')
        assert hasattr(cursor, 'fetchmany')
        assert hasattr(cursor, 'fetchall')
        assert hasattr(cursor, 'setinputsizes')
        assert hasattr(cursor, 'setoutputsize')
        cursor.close()
        conn.close()

    @allure.title("connection.commit, cursor, close")
    def test_2(self,host,port,user,password,database1):
        # exec sql not commit
        conn = rdsdriver.connect(host=host, port=int(port), user=user, password=password, database=database1, autocommit=False)
        cursor = conn.cursor()
        cursor.execute("drop table if exists t1")
        cursor.execute("create table if not exists t1(id int)")
        cursor.execute("insert into t1 values(%s)", (1,))
        cursor.execute("select * from t1")
        r = cursor.fetchall()
        assert len(r) == 1

        # another session can't see the record
        conn2 = rdsdriver.connect(host=host, port=int(port), user=user, password=password, database=database1, autocommit=True)
        cursor2 = conn2.cursor()
        cursor2.execute("select * from t1")
        r = cursor2.fetchall()
        assert len(r) == 0

        # commit and can see the record
        conn.commit()
        cursor2.execute("select * from t1")
        r = cursor2.fetchall()
        assert len(r) == 1
        cursor2.execute("delete from t1")
        cursor2.execute("select * from t1")
        r = cursor2.fetchall()
        assert len(r) == 0

        cursor2.close()
        cursor.close()
        conn.close()
        conn2.close()

    @allure.title("cursor.execute, executemany, fetchone,fetchall,fetchmany")
    def test_3(self,host,port,user,password,database1):
        conn = rdsdriver.connect(host=host, port=int(port), user=user, password=password, database=database1, autocommit=True)
        cursor = conn.cursor()
        cursor.execute("drop table if exists t1")
        cursor.execute("create table if not exists t1(id int)")
        cursor.execute("insert into t1 values(%s)", (1,))
        cursor.executemany("insert into t1 values(%s)", ((2,),(3,),(4,),(5,)))

        assert cursor.rowcount == 4
        cursor.execute("select * from t1")
        r = cursor.fetchall()
        assert len(r) == 5

        cursor.execute("select * from t1")
        r = cursor.fetchone()
        assert len(r) == 1

        r = cursor.fetchmany(2)
        assert len(r) == 2

        cursor.close()
        conn.close()





