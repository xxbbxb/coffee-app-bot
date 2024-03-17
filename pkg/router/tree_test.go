package router

import (
	"testing"
)

// TestParsePrefix calls ParsePrefix checking
// for a valid return value.
func TestParsePrefix(t *testing.T) {
	input := "/a/2022-12-15/def"
	pattern := "/a/{year:[0-9]+}-{month}-{day}/"

	if PrefixParse(pattern, input) == nil {
		t.Fatalf(`PrefixParse("%s", "%s") failed\n`, pattern, input)
	}
	if PrefixParse("/a", "/a") == nil {
		t.Fatalf(`PrefixParse("/a", "/a") failed\n`)
	}
	if PrefixParse("/a", "/a/asda/asd") == nil {
		t.Fatalf(`PrefixParse("/a", "/a/asda/asd") failed\n`)
	}
	if PrefixParse("/a", "/b") != nil {
		t.Fatalf(`PrefixParse("/a", "/b") failed\n`)
	}
	if PrefixParse("/{year}-{month}", "/2023") != nil {
		t.Fatalf(`PrefixParse("/{year}-{month}", "/2023") failed\n`)
	}
}
