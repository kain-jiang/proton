package main

import (
	"os"
	"testing"

	"taskrunner/test"

	// load testcase
	_ "embed"
)

func TestUploadApplicationAT(t *testing.T) {
	config := os.Getenv("TASKRUNNER_UPLOAD_CONFIG")
	pckPath := os.Getenv("TASKRUNNER_UPLOAD_APP")
	if config == "" || pckPath == "" {
		t.SkipNow()
	}
	tt := test.TestingT{T: t}
	cmd := NewUploadCmd()
	cmd.SetArgs([]string{"--config", config, pckPath})
	err := cmd.Execute()
	tt.AssertNil(err)
}
