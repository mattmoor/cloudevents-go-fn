package function

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
	"github.com/vaikas/gofunctypechecker/pkg/detect"
)

// Detector is a stateful implementation of packit.DetectFunc to implement
// the detect phase of a Paketo buildpack.  It is expected to be initialized
// from the environment via Kelsey's envconfig library.
type Detector struct {
	// Package holds the name of the user-configured Go package containing
	// a CloudEvents function.
	Package string `envconfig:"CE_GO_PACKAGE" default:"."`

	// Function holds the name of the CloudEvent receiver in Package that this
	// buildpack should wrap in CloudEvents scaffolding.
	Function string `envconfig:"CE_GO_FUNCTION" default:"Receiver"`

	// Protocol holds the name of the protocol to which we will
	// bind the receiver function.
	Protocol string `envconfig:"CE_PROTOCOL" default:"http"`
}

var (
	detectors = map[string]*detect.Detector{
		"http": detect.NewDetector([]detect.FunctionSignature{{
			In: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2/protocol",
				Name:       "Result",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				Name: "error",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "context",
				Name:       "Context",
			}, {
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "context",
				Name:       "Context",
			}, {
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2/protocol",
				Name:       "Result",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "context",
				Name:       "Context",
			}, {
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				Name: "error",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
				Pointer:    true,
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
				Pointer:    true,
			}, {
				ImportPath: "github.com/cloudevents/sdk-go/v2/protocol",
				Name:       "Result",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
				Pointer:    true,
			}, {
				Name: "error",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "context",
				Name:       "Context",
			}, {
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
				Pointer:    true,
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "context",
				Name:       "Context",
			}, {
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
				Pointer:    true,
			}, {
				ImportPath: "github.com/cloudevents/sdk-go/v2/protocol",
				Name:       "Result",
			}},
		}, {
			In: []detect.FunctionArg{{
				ImportPath: "context",
				Name:       "Context",
			}, {
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
			}},
			Out: []detect.FunctionArg{{
				ImportPath: "github.com/cloudevents/sdk-go/v2",
				Name:       "Event",
				Pointer:    true,
			}, {
				Name: "error",
			}},
		}}),
	}
)

// Detect is a member function that implements packit.DetectFunc
func (d *Detector) Detect(dctx packit.DetectContext) (packit.DetectResult, error) {
	moduleName, err := readModuleName(dctx)
	if err != nil {
		return packit.DetectResult{}, err
	}
	pkg := filepath.Join(moduleName, d.Package)
	fn := d.Function

	if detector, ok := detectors[d.Protocol]; !ok {
		return packit.DetectResult{}, fmt.Errorf("unsupported protocol: %q", d.Protocol)
	} else if err := d.checkFunction(dctx, pkg, fn, detector); err != nil {
		return packit.DetectResult{}, err
	}

	return packit.DetectResult{
		Plan: packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{{
				Name: "ce-go-function",
			}},
			Requires: []packit.BuildPlanRequirement{{
				Name: "ce-go-function",
				Metadata: map[string]interface{}{
					"package":  pkg,
					"function": fn,
					"protocol": d.Protocol,
				},
			}},
		},
	}, nil
}

func (d *Detector) checkFunction(dctx packit.DetectContext, pkg, fn string, detector *detect.Detector) error {
	// read all go files from the directory that was given. Note that if no directory (CE_GO_PACKAGE)
	// was given, this is ./
	files, err := filepath.Glob(filepath.Join(dctx.WorkingDir, d.Package, "*.go"))
	if err != nil {
		return err
	}

	for _, f := range files {
		srcbuf, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		deets, err := detector.CheckFile(&detect.Function{
			File:   f,
			Source: string(srcbuf),
		})
		if err != nil {
			return err
		}
		if deets == nil {
			continue
		}

		if deets.Name != fn {
			// TODO(mattmoor): Add help text to tell the user how to properly configure project.toml,
			// or make the defaulting smart when it is not explicitly specified.
			log.Printf("Found supported function %q in package %q signature %q", deets.Name, deets.Package, deets.Signature)
			continue
		}
		return nil
	}

	return fmt.Errorf("unable to find function %q in %q with matching signature", fn, pkg)
}

// readModuleName is a terrible hack for yanking the module from go.mod file.
// Should be replaced with something that actually understands go...
func readModuleName(dctx packit.DetectContext) (string, error) {
	modFile, err := os.Open(filepath.Join(dctx.WorkingDir, "go.mod"))
	if err != nil {
		return "", err
	}
	defer modFile.Close()

	scanner := bufio.NewScanner(modFile)
	for scanner.Scan() {
		pieces := strings.Split(scanner.Text(), " ")
		if len(pieces) >= 2 && pieces[0] == "module" {
			return pieces[1], nil
		}
	}
	return "", nil
}
