package stmt

type KDB9Stmt struct {
	KDB9DML
	KDB9Admin
	MysqlStmt
}

type KDB9Admin struct {
	CreateDataBaseStmt string
	DeleteDataBaseStmt string
	MysqlAdmin
}

type KDB9DML struct {
	MysqlDML
}

func NewKDB9Stmt() KDB9Stmt {
	s := KDB9Stmt{}
	s.MysqlStmt = NewMysqlStmt()
	s.KDB9DML.MysqlDML = s.MysqlStmt.MysqlDML
	s.KDB9Admin.MysqlAdmin = s.MysqlStmt.MysqlAdmin

	s.KDB9Admin.CreateDataBaseStmt = `CREATE SCHEMA %s;`
	s.KDB9Admin.DeleteDataBaseStmt = `DROP SCHEMA IF EXISTS %s CASCADE;`
	return s
}
