package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"testing"

	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

type TestingTVBC interface {
	Fatalf(string, ...interface{})
}

func PrintStack() {
	buf := []string{}
	for i := 2; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buf = append(buf, (fmt.Sprintf("    %s:%d: %d\n", file, line, pc)))
	}
	for i := len(buf) - 3; i >= 0; i-- {
		os.Stderr.WriteString(buf[i])
	}
}

// GetSourceBranch get source branch name
func GetSourceBranch() string {
	return os.Getenv("TestTaskRunnerDataSourceBranch")
}

// TestingT wrapper *testing.T
type TestingT struct {
	T TestingTVBC
}

// Assert interface
func (t *TestingT) Assert(want, get interface{}) {
	if !reflect.DeepEqual(want, get) {
		PrintStack()
		t.T.Fatalf("want %#v, get %#v", want, get)
	}
}

// AssertNil is nil
func (t *TestingT) AssertNil(input interface{}) {
	if input == nil {
		return
	}
	kind := reflect.ValueOf(input).Kind()
	switch kind {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan:
		if reflect.ValueOf(input).IsNil() {
			return
		}
	}
	PrintStack()
	if err, ok := input.(*trait.Error); ok {
		t.T.Fatalf("want nil, get error %s", err.Error())
	} else {
		t.T.Fatalf("want nil, get %#v", input)
	}
}

// AssertError error
func (t *TestingT) AssertError(want int, get *trait.Error) {
	if !trait.IsInternalError(get, want) {
		PrintStack()
		t.T.Fatalf("want %#v, get error: %#v", want, get)
	}
}

// AssertArray array
func (t *TestingT) AssertArray(want, get []interface{}) {
	if len(want) != len(get) {
		PrintStack()
		t.T.Fatalf("want %#v, get %#v", want, get)
	}

	for i, v := range want {
		if get[i] != v {
			PrintStack()
			t.T.Fatalf("%d, want %#v, get %#v", i, v, get[i])
		}
	}
}

// Assert interface
// TODO abadon
func Assert(t *testing.T, want, get interface{}) {
	if want != get {
		PrintStack()
		t.Fatalf("want %#v, get %#v", want, get)
	}
}

// AssertArray array
// TODO abadon
func AssertArray(t *testing.T, want, get []interface{}) {
	if len(want) != len(get) {
		PrintStack()
		t.Fatalf("want %#v, get %#v", want, get)
	}

	for i, v := range want {
		if get[i] != v {
			PrintStack()
			t.Fatalf("%d, want %#v, get %#v", i, v, get[i])
		}
	}
}

// TestHttpJson test http request with gin engine
func TestHttpJson(t *testing.T, e *gin.Engine, url, method string, body any, want int, receiver any, headers map[string]string) {
	tt := &TestingT{T: t}
	bs, rerr := json.Marshal(body)
	tt.AssertNil(rerr)
	req, rerr := http.NewRequest(method, url, bytes.NewReader(bs))
	tt.AssertNil(rerr)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res := testHander(req, e)
	defer res.Body.Close()

	bs, rerr = io.ReadAll(res.Body)
	tt.AssertNil(rerr)

	if want != res.StatusCode {
		PrintStack()
		tt.T.Fatalf("want %d, get %d, resp: %s", want, res.StatusCode, string(bs))
	}
	if receiver != nil {
		if err := json.Unmarshal(bs, receiver); err != nil {
			tt.AssertNil(&trait.Error{
				Internal: trait.ErrComponentDecodeError,
				Err:      err,
				Detail:   string(bs),
			})
		}
	}
}

func testHander(req *http.Request, e *gin.Engine) *http.Response {
	r := httptest.NewRecorder()
	ctx0, _ := gin.CreateTestContext(r)
	ctx0.Request = req
	e.ServeHTTP(r, req)
	res := r.Result()
	return res
}
