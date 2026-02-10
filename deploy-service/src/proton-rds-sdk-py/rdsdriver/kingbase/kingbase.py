import psycopg2

import psycopg2.extras

TupleCursor = psycopg2.extensions.cursor
DictCursor = psycopg2.extras.RealDictCursor

rename_kw = {
    'cursorclass': 'cursor_factory'
}


avaliable_kw = {
    'user',
    'password',
    'host',
    'port',
    'connect_timeout',
    'autocommit',
    'database',
}
         
class MyConnection:
    def __init__(self, *args, **kwargs):
        self.args = args
        self.kwargs = kwargs
        self.session = dict()
        self.conn = psycopg2.connect(**kwargs)
        #self.conn.set_session(autocommit=autocommit)

    def set_session(self, **kwargs):
        self.session = kwargs
        if self.conn:
            self.conn.set_session(**kwargs)

    def ping(self):
        try:
            with self.conn.cursor() as c:
                c.execute("select 1")
        except Exception as e:
            self.conn = psycopg2.connect(**self.kwargs)
            self.conn.set_session(**self.session)
    
    def __getattr__(self, __name: str) :
        if __name in ("close", "commit", "rollback", "cursor"):
            return self.conn.__getattribute__(__name)
        raise AttributeError



def connect(*args, **kwargs):
    keys = list(kwargs.keys())
    for k in keys:
        if k not in avaliable_kw:
            if k in rename_kw:
                kwargs[rename_kw[k]] = kwargs[k]
            del kwargs[k]
    
    if "database" in kwargs:
        schema = kwargs["database"]
        kwargs["options"] = f'-c search_path={schema}'
    kwargs["database"] = "proton"
    kwargs["client_encoding"] = "utf-8"

    autocommit=0
    if 'autocommit' in kwargs:
        autocommit=kwargs['autocommit']
        del kwargs['autocommit']
    

    #conn = psycopg2.connect(sslmode='disable', **kwargs)
    #conn.set_session(autocommit=autocommit)
    conn = MyConnection(sslmode='disable',**kwargs)
    conn.set_session(autocommit=autocommit)
    return conn
    


def process_last_row_id(last_row_id):
    return last_row_id