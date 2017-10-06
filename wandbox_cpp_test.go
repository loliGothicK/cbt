package main

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

type TestPattern struct {
	in  string
	out string
	err error
}

func TestWandboxCpp(t *testing.T) {
	cases := []TestPattern{
		{`cbt wandbox cpp ./test_samples/cpp/simple_test.cpp`, `hello cbt`, nil},
	}
	for _, test := range cases {
		out, err := NewCLI().TestRun(strings.Split(test.in, " "))
		if !regexp.MustCompile(test.out).Match(out) || !reflect.DeepEqual(test.err, err) {
			t.Errorf("cbt (%q)\nout: %v, %v\nrequire: %v, %v",
				test.in, string(out), err, test.out, test.err)
		}
	}
}
