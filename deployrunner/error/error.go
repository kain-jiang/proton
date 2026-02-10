package error

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	// "sort"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

type ErrorCode struct {
	Code        int
	MCode       int
	Name        string
	Description string
}

func (e *ErrorCode) RealCode() int {
	return ((e.MCode & 0xffff) << 16) | e.Code&0xffff
}

type ErrorCache struct {
	Errors [][]ErrorCode
}

func newErrorCache(modules []module, errs []ErrorCode) ErrorCache {
	ec := ErrorCache{
		Errors: make([][]ErrorCode, len(modules)),
	}

	for _, e := range errs {
		ec.Errors[e.MCode] = append(ec.Errors[e.MCode], e)
	}

	for _, es := range ec.Errors {
		sort.Slice(es, func(i, j int) bool {
			return es[i].Code < es[j].Code
		})
	}

	return ec
}

func (e *ErrorCache) GetCode(code int) ErrorCode {
	ecode := code & 0xffff
	mcode := (code >> 16) & 0xffff
	return e.Errors[mcode][ecode]
}

type module struct {
	mcode  int
	mname  string
	errors map[string]ErrorCode
}

// i is module index, j is code index fo Error[i][j]
type errorIndex struct {
	moduleIndex map[string]module
	modules     []module
}

func newErrorIndex(modules []module, errs []ErrorCode) errorIndex {
	index := errorIndex{
		moduleIndex: make(map[string]module),
		modules:     make([]module, len(modules)),
	}
	for _, m := range modules {
		index.modules[m.mcode] = m
	}

	for _, e := range errs {
		m := index.modules[e.MCode]
		m.errors[e.Name] = e
	}

	return index
}

func (e *errorIndex) IndexModule(mname string) module {
	if m, ok := e.moduleIndex[mname]; ok {
		return m
	} else {
		m := module{
			mcode:  len(e.moduleIndex),
			mname:  mname,
			errors: make(map[string]ErrorCode),
		}
		e.moduleIndex[mname] = m
		return m
	}
}

func (e *errorIndex) IndexError(mname string, err ErrorCode) ErrorCode {
	m := e.IndexModule(mname)
	if err0, ok := m.errors[err.Name]; ok {
		err0.Description = err.Description
		return err0

	}
	err.Code = len(m.errors)
	err.MCode = m.mcode
	m.errors[err.Name] = err
	return err
}

type ErrorStore struct {
	*sql.DB
	errorIndex
	codeDir string
}

type cursor interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
}

func NewErrorStore(fpath string) ErrorStore {
	db, err := sql.Open("sqlite3", filepath.Join(fpath, "error_codes.sqlite3"))
	if err != nil {
		panic(fmt.Sprintf("open error code file fail: %s", err.Error()))
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS module (
		mcode INTERGER PRIMARY KEY,
		mname TEXT UNIQUE
		);`); err != nil {
		panic(err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS errorcode (
		code INTERGER PRIMARY KEY,
		mcode INTERGER,
		name TEXT UNIQUE,
		description TEXT
	);`); err != nil {
		panic(err)
	}
	s := ErrorStore{
		DB:      db,
		codeDir: fpath,
	}
	return s
}

func (e *ErrorStore) Close() {
	if err := e.DB.Close(); err != nil {
		panic(err)
	}
}

func (e *ErrorStore) loadErr() ([]module, []ErrorCode) {
	db := e.DB
	errs := make([]ErrorCode, 0)
	rows, err := db.Query("SELECT code, mcode, name, description FROM errorcode;")
	if err != nil {
		panic(fmt.Sprintf("read error code file fail: %s", err.Error()))
	}

	for rows.Next() {
		err0 := ErrorCode{}
		if err := rows.Scan(&err0.Code, &err0.MCode, &err0.Name, &err0.Description); err != nil {
			panic(fmt.Sprintf("read row from error code sqlite file fail: %s", err.Error()))
		}
		errs = append(errs, err0)
	}

	modules := make([]module, 0)
	rows, err = db.Query("SELECT mcode, mname FROM module;")
	if err != nil {
		panic(fmt.Sprintf("read module code file fail: %s", err.Error()))
	}

	for rows.Next() {
		m := module{}
		if err := rows.Scan(&m.mcode, &m.mname); err != nil {
			panic(fmt.Sprintf("read row from mode code sqlite file fail: %s", err.Error()))
		}
		m.errors = make(map[string]ErrorCode)
		modules = append(modules, m)
	}
	return modules, errs
}

