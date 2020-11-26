package function

import (
	"testing"

	"github.com/paketo-buildpacks/packit"
)

func TestDetect(t *testing.T) {
	const goodWD = "../../" // where our go.mod file lives

	tests := []struct {
		name  string
		wd    string
		pkg   string
		fn    string
		proto string
		match bool
	}{{
		name:  "default function",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/default",
		fn:    "Receiver",
		proto: "http",
		match: true,
	}, {
		name:  "non-default function (no override)",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/nondefault",
		fn:    "Receiver",
		proto: "http",
		match: false,
	}, {
		name:  "non-default function (correct override)",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/nondefault",
		fn:    "MyCustomReceiver",
		proto: "http",
		match: true,
	}, {
		name:  "bad signature function",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/nomatch",
		fn:    "Receiver",
		proto: "http",
		match: false,
	}, {
		name:  "no functions",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata",
		fn:    "Receiver",
		proto: "http",
		match: false,
	}, {
		name:  "no go.mod",
		wd:    ".",
		pkg:   "./testdata/default",
		fn:    "Receiver",
		proto: "http",
		match: false,
	}, {
		name:  "unsupported protocol",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/default",
		fn:    "Receiver",
		proto: "matt",
		match: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := Detector{
				Package:  test.pkg,
				Function: test.fn,
				Protocol: test.proto,
			}
			p, err := d.Detect(packit.DetectContext{
				WorkingDir: test.wd,
			})
			if err != nil && test.match {
				t.Fatal("Unexpected error:", err)
			} else if err == nil && !test.match {
				t.Fatal("Unexpected match:", p)
			}
		})
	}
}
