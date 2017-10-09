package core_test

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/LoliGothick/cbt/internal/core"
	"github.com/LoliGothick/cbt/internal/test"
)

func TestWandboxCpp(t *testing.T) {
	cases := []test.TestPattern{
		{In: `cbt wandbox cpp ./test_samples/cpp/simple_test.cpp`, Out: `Hello, cbt`, Err: nil},
	}
	for _, test := range cases {
		out, err := core.NewCLI().TestRun(strings.Split(test.In, " "))
		if !regexp.MustCompile(test.Out).Match(out) || !reflect.DeepEqual(test.Err, err) {
			t.Errorf("cbt (%q)\nout: %v, %v\nrequire: %v, %v",
				test.In, string(out), err, test.Out, test.Err)
		}
	}
}