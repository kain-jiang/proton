import os

db_type = os.environ.get('DB_TYPE', 'MYSQL').upper()

if db_type == 'DM8':
    from .dm import *
elif db_type == 'KDB9':
    from .kingbase import *
else:
    from .mysql import *