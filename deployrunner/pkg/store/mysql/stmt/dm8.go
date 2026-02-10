package stmt

type DM8Stmt struct {
	DM8DML
	DM8Admin
	MysqlStmt
}

type DM8Admin struct {
	CreateDataBaseStmt string
	DeleteDataBaseStmt string
	MysqlAdmin
}

type DM8DML struct {
	MysqlDML
}

func NewDM8Stmt() DM8Stmt {
	s := DM8Stmt{}
	s.MysqlStmt = NewMysqlStmt()
	s.DM8DML.MysqlDML = s.MysqlStmt.MysqlDML
	s.DM8Admin.MysqlAdmin = s.MysqlStmt.MysqlAdmin

	s.DM8Admin.MysqlAdmin.CreateUser = `CREATE USER IF NOT EXISTS "%s" IDENTIFIED BY "%s";`
	// s.DM8Admin.MysqlAdmin.CreateUser = `CREATE USER "%s" IDENTIFIED BY "%s";`
	s.DM8Admin.MysqlAdmin.DeleteUser = `DROP USER IF EXISTS "%s" CASCADE;`
	s.DM8Admin.MysqlAdmin.UpdateUser = `ALTER USER "%s" IDENTIFIED BY "%s";`
	s.DM8Admin.MysqlAdmin.GrantUserDB = `GRANT %s %s to %s; `
	s.DM8Admin.MysqlAdmin.RwPrivileges = `DBA`

	s.DM8Admin.CreateDataBaseStmt = `CREATE SCHEMA %s;`
	s.DM8Admin.DeleteDataBaseStmt = `DROP SCHEMA IF EXISTS %s CASCADE;`
	s.DM8Admin.CreateDataBaseStmt = `CREATE SCHEMA %s;`
	s.DM8Admin.DeleteDataBaseStmt = `DROP SCHEMA IF EXISTS %s CASCADE;`

	return s
}
