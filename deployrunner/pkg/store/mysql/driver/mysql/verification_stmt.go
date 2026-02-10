package store

const (
	// verify record
	// 当前deployrunner只提供更新验证的查询接口，数据验证功能由data-model提供，模块功能验证由模块化服务对应的验证组件提供
	getDataVerifyRecordStmt     = `SELECT d_id, verify_result, verify_end_time FROM verification_data_records WHERE ai_id=? ORDER BY d_id DESC; `
	getFunctionVerifyRecordStmt = `SELECT f_id, verify_result, verify_end_time FROM verification_function_records WHERE ai_id=? ORDER BY f_id DESC;`
	getDataTestEntriesStmt      = `SELECT t_id, test_result, test_result_details, service_name FROM data_test_entries WHERE d_id=? ORDER BY t_id DESC LIMIT ? OFFSET ?;`
	CountDataTestEntriesStmt    = `SELECT count(t_id) FROM data_test_entries WHERE d_id=?`
	getFunctionTestEntriesStmt  = `SELECT t_id, test_function_name, test_description, test_result FROM function_test_entries WHERE f_id=? ORDER BY t_id DESC LIMIT ? OFFSET ?;`
	CountFunctionEntriesStmt    = `SELECT count(t_id) FROM function_test_entries WHERE f_id=?`
)
