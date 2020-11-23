package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vaikas/buildpackstuff/pkg/detect"
)

const supportedFuncs = `
Could not find a supported function signature. Examples of supported functions are
shown below, also showing the imports that you can use. The function must also be visible
outside of the package (capitalized, for example, Receive vs. receive).

import (
    "context"
    event "github.com/cloudevents/sdk-go/v2"
    "github.com/cloudevents/sdk-go/v2/protocol"
)

The following function signatures are supported by this builder:
func(event.Event)
func(event.Event) protocol.Result
func(event.Event) error
func(context.Context, event.Event)
func(context.Context, event.Event) protocol.Result
func(context.Context, event.Event) error
func(event.Event) *event.Event
func(event.Event) (*event.Event, protocol.Result)
func(event.Event) (*event.Event, error)
func(context.Context, event.Event) *event.Event
func(context.Context, event.Event) (*event.Event, protocol.Result)
func(context.Context, event.Event) (*event.Event, error)
`

const planFileFormat = `
[[provides]]
name = "ce-go-function"
[[requires]]
name = "ce-go-function"
[requires.metadata]
package = "PACKAGE"
function = "CE_GO_FUNCTION"
protocol = "CE_PROTOCOL"
`

func printSupportedFunctionsAndExit() {
	fmt.Println(supportedFuncs)
	os.Exit(100)
}

func main() {
	log.Println("ARGS: ", os.Args)
	for _, e := range os.Environ() {
		log.Println(e)
	}

	if len(os.Args) < 3 {
		log.Printf("Usage: %s <PLATFORM_DIR> <BUILD_PLAN>\n", os.Args[0])
		os.Exit(100)
	}

	moduleName, err := readModuleName()
	if err != nil {
		log.Println("Failed to read go.mod file: ", err)
		os.Exit(100)
	}

	// There are two ENV variables that control what should be checked.
	// We yank the base package from go.mod and append CE_GO_PACKAGE into it
	// if it's given.
	// TODO: Use library for these...
	goPackage := os.Getenv("CE_GO_PACKAGE")
	if goPackage == "" {
		goPackage = "./"
	}
	if !strings.HasSuffix(goPackage, "/") {
		goPackage = goPackage + "/"
	}
	fullGoPackage := moduleName
	if goPackage != "./" {
		fullGoPackage = fullGoPackage + "/" + filepath.Clean(goPackage)
	}
	log.Println("Using relative path to look for function: ", goPackage)

	goFunction := os.Getenv("CE_GO_FUNCTION")
	goProtocol := os.Getenv("CE_PROTOCOL")
	if goProtocol == "" {
		goProtocol = "http"
	}

	planFileName := os.Args[2]
	log.Println("using plan file: ", planFileName)

	// read all go files from the directory that was given. Note that if no directory (CE_PACKAGE)
	// was given, this is ./
	files, err := filepath.Glob(fmt.Sprintf("%s*.go", goPackage))
	if err != nil {
		log.Printf("failed to read directory %s : %s\n", goPackage, err)
		printSupportedFunctionsAndExit()
	}

	for _, f := range files {
		log.Printf("Processing file %s\n", f)
		// read file
		file, err := os.Open(f)
		if err != nil {
			log.Println(err)
			printSupportedFunctionsAndExit()
		}
		defer file.Close()

		// read the whole file in
		srcbuf, err := ioutil.ReadAll(file)
		if err != nil {
			log.Println(err)
			printSupportedFunctionsAndExit()
		}
		f := &detect.Function{File: f, Source: string(srcbuf)}
		if deets := detect.CheckFile(f); deets != nil {
			log.Printf("Found supported function %q in package %q signature %q", deets.Name, deets.Package, deets.Signature)
			// If the user didn't specify a specific function, use it. If they specified the function, make sure it
			// matches what we found.
			if goFunction == "" || goFunction == deets.Name {
				deets.Package = fullGoPackage
				if err := writePlan(planFileName, goProtocol, deets); err != nil {
					log.Println("failed to write the build plan: ", err)
				}
				os.Exit(0)
			}
		}
	}
	printSupportedFunctionsAndExit()
}

// writePlan writes the planFileName with the following format:
//[[provides]]
//name = "ce-go-function"
//[[requires]]
//name = "ce-go-function"
//[requires.metadata]
//package = <details.packageName>
//function = "details.Name"
//protocol = "http"
func writePlan(planFileName, protocol string, details *detect.FunctionDetails) error {
	planFile, err := os.OpenFile(planFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("failed to open the plan file for writing", os.Args[2], err)
		printSupportedFunctionsAndExit()
	}
	defer planFile.Close()

	// Replace the placeholders with valid values
	replacedPlan := strings.Replace(string(planFileFormat), "PACKAGE", details.Package, 1)
	replacedPlan = strings.Replace(replacedPlan, "CE_GO_FUNCTION", details.Name, 1)
	replacedPlan = strings.Replace(replacedPlan, "CE_PROTOCOL", protocol, 1)
	if _, err := planFile.WriteString(replacedPlan); err != nil {
		printSupportedFunctionsAndExit()
	}
	return nil
}

// readModuleName is a terrible hack for yanking the module from go.mod file.
// Should be replaced with something that actually understands go...
func readModuleName() (string, error) {
	modFile, err := os.Open("./go.mod")
	if err != nil {
		return "", err
	}
	defer modFile.Close()
	scanner := bufio.NewScanner(modFile)
	for scanner.Scan() {
		pieces := strings.Split(scanner.Text(), " ")
		fmt.Printf("Found pieces as %+v\n", pieces)
		if len(pieces) >= 2 && pieces[0] == "module" {
			return pieces[1], nil
		}
	}
	return "", nil
}
