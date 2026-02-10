package stmt

type MysqlStmt struct {
	MysqlDML
	MysqlAdmin
}

type MysqlAdmin struct {
	CaculatePassword   string
	CreateDataBaseStmt string
	CreateUser         string
	CreateMysqlUser    string
	DeleteDataBaseStmt string
	DeleteUser         string
	GrantUserDB        string
	RevokeUserDB       string
	RwPrivileges       string
	UpdateUser         string
	UpdateMysqlUser    string
}

type MysqlDML struct {
	AddEdgeStmt                  string
	ChangeEdgeToStmt             string
	ChangeEdgeFromStmt           string
	GetEdgeStmt                  string
	CountAPPRelateInsStmt        string
	CountEdgeToStmt              string
	CountJobLogStmt              string
	DeleteAPPComponentsStmt      string
	DeleteAPPInsComponentInsStmt string
	DeleteAPPStmt                string
	DeleteAppinsStatusStmt       string
	DeleteEdgeFromStmt           string
	DeleteEdgeStmt               string
	DeleteEdgetoStmt             string
	DeleteJobLogStmt             string
	DeleteJobRecordStmt          string
	GetAPPComponentStmt          string
	GetAPPComponentsStmt         string
	GetAPPInsComponentInsStmt    string
	GetAPPInsStmt                string
	GetAPPStmt                   string
	GetAPPIDStmt                 string
	GetChangeEdgeToConflictStmt  string
	GetComponentInsStmt          string
	GetComponentLockStmt         string
	GetInsertAPPInsStmt          string
	GetInsertAPPStmt             string
	GetInsertComponentInsStmt    string
	GetInsertJobRecordStmt       string
	GetInsertSystemInfoStmt      string
	GetJobRecordStmt             string
	GetPointFromStmt             string
	GetPointToStmt               string
	GetSystemByNameStmt          string
	GetSystemInfoStmt            string
	GetWorkComponentInsStmt      string
	GetworkAPPInsStmt            string
	InsertAPPComponentStmt       string
	InsertAPPInsStmt             string
	InsertAPPStmt                string
	InsertComponentInsStmt       string
	InsertJobLogStmt             string
	InsertJobRecordStmt          string
	InsertSystemInfoStmt         string
	LayOffAPPInsStmt             string
	LayOffAPPInsByidStmt         string
	LayoffComponentInsStmt       string
	ListAPPStmt                  string
	ListJobLogStmt               string
	ListJobRecordCount           string
	ListJobRecordStmt            string
	ListSystemAPPNoWorked        string
	ListSystemInfoStmt           string
	ListSystemWithNameInfoStmt   string
	// ListworkAPPInsStmt           string
	// ListworkAPPInstCount         string
	LockComponentStmt            string
	SearchAPPStmt                string
	UnlockComponentStmt          string
	UnlockJobComponentStmt       string
	UpdateAPPInsConfigStmt       string
	UpdateAPPInsStatusStmt       string
	UpdateComponentInsStatusStmt string
	UpdateComponentInsStmt       string
	UpdateSystemInfoStmt         string
	WorkAPPInsStmt               string
	WorkComponentInsStmt         string

	InsertConfigTemplate      string
	GetConfigTemplateID       string
	UpdateConfigTemplate      string
	DeleteConfigTemplate      string
	GetConfigTemplate         string
	ListConfigTemplateFields  string
	ListConfigTemplate        string
	CountConfigTemplate       string
	InsertConfigTemplateIndex string
	DeleteConfigTemplateIndex string

	GetProtonComponent    string
	InsertProtonComponent string
	UpdateProtonComponent string

	InsertComposeJob        string
	UpdateComposeJobProcess string
	UpdateComposeJobStatus  string
	SetComposeJob           string
	InserComposeJobTask     string
	GetCompoesJobTask       string
	GetCompoesJobTasks      string
	DeleteComposeJobTask    string
	UpdateComposeJobTask    string
	DeleteComposeJobTasks   string
	// ListComposeJobTasks     string

	IDENTITY string

	InsertAppLang string
	UpdateAppLang string
	GetAppLang    string

	GetAppLock string
	LockApp    string
	UnlockApp  string

	InsertComposeManifests string
	GetComposeManifests    string
	ListComposeManifest    string

	InsertWorkComposeManifests    string
	GetWorkComposeJobManifests    string
	DeleteWorkComposeJobManifests string
	ListWorkComposeJobManifests   string
}

