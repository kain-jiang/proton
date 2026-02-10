# 类图设计

```mermaid
classDiagram

class DBConn {
	<<interface>>
	Close() Error
	BeginTx() (Transaction, Error)
}

class CursorConn {
	<<interface>>
	Exec() Result, Error
	Query() Rows, Error
	QueryRow() Row
	Prepare() Stmt, Error
}

class Transaction {
	<<interface>>
	Commit() Error
	Rollback() Error
}

DBConn *-- CursorConn
Transaction *-- CursorConn

class Store {
	<<interface-op>>
}

class Tx {
	<<interface-op>>
}

class Cursor {
	<<interface-op>>
	SQLStatementsContainer
}

class Result {
	<<interface>>
}
class Rows {
	<<interface>>
}
class Row {
	<<interface>>
}
class Stmt {
	<<interface>>
}
Tx ..> Cursor
Store ..>Tx
Store ..> Cursor

DBConn --> Transaction: create
Store ..> DBConn
Tx ..> Transaction
Cursor ..> CursorConn


```

# mysqlDriver实现

```mermaid
classDiagram

class DBConn {
	<<interface>>
	Close() Error
	BeginTx() (Transaction, Error)
}

class CursorConn {
	<<interface>>
	Exec() Result, Error
	Query() Rows, Error
	QueryRow() Row
	Prepare() Stmt, Error
}

class Transaction {
	<<interface>>
	Commit() Error
	Rollback() Error
}

class RawCursorConn {
	<<interfaace>>
	Exec() sql.Result, error
	Query() sql.Rows, error
	QueryRow() sql.Row
	Prepare() sql.Stmt, error
}

class cursorConn {
	RawCursorConn
}

class transaction {
	sql.Tx
	CursorConn
}

cursorConn ..> RawCursorConn

CursorConn <|.. cursorConn

Transaction <|.. transaction
transaction o.. CursorConn
transaction o.. sqlTx
sqlTx ..|> RawCursorConn

class dbConn {
	CursorConn
	sql.DB
}

dbConn o.. sqlDB
dbConn o.. CursorConn
DBConn <|.. dbConn
sqlDB ..|> RawCursorConn



```


## job erDiagram
```mermaid
erDiagram

system ||-- o{ workAPP: compose
workAPP ||--|{ workComponent: owner
system || --o{ workComponent: compose

system ||--o{ JobRecord: "install/update job"
JobRecord ||--|| curAPP: cur
JobRecord ||--|| targetAPP: target

curAPP ||--|| APPins: is
targetAPP ||--|| APPins: is

APPins ||--|{ component: compose

workAPP ||--|| APPins: is
workComponent ||--|| component: is

edge ||--|| fromComponent : from
edge ||--|| toComponent: to
system ||--|{ edge: compose

fromComponent ||--|| component: is
toComponent ||--|| component: is

```