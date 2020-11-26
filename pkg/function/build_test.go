package function

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func TestBuild(t *testing.T) {
	const (
		layersPath = "/layers"
		// These don't need to be real, we just need to be able to check them.
		pkg   = "paketo.io/my-fn"
		fn    = "MyHandler"
		proto = "http"
	)
	wantBuildPlan := packit.BuildResult{
		Layers: []packit.Layer{{
			Name:  "ce-go-function-cmd",
			Path:  filepath.Join(layersPath, "ce-go-function-cmd"),
			Build: true,
			BuildEnv: packit.Environment{
				"BP_GO_TARGETS.override": targetPackage,
				"GOFLAGS.append":         " -tags=" + proto,
			},
		}},
	}
	tests := []struct {
		name    string
		plan    packit.BuildpackPlan
		success bool
	}{{
		name: "successful build",
		plan: packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{{
				Name: "unrelated-entry",
			}, {
				Name: "ce-go-function",
				Metadata: map[string]interface{}{
					"package":  pkg,
					"function": fn,
					"protocol": proto,
				},
			}},
		},
		success: true,
	}, {
		name: "unsupported protocol",
		plan: packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{{
				Name: "unrelated-entry",
			}, {
				Name: "ce-go-function",
				Metadata: map[string]interface{}{
					"package":  pkg,
					"function": fn,
					"protocol": "matt",
				},
			}},
		},
		success: false,
	}, {
		name: "missing plan entry",
		plan: packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{{
				Name: "unrelated-entry",
			}},
		},
		success: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := Builder{
				Logger: scribe.NewLogger(ioutil.Discard),
			}
			dir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatal("TempDir() =", err)
			}
			defer os.RemoveAll(dir)

			bp, err := b.Build(packit.BuildContext{
				WorkingDir: dir,
				Layers: packit.Layers{
					Path: layersPath,
				},
				Plan: test.plan,
			})
			if err != nil && test.success {
				t.Fatal("Unexpected error:", err)
			} else if err == nil && !test.success {
				t.Fatal("Unexpected failure:", bp)
			}
			if err != nil {
				return
			}

			// Check that the build plan matches what we want.
			if !cmp.Equal(bp, wantBuildPlan) {
				t.Error("Build (-want, +got): ", cmp.Diff(wantBuildPlan, bp))
			}

			for _, file := range []string{"main", proto} {
				buf := bytes.NewBuffer(nil)
				if err := templates[file].Execute(buf, info{
					Package:  pkg,
					Function: fn,
					Protocol: proto,
				}); err != nil {
					t.Fatalf("templates[%q].Execute() = %v", file, err)
				}
				wantFileContents := buf.String()

				gotFileContents, err := ioutil.ReadFile(filepath.Join(dir, targetPackage, file+".go"))
				if err != nil {
					t.Fatal("ReadFile() =", err)
				}

				if wantFileContents != string(gotFileContents) {
					t.Fatalf("ReadFile() = %q, wanted %q", string(gotFileContents), wantFileContents)
				}
			}
		})
	}
}