func NewMysqlStmt() MysqlStmt {
	s := MysqlStmt{}

	s.MysqlDML.AddEdgeStmt = `INSERT INTO task_edge (cfrom,cto) VALUES (?,?);`
	s.MysqlDML.GetEdgeStmt = `SELECT COUNT(1) FROM task_edge WHERE cfrom=? AND cto=?;`
	s.MysqlDML.DeleteAPPComponentsStmt = `DELETE FROM task_component WHERE aid=?;`
	s.MysqlDML.GetWorkComponentInsStmt = `SELECT ins_id FROM task_work_component WHERE sid=? AND cname=?;`
	s.MysqlDML.CountAPPRelateInsStmt = `SELECT COUNT(ins_id) FROM task_application_instance WHERE aid=?;`
	s.MysqlDML.UpdateAPPInsStatusStmt = `UPDATE task_application_instance SET status=?, owner=?, start_time=?, end_time=? WHERE ins_id=?;`
	s.MysqlDML.DeleteEdgeStmt = `DELETE FROM task_edge WHERE (cfrom=?) AND (cto=?);`
	s.MysqlDML.GetPointToStmt = `SELECT cfrom FROM task_edge WHERE cto=?;`
	s.MysqlDML.UnlockComponentStmt = `DELETE FROM task_component_lock WHERE jid=? AND sid=? AND cname=?;`
	s.MysqlDML.UpdateComponentInsStmt = `UPDATE task_component_instance set status=?, config=?, c_attribute=?, timeout=? WHERE ins_id=?;`
	s.MysqlDML.GetAPPComponentStmt = `SELECT aid, cname, cversion, ctype, crtype, ctimeout, cconfig, cattribute, spec FROM task_component WHERE cid=?;`
	s.MysqlDML.GetChangeEdgeToConflictStmt = `SELECT t1.cfrom, t1.cto
FROM task_edge t1
JOIN task_edge t2 ON t1.cfrom = t2.cfrom AND t2.cto = ?
WHERE t1.cto = ?;
`
	s.MysqlDML.GetComponentInsStmt = `SELECT ains_id, acid, sid, cname, ctype, crtype, version, aname, status, config, c_attribute, timeout, revission, create_time, start_time, end_time FROM task_component_instance WHERE ins_id=?;`
	s.MysqlDML.InsertAPPStmt = `INSERT INTO task_application (app_type, version, aname, aschema, graph, adependence) VALUES (?,?,?,?,?,?);`
	s.MysqlDML.LayoffComponentInsStmt = `DELETE FROM task_work_component WHERE ins_id=?;`
	s.MysqlDML.GetInsertComponentInsStmt = `SELECT ins_id FROM task_component_instance WHERE ains_id=? AND acid=?;`
	s.MysqlDML.GetInsertSystemInfoStmt = `SELECT sid FROM task_system WHERE namespace=? AND sname=?;`
	s.MysqlDML.GetworkAPPInsStmt = `SELECT ins_id FROM task_work_application WHERE (sid=?) and (aname=?);`
	s.MysqlDML.UnlockJobComponentStmt = `DELETE FROM task_component_lock WHERE jid=?;`
	s.MysqlDML.UpdateComponentInsStatusStmt = `UPDATE task_component_instance set status=?, start_time=?, end_time=?, revission=? WHERE ins_id=? AND revission=?;`
	s.MysqlDML.WorkAPPInsStmt = `INSERT INTO task_work_application (ins_id, aid, sid, version, aname, icomment, create_time, start_time, end_time, status) VALUES (?,?,?,?,?,?,?,?,?,?);`
	s.MysqlDML.WorkComponentInsStmt = `INSERT INTO task_work_component (ins_id, sid, cname, version, create_time, start_time, end_time) VALUES (?,?,?,?,?,?,?);`
	s.MysqlDML.DeleteAPPStmt = `DELETE FROM task_application WHERE aid=?;`
	s.MysqlDML.DeleteJobLogStmt = `DELETE FROM task_job_log WHERE jlid=?;`
	s.MysqlDML.ListSystemInfoStmt = `SELECT namespace, sname, sid from task_system WHERE sid>? LIMIT ?;`
	s.MysqlDML.DeleteEdgeFromStmt = `DELETE FROM task_edge WHERE cfrom=?;`
	s.MysqlDML.ListAPPStmt = `SELECT MAX(aid), aname FROM task_application WHERE aid>? GROUP BY aname ORDER BY MAX(aid) DESC LIMIT ? ;`
	s.MysqlDML.ListJobLogStmt = `SELECT jlid, jid, cid, aiid, aname, cname, ecode, ltime, msg FROM task_job_log %s;`
	s.MysqlDML.InsertAPPInsStmt = `INSERT INTO task_application_instance (aid,sid,version,aname,status,config, owner, icomment, create_time, start_time, end_time, otype, atrait) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?);`
	s.MysqlDML.ListSystemAPPNoWorked = `SELECT MAX(aid), aname FROM task_application WHERE aid>? AND aname NOT IN (SELECT DISTINCT aname FROM task_work_application WHERE sid=? ) GROUP BY aname ORDER BY MAX(aid) DESC LIMIT ?;`
	s.MysqlDML.UpdateSystemInfoStmt = `UPDATE task_system SET sname=?, config=?, sdescription=? WHERE sid=?;`
	s.MysqlDML.CountEdgeToStmt = `SELECT COUNT(cto) FROM task_edge WHERE cto=?;`
	s.MysqlDML.DeleteAppinsStatusStmt = `DELETE FROM task_application_instance WHERE ins_id=?;`
	s.MysqlDML.InsertAPPComponentStmt = `INSERT INTO task_component (aid, cname, cversion, ctype, crtype, ctimeout, cconfig, cattribute, spec) VALUES (?,?,?,?,?,?,?,?,?);`
	s.MysqlDML.InsertJobRecordStmt = `INSERT INTO task_job (target_id, current_id) VALUES (?,?);`
	s.MysqlDML.LayOffAPPInsStmt = `DELETE FROM task_work_application WHERE aname=? AND sid=?;`
	s.MysqlDML.LayOffAPPInsByidStmt = `DELETE FROM task_work_application WHERE ins_id=?;`
	s.MysqlDML.DeleteAPPInsComponentInsStmt = `DELETE FROM task_component_instance WHERE ains_id=?;`
	s.MysqlDML.DeleteJobRecordStmt = `DELETE FROM task_job WHERE jid=?;`
	s.MysqlDML.GetInsertJobRecordStmt = `SELECT jid from task_job WHERE target_id=? AND current_id=?;`
	s.MysqlDML.GetSystemInfoStmt = `SELECT  namespace, sname, config, sdescription from task_system WHERE sid=?;`
	s.MysqlDML.InsertJobLogStmt = `INSERT INTO task_job_log (jid, cid, aiid, aname, cname, ecode, ltime, msg) VALUES (?,?,?,?,?,?,?,?);`
	s.MysqlDML.GetAPPInsStmt = `SELECT aid, sid , version, aname, status, config, owner, icomment, create_time, start_time, end_time, otype, atrait FROM task_application_instance WHERE ins_id=?;`
	s.MysqlDML.GetAPPStmt = `SELECT app_type, version, aname, aschema, graph, adependence FROM task_application WHERE aid=?;`
	s.MysqlDML.GetPointFromStmt = `SELECT cto FROM task_edge WHERE cfrom=?;`
	s.MysqlDML.GetComponentLockStmt = `SELECT jid FROM task_component_lock WHERE sid=? AND cname=?;`
	s.MysqlDML.GetInsertAPPInsStmt = `SELECT ins_id FROM task_application_instance WHERE aid=? AND sid=? AND version=? AND aname=? AND status=? AND owner=? AND icomment=? AND create_time=? AND start_time=? AND end_time=?;`
	s.MysqlDML.LockComponentStmt = `INSERT INTO task_component_lock (jid, sid, cname) VALUES (?,?,?);`
	s.MysqlDML.ChangeEdgeToStmt = `UPDATE task_edge SET cto=? WHERE cto=?;`
	s.MysqlDML.ChangeEdgeFromStmt = `UPDATE task_edge SET cfrom=? WHERE cfrom=?;`
	s.MysqlDML.CountJobLogStmt = `SELECT count(jlid) FROM task_job_log %s;`
	s.MysqlDML.DeleteEdgetoStmt = `DELETE FROM task_edge WHERE cto=?;`
	s.MysqlDML.SearchAPPStmt = `SELECT aid, version FROM task_application WHERE (aname=?) AND (aid>?) ORDER BY aid LIMIT ? ;`
	s.MysqlDML.UpdateAPPInsConfigStmt = `UPDATE task_application_instance SET config=?, icomment=? WHERE ins_id=?;`
	s.MysqlDML.GetSystemByNameStmt = `SELECT namespace, sname, sid, config from task_system WHERE sname=?;`
	s.MysqlDML.InsertSystemInfoStmt = `INSERT INTO task_system (namespace, sname, config, sdescription) VALUES (?,?,?,?); `
	// s.MysqlDML.ListJobRecordCount = `SELECT count(j.jid) FROM task_job as j, task_application_instance as a WHERE j.target_id=a.ins_id %s ;`
	// s.MysqlDML.ListworkAPPInstCount = `SELECT count(ins_id) FROM task_work_application WHERE %s;`
	s.MysqlDML.GetInsertAPPStmt = `SELECT aid from task_application WHERE aname=? AND version=?;`
	s.MysqlDML.InsertComponentInsStmt = `INSERT INTO task_component_instance (ains_id, acid, sid, version, aname, cname, ctype, crtype, status, config, c_attribute, timeout, revission, create_time, start_time, end_time) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);`
	s.MysqlDML.ListSystemWithNameInfoStmt = `SELECT namespace, sname, sid from task_system WHERE (sid>?) AND (sname=?) LIMIT ?;`
	// s.MysqlDML.ListJobRecordStmt = `
	// SELECT
	// j.jid, j.target_id, j.current_id,
	// a.aid, a.aname, a.status, a.sid, a.version,
	// a.icomment, a.create_time, a.start_time,
	// a.end_time, a.otype
	// FROM task_job as j, task_application_instance as a
	// WHERE j.jid <= (
	// 	SELECT j.jid FROM
	// 	task_job as j, task_application_instance as a
	// 	WHERE j.target_id=a.ins_id %s ORDER BY j.jid DESC LIMIT ?,1
	// ) AND j.target_id=a.ins_id %s ORDER BY j.jid DESC LIMIT ?;`
	s.MysqlDML.ListJobRecordCount = `
	SELECT count(j.jid) 
	FROM task_job j JOIN task_application_instance  a ON j.target_id=a.ins_id %s ;`

	s.MysqlDML.ListJobRecordStmt = `
	SELECT j.jid, j.target_id, j.current_id, 
	a.aid, a.aname, a.status, a.sid, a.version, 
	a.icomment, a.create_time, a.start_time, 
	a.end_time, a.otype, s.sname 
	FROM task_job j
	JOIN task_application_instance a ON j.target_id=a.ins_id
	JOIN task_system s ON a.sid=s.sid 
	%s ORDER BY jid DESC LIMIT ? offset %d`

	// s.MysqlDML.ListworkAPPInsStmt = `SELECT ins_id, aid, version, aname, icomment, create_time, start_time, end_time, status FROM task_work_application WHERE ins_id <= (SELECT ins_id FROM task_work_application WHERE %s ORDER BY ins_id DESC LIMIT ?,1) AND %s ORDER BY ins_id DESC LIMIT ?;`
	s.MysqlDML.GetAPPComponentsStmt = `SELECT cid, cname, cversion, ctype, crtype, ctimeout, cconfig, cattribute, spec FROM task_component WHERE aid=?;`
	s.MysqlDML.GetAPPInsComponentInsStmt = `SELECT ins_id, acid, sid, cname, ctype, crtype, version, aname, status, config, c_attribute, timeout, revission, create_time, start_time, end_time FROM task_component_instance WHERE ains_id=?;`
	s.MysqlDML.GetJobRecordStmt = `SELECT target_id, current_id FROM task_job WHERE jid=?;`
	s.MysqlAdmin.RwPrivileges = `SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, REFERENCES, INDEX, ALTER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, EVENT, TRIGGER`
	s.MysqlAdmin.UpdateUser = `ALTER USER '%s'@'%%' IDENTIFIED BY '%s';`
	s.MysqlAdmin.UpdateMysqlUser = `ALTER USER IF EXISTS '%s'@'%%' IDENTIFIED with mysql_native_password BY '%s';`
	s.MysqlAdmin.CreateUser = `CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s';`
	s.MysqlAdmin.CreateMysqlUser = `CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED with mysql_native_password BY '%s';`
	s.MysqlAdmin.CreateDataBaseStmt = `CREATE DATABASE IF NOT EXISTS %s`
	s.MysqlAdmin.DeleteDataBaseStmt = `DROP DATABASE %s;`
	s.MysqlAdmin.DeleteUser = `DROP USER '%s'@'%%';`
	s.MysqlAdmin.GrantUserDB = `GRANT %s ON %s.* TO '%s'@'%%'; FLUSH PRIVILEGES;`
	s.MysqlAdmin.RevokeUserDB = `REVOKE ALL ON *.* FROM '%s'@'%%';`

	s.InsertConfigTemplateIndex = `INSERT INTO task_app_config_template_index (tid, label) VALUES (?,?);`
	s.DeleteConfigTemplateIndex = `DELETE FROM task_app_config_template_index WHERE tid=?;`

	s.InsertConfigTemplate = `INSERT INTO task_app_config_template (tversion, tname, tdescription, aname, aversion, config, labels, minversion, maxversion) VALUES (?,?,?,?,?,?,?,?,?);`
	s.GetConfigTemplateID = `SELECT tid from task_app_config_template WHERE tversion=? AND tname=? AND aname=?;`
	s.UpdateConfigTemplate = `UPDATE task_app_config_template SET tdescription=?, aversion=?, config=?, labels=?, minversion=?, maxversion=? WHERE tid=?;`
	s.DeleteConfigTemplate = `DELETE FROM task_app_config_template WHERE tid=?;`
	s.GetConfigTemplate = `SELECT tversion, tname, tdescription, aname, aversion, config, labels FROM task_app_config_template WHERE tid=?;`
	s.ListConfigTemplateFields = `ct.tid, ct.tversion, ct.tname, ct.tdescription, ct.aname, ct.aversion, ct.labels`
	s.ListConfigTemplate = `SELECT %s %s ORDER BY ct.tid DESC LIMIT ? OFFSET %d;`
	s.CountConfigTemplate = `SELECT COUNT(1) FROM (SELECT ct.tid %s) AS count;`

	s.InsertProtonComponent = `INSERT INTO task_proton_component (cname, ctype, sid, cattribute, coptions) VALUES (?,?,?,?,?);`
	s.UpdateProtonComponent = `UPDATE task_proton_component SET cattribute=? , coptions=? WHERE sid=? AND cname=? AND ctype=?;`

	s.GetAPPIDStmt = `SELECT aid FROM task_application WHERE aname=? and version=?;`
	s.IDENTITY = `SELECT LAST_INSERT_ID();`

	s.GetAppLock = `SELECT jid FROM task_job_lock WHERE aname=? AND sid=?;`
	s.LockApp = `INSERT INTO task_job_lock (jid, sid, aname) VALUES (?,?,?);`
	s.UnlockApp = `DELETE FROM task_job_lock WHERE aname=? AND sid=? AND jid=?;`

	s.InsertAppLang = `INSERT INTO task_app_lang (alang, aname, alias) VALUES (?,?,?);`
	s.UpdateAppLang = `UPDATE task_app_lang set alias=? WHERE alang=? AND aname=?;`
	s.GetAppLang = `SELECT alias FROM task_app_lang WHERE alang=? AND aname=?;`

	s.InsertComposeJob = "INSERT INTO task_compose_job (jname, sid, status, processed, total, config, mversion, create_time, start_time, end_time, mdescription) VALUES (?,?,?,?,?,?,?,?,?,?,?);"
	s.SetComposeJob = "UPDATE task_compose_job SET jname=?, sid=?, status=?, processed=?, total=?, config=? WHERE jid=?;"
	s.UpdateComposeJobProcess = "UPDATE task_compose_job set processed=? WHERE JID=?;"
	s.UpdateComposeJobStatus = "UPDATE task_compose_job set status=?, end_time=?%s WHERE JID=?;"
	s.InserComposeJobTask = "INSERT INTO task_compose_job_task (jid, jtindex, ajid) VALUES (?,?,?);"
	s.GetCompoesJobTask = "SELECT ajid FROM task_compose_job_task WHERE jid=? AND jtindex=?;"
	s.GetCompoesJobTasks = "SELECT ajid,jtindex FROM task_compose_job_task WHERE jid=? ORDER BY jtindex ASC;"
	s.DeleteComposeJobTask = "DELETE FROM task_compose_job_task WHERE jid=? AND jtindex=?;"
	s.UpdateComposeJobTask = "UPDATE task_compose_job_task SET ajid=? WHERE jid=? AND jtindex=?;"
	s.DeleteComposeJobTasks = "DELETE FROM task_compose_job_task WHERE jid=?;"

	s.InsertComposeManifests = `INSERT INTO task_compose_manifests (mname, mversion, manifests, mdescription) VALUES (?,?,?,?);`
	s.GetComposeManifests = `SELECT mname, mversion, mdescription, manifests FROM task_compose_manifests WHERE mname=? AND mversion=?;`
	s.ListComposeManifest = `SELECT %s FROM task_compose_manifests %s LIMIT ? OFFSET %d`

	s.ListWorkComposeJobManifests = `SELECT %s FROM task_work_compose_manifests %s ORDER BY jid DESC LIMIT ? OFFSET %d;`
	s.InsertWorkComposeManifests = `INSERT INTO task_work_compose_manifests (jid, mname, mversion, sid, status, mdescription) VALUES (?,?,?,?,?,?);`
	s.GetWorkComposeJobManifests = `SELECT jid, mname, mversion, sid FROM task_work_compose_manifests WHERE mname=? AND sid=?;`
	s.DeleteWorkComposeJobManifests = `DELETE FROM task_work_compose_manifests WHERE mname=? AND sid=?;`
	return s
}