func (e *ErrorStore) LoadIndex() {
	modules, errs := e.loadErr()
	e.errorIndex = newErrorIndex(modules, errs)
}

func (e *ErrorStore) loadCache() ErrorCache {
	modules, errs := e.loadErr()
	return newErrorCache(modules, errs)
}

func (e *ErrorStore) getModuleID(db cursor, mname string) module {
	res := module{
		mname: mname,
	}
	row := db.QueryRow("SELECT mcode FROM module WHERE mname=?", mname)
	if err := row.Scan(&res.mcode); err != nil {
		if err == sql.ErrNoRows {
			res.mcode = -1
			return res
		}
		panic(err)
	}
	return res
}

func (e *ErrorStore) newModule(db cursor, m module) int {
	res, err := db.Exec("INSERT INTO module (mcode, mname) VALUES (?,?);", m.mcode, m.mname)
	if err != nil {
		panic(fmt.Sprintf("%d %s: %s", m.mcode, m.mname, err))
	}
	// sqlite3 last insert
	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	return int(id)
}

func (e *ErrorStore) getError(db cursor, name string) ErrorCode {
	res := ErrorCode{
		Name: name,
	}
	row := db.QueryRow("SELECT mcode, code, name, description FROM errorcode WHERE name=?", name)
	if err := row.Scan(&res.MCode, &res.Code, &res.Name, &res.Description); err != nil {
		if err == sql.ErrNoRows {
			res.Code = -1
			return res
		}
		panic(err)
	}
	return res
}

func (e *ErrorStore) newError(db cursor, ec ErrorCode) int {
	res, err := db.Exec("INSERT INTO errorcode (mcode, code, name, description) VALUES (?,?,?,?);", ec.MCode, ec.Code, ec.Name, ec.Description)
	if err != nil {
		panic(err)
	}
	// sqlite3 last insert
	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	return int(id)
}

func (e *ErrorStore) updateError(db cursor, ec ErrorCode) {
	_, err := db.Exec("UPDATE errorcode SET description=? WHERE name=?;", ec.Description, ec.Name)
	if err != nil {
		panic(err)
	}
}

func (e *ErrorStore) Dumps() {
	db, err := e.DB.Begin()
	if err != nil {
		panic(err)
	}

	for k, v := range e.moduleIndex {
		m := e.getModuleID(db, k)
		if m.mcode == -1 {
			e.newModule(db, v)
		}

		for _, v0 := range v.errors {
			e0 := e.getError(db, v0.Name)
			if e0.Code == -1 {
				e.newError(db, v0)
				e0 = v0
			} else {
				e.updateError(db, v0)
			}
		}

	}

	if err := db.Commit(); err != nil {
		panic(err)
	}
}

func (e *ErrorStore) DumpsCode() {
	for _, m := range e.moduleIndex {
		if len(m.errors) <= 0 {
			return
		}
		file, err := os.OpenFile(filepath.Join(m.mname, "code.go"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			panic(err)
		}
		if _, err := file.WriteString("// Code generated by 'go run tools/generate_errors/main.go  error/codes'. DO NOT EDIT.\npackage code\n\nconst (\n"); err != nil {
			panic(err)
		}

		sorted := make([]ErrorCode, len(m.errors))
		for _, e0 := range m.errors {
			sorted[e0.Code] = e0
		}

		for _, e0 := range sorted {
			if _, err := file.WriteString(fmt.Sprintf("	%s = %d\n", e0.Name, e0.RealCode())); err != nil {
				panic(err)
			}
		}

		if _, err := file.WriteString(")"); err != nil {
			panic(err)
		}
		if err := file.Close(); err != nil {
			panic(err)
		}
	}
}

func (e *ErrorStore) DumpErrorCache() {
	ec := e.loadCache()
	bs, err := json.MarshalIndent(ec.Errors, "", " ")
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(filepath.Join(e.codeDir, "code_cache.json"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	if _, err := file.Write(bs); err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}
