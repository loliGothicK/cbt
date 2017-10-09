package wandbox_test

import (
	"testing"

	"github.com/LoliGothick/cbt/internal/wandbox"
)

func TestTransformToMap(t *testing.T) {
	ret := wandbox.TransformToMap([]wandbox.Code{
		{"a", "a"},
		{"b", "b"},
		{"c", "c"},
	})
	for key, val := range ret {
		if key != val {
			t.Errorf("unknown error")
		}
	}
}

func TestTransformToCodes(t *testing.T) {
	ret := wandbox.TransformToCodes(map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
	})
	for _, c := range ret {
		if c.Code != c.FileName {
			t.Errorf("unknown error")
		}
	}
}
