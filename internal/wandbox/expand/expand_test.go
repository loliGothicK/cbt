package expand_test

import (
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/LoliGothick/cbt/internal/wandbox/expand"
)

func TestExpandInclude(t *testing.T) {
	code, codes := expand.ExpandInclude(`test_sample/main_.cpp`, `#include.*".*"|".*"/\*cbt-require\*/`)
	if !regexp.MustCompile(`main\(\)`).Match([]byte(code)) {
		t.Fatalf("main src fail:\n%v", code)
	}
	for key, val := range codes {
		b, err := ioutil.ReadFile(`test_sample/` + key)
		if err != nil {
			t.Fatalf("include open error: %v", err)
		}
		if !regexp.MustCompile(`#define ` + key).Match(b) {
			t.Fatalf("include expect fail:\n%v", val)
		}
	}
}

func TestExpandIncludeMulti(t *testing.T) {
	main, src, headers := expand.ExpandIncludeMulti([]string{`test_sample/main.cpp`, `test_sample/func.cpp`}, `#include.*".*"|".*"/\*cbt-require\*/`)
	if !regexp.MustCompile(`main\(\)`).Match([]byte(main)) {
		t.Fatalf("main src fail:\n%v", main)
	}
	if len(src) != 1 {
		t.Fatalf("sub src too many:\n%v", len(src))
	} else if src[0] != `func.cpp` {
		t.Fatalf("sub src fail:\n%v", src[0])
	}
	if !regexp.MustCompile(`void func`).Match([]byte(headers[`func.cpp`])) {
		t.Fatalf("src open fail: %v", headers[`func.cpp`])
	}
	delete(headers, "func.cpp")
	if !regexp.MustCompile(`void func\(\);`).Match([]byte(headers[`func.hpp`])) {
		t.Fatalf("include open fail: %v", headers[`func.hpp`])
	}
	delete(headers, "func.hpp")
	for key, val := range headers {
		b, err := ioutil.ReadFile(`test_sample/` + key)
		if err != nil {
			t.Fatalf("include open error: %v", err)
		}
		if !regexp.MustCompile(`#define ` + key).Match(b) {
			t.Fatalf("include expect fail:\n%v", val)
		}
	}
}
func TestExpandIncludeAll(t *testing.T) {
	all := expand.ExpandAll([]string{`test_sample/main.cpp`, `test_sample/func.cpp`}, `#include.*".*"|".*"/\*cbt-require\*/`)

	if !regexp.MustCompile(`main\(\)`).Match([]byte(all[`main.cpp`])) {
		t.Fatalf("include open fail: %v", all[`main.cpp`])
	}
	delete(all, "main.cpp")
	if !regexp.MustCompile(`void func\(\)`).Match([]byte(all[`func.cpp`])) {
		t.Fatalf("include open fail: %v", all[`func.cpp`])
	}
	delete(all, "func.cpp")
	if !regexp.MustCompile(`void func\(\);`).Match([]byte(all[`func.hpp`])) {
		t.Fatalf("include open fail: %v", all[`func.hpp`])
	}
	delete(all, "func.hpp")
	for key, val := range all {
		b, err := ioutil.ReadFile(`test_sample/` + key)
		if err != nil {
			t.Fatalf("include open error: %v", err)
		}
		if !regexp.MustCompile(`#define ` + key).Match(b) {
			t.Fatalf("include expect fail:\n%v", val)
		}
	}
}
