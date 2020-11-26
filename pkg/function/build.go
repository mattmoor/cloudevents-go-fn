package function

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

// Builder is a stateful implementation of packit.BuildFunc to implement
// the build phase of a Paketo buildpack.
type Builder struct {
	// Logger is used to emit waypoints through the build phase of th elifecycle.
	Logger scribe.Logger
}

const targetPackage = "./ce-cmd/function"

// Build is a member function that implements packit.BuildFunc
func (b *Builder) Build(bctx packit.BuildContext) (packit.BuildResult, error) {
	b.Logger.Title("%s %s", bctx.BuildpackInfo.Name, bctx.BuildpackInfo.Version)
	defer b.Logger.Break()

	info, err := b.getInfo(bctx)
	if err != nil {
		return packit.BuildResult{}, err
	}
	b.Logger.Process("Package:  %s", info.Package)
	b.Logger.Process("Function: %s", info.Function)
	b.Logger.Process("Protocol: %s", info.Protocol)

	if err := os.MkdirAll(filepath.Join(bctx.WorkingDir, targetPackage), os.ModePerm); err != nil {
		return packit.BuildResult{}, err
	}

	for _, file := range []string{"main", info.Protocol} {
		err := func() error { // Scope the defer
			p := filepath.Join(bctx.WorkingDir, targetPackage, file+".go")
			mg, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer mg.Close()

			tmpl, ok := templates[file]
			if !ok {
				return fmt.Errorf("unsupported template: %q", file)
			}

			return tmpl.Execute(mg, info)
		}()
		if err != nil {
			return packit.BuildResult{}, err
		}
	}

	return packit.BuildResult{
		Layers: []packit.Layer{{
			Name:  "ce-go-function-cmd",
			Path:  filepath.Join(bctx.Layers.Path, "ce-go-function-cmd"),
			Build: true,
			BuildEnv: packit.Environment{
				"BP_GO_TARGETS.override": targetPackage,
				"GOFLAGS.append":         fmt.Sprint(" -tags=", info.Protocol),
			},
		}},
	}, nil
}

type info struct {
	Package  string
	Function string
	Protocol string
}

func (b *Builder) getInfo(bctx packit.BuildContext) (*info, error) {
	for _, entry := range bctx.Plan.Entries {
		if entry.Name != "ce-go-function" {
			continue
		}
		return &info{
			Package:  entry.Metadata["package"].(string),
			Function: entry.Metadata["function"].(string),
			Protocol: entry.Metadata["protocol"].(string),
		}, nil
	}

	return nil, errors.New("missing metadata for ce-go-function")
}
