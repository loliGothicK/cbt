package core_test

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/LoliGothick/cbt/internal/core"
	"github.com/LoliGothick/cbt/internal/test"
)

func TestWandboxGo(t *testing.T) {
	cases := []test.TestPattern{
		{In: `cbt wandbox go ./test_samples/go/simple.go`, Out: `Hello, cbt`, Err: nil},
		{In: `cbt wandbox go ./test_samples/go/simple.go -x=go-1.9`, Out: `Hello, cbt`, Err: nil},
		{In: `cbt wandbox go ./test_samples/go/runtime_option.go -r="hoge"`, Out: `hoge`, Err: nil},
		{In: `cbt wandbox go ./test_samples/go/stdin.go -in="hoge"`, Out: `hoge`, Err: nil},
	}
	for _, test := range cases {
		out, err := core.NewCLI().TestRun(strings.Split(test.In, " "))
		if !regexp.MustCompile(test.Out).Match(out) || !reflect.DeepEqual(test.Err, err) {
			t.Errorf("cbt (%q)\nout: %v, %v\nrequire: %v, %v",
				test.In, string(out), err, test.Out, test.Err)
		}
	}
}
