package server

import (
	"encoding/json"
	"testing"
)

func BenchmarkJSONString(t *testing.B) {
	var s = "the need for mead"
	for i := 0; i < t.N; i++ {
		jsonString(s)
	}
}

func BenchmarkJSONMarshal(t *testing.B) {
	var s = "the need for mead"
	for i := 0; i < t.N; i++ {
		json.Marshal(s)
	}
}

func TestIsJsonNumber(t *testing.T) {
	test := func(expected bool, val string) {
		actual := isJSONNumber(val)
		if expected != actual {
			t.Fatalf("Expected %t == isJsonNumber(\"%s\") but was %t", expected, val, actual)
		}
	}
	test(false, "")
	test(false, "-")
	test(false, "foo")
	test(false, "0123")
	test(false, "1.")
	test(false, "1.0e")
	test(false, "1.0e-")
	test(false, "1.0E10NaN")
	test(false, "1.0ENaN")
	test(true, "-1")
	test(true, "0")
	test(true, "0.0")
	test(true, "42")
	test(true, "1.0E10")
	test(true, "1.0e10")
	test(true, "1E+5")
	test(true, "1E-10")
}
