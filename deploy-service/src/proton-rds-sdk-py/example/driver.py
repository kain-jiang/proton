import os
os.environ['DB_TYPE']='KDB9'
import rdsdriver

conn = rdsdriver.connect(host='127.0.0.1',port=4321, user='example_user', password=os.environ.get('DB_PASSWORD', ''),database='test2',cursorclass=rdsdriver.TupleCursor,autocommit=0)
cursor = conn.cursor()
#cursor.execute("drop table t1")
#cursor.execute("create table t1(id int)")
#cursor.execute("insert into t1 values(%s)", (1,))
cursor.execute("insert into t1 values(%s)", (2,))
#cursor.execute("select * from t1 where id in (?,?,?)",(1,2,3))

#cursor.execute("delete from t1")
cursor.execute("select id from t1")
print(cursor.fetchall())
cursor.close()
conn.close()




