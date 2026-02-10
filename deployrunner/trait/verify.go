package trait

// VerifyRecord 返回单个job的所有验证记录
type VerifyRecord struct {
	// 数据验证的结果集合
	DataSchemaVerifyList []DataSchemaVerify `json:"dataSchemaVerifyList"`
	FuncVerifyList       []FunctionVerify   `json:"funcVerifyList"`
}

type DataSchemaVerify struct {
	VerifyResult  string `json:"verifyResult"`
	VerifyEndTime string `json:"verifyEndTime"`
	Did           int    `json:"did"`
}

type FunctionVerify struct {
	VerifyResult  string `json:"verifyResult"`
	VerifyEndTime string `json:"verifyEndTime"`
	Fid           int    `json:"fid"`
}

// DataTestEntry 单个数据测试用例结果
type DataTestEntry struct {
	// 测试用例id
	Tid int `json:"tid"`
	// 测试结果
	TestResult string `json:"testResult"`
	// 测试结果具体信息
	TestTesultDetail string `json:"testResultDetail"`
	// 测试的服务名称
	ServiceName string `json:"serviceName"`
}

// FunctionTestEntry 单个功能用例结果
type FunctionTestEntry struct {
	// 功能测试用例id
	Tid int `json:"tid"`
	// 测试名称
	TestFunctionName string `json:"testFunctionName"`
	// 测试描述
	TestDescription string `json:"testDescription"`
	// 测试结果，值为pass/fail
	TestResult string `json:"testResult"`
}
