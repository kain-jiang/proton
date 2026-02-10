package codes

import (
	_ "embed"
	"encoding/json"

	ecode "taskrunner/error"
)

//go:embed code_cache.json
var errorcodes []byte

var ErrorCache ecode.ErrorCache

func init() {
	if err := json.Unmarshal(errorcodes, &ErrorCache.Errors); err != nil {
		panic(err)
	}
}
