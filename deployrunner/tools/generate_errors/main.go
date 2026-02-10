package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	terror "taskrunner/error"
)

func main() {
	workDir, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	s := terror.NewErrorStore(workDir)
	s.LoadIndex()
	defer s.Close()
	if err := filepath.Walk(workDir, func(path string, info fs.FileInfo, err error) error {
		if info.Name() != "code.json" {
			return nil
		}
		fmt.Println(info.Name())
		mname := filepath.Dir(path)
		s.IndexModule(mname)
		errs := []terror.ErrorCode{}
		bs, err0 := os.ReadFile(path)
		if err0 != nil {
			return err0
		}
		if err0 := json.Unmarshal(bs, &errs); err0 != nil {
			return err0
		}
		for _, e := range errs {
			s.IndexError(mname, e)
		}
		return nil
	}); err != nil {
		defer s.Close()
		panic(err)
	}
	s.Dumps()
	s.DumpsCode()
	s.DumpErrorCache()
	s.Close()
}
