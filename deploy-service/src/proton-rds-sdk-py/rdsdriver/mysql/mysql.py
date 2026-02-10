import pymysql.cursors
TupleCursor = pymysql.cursors.Cursor
DictCursor = pymysql.cursors.DictCursor

def process_last_row_id(last_row_id):
    return last_row_id