package store

import (
	"context"
	"time"

	"taskrunner/trait"
)

/*
+++++++++ basic verify impl +++++++++++++++++++++++++
*/

func (c *SQLCursor) GetVerifyRecord(ctx context.Context, jid int) (trait.VerifyRecord, *trait.Error) {
	vr := trait.VerifyRecord{}

	dataResults, err := c.QueryContext(ctx, getDataVerifyRecordStmt, jid)
	if err != nil {
		return vr, err
	}
	defer dataResults.Close()
	dvs := make([]trait.DataSchemaVerify, 0)
	for dataResults.Next() {
		var timestamp int64
		dv := trait.DataSchemaVerify{}
		if err0 := dataResults.Scan(&dv.Did, &dv.VerifyResult, &timestamp); err0 != nil {
			return vr, err0
		}
		timeObj := time.Unix(timestamp, 0)
		dv.VerifyEndTime = timeObj.Format("2006/01/02 15:04:05")

		dvs = append(dvs, dv)
	}
	funcResults, err := c.QueryContext(ctx, getFunctionVerifyRecordStmt, jid)
	if err != nil {
		return vr, err
	}
	defer funcResults.Close()
	fvs := make([]trait.FunctionVerify, 0)
	for funcResults.Next() {
		var timestamp int64
		fv := trait.FunctionVerify{}
		if err0 := funcResults.Scan(&fv.Fid, &fv.VerifyResult, &timestamp); err0 != nil {
			return vr, err0
		}
		timeObj := time.Unix(timestamp, 0)
		fv.VerifyEndTime = timeObj.Format("2006/01/02 15:04:05")
		fvs = append(fvs, fv)
	}
	vr.FuncVerifyList = fvs
	vr.DataSchemaVerifyList = dvs
	return vr, nil
}

func (c *SQLCursor) CountDataTestEntries(ctx context.Context, did int) (int, *trait.Error) {
	row := c.QueryRowContext(ctx, CountDataTestEntriesStmt, did)
	total := 0
	err := row.Scan(&total)
	return total, err
}

func (c *SQLCursor) GetDataTestEntries(ctx context.Context, did int, limit int, offset int) ([]trait.DataTestEntry, *trait.Error) {
	rows, err := c.QueryContext(ctx, getDataTestEntriesStmt, did, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dvs := make([]trait.DataTestEntry, 0)

	for rows.Next() {
		dv := trait.DataTestEntry{}
		// var jsonData string
		if err0 := rows.Scan(&dv.Tid, &dv.TestResult, &dv.TestTesultDetail, &dv.ServiceName); err0 != nil {
			return nil, err0
		}
		//if err = json.Unmarshal([]byte(jsonData), &dv.TestTesultDetail); err != nil {
		//	return nil, err
		//}

		dvs = append(dvs, dv)
	}

	return dvs, nil
}

func (c *SQLCursor) GetFunctionTestEntries(ctx context.Context, fid int, limit int, offset int) ([]trait.FunctionTestEntry, *trait.Error) {
	rows, err := c.QueryContext(ctx, getFunctionTestEntriesStmt, fid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fvs := make([]trait.FunctionTestEntry, 0)

	for rows.Next() {
		fv := trait.FunctionTestEntry{}
		if err0 := rows.Scan(&fv.Tid, &fv.TestFunctionName, &fv.TestDescription, &fv.TestResult); err0 != nil {
			return nil, err0
		}
		fvs = append(fvs, fv)
	}

	return fvs, nil
}

func (c *SQLCursor) CountFunctionTestEntries(ctx context.Context, fid int) (int, *trait.Error) {
	row := c.QueryRowContext(ctx, CountFunctionEntriesStmt, fid)
	total := 0
	err := row.Scan(&total)
	return total, err
}
