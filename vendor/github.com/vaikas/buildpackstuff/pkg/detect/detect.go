package detect

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	//	"unicode"
)

type Function struct {
	File   string // full path to the file
	Source string // full source file
}

const (
	ceImport         = `"github.com/cloudevents/sdk-go/v2"`
	ceProtocolImport = `"github.com/cloudevents/sdk-go/v2/protocol"`
	contextImport    = `"context"`
)

type paramType int

const (
	notSuportedType paramType = iota
	contextType
	eventType
	ptrEventType
	protocolResultType
	errorType
)

type functionSignature struct {
	in  []paramType
	out []paramType
}

// Valid function signatures are like so (defined in: github.com/cloudevents/sdk-go/v2/client/receiver.go):
// * func()                                  *** NOT SUPPORTED ***
// * func() error                            *** NOT SUPPORTED ***
// * func(context.Context)                   *** NOT SUPPORTED ***
// * func(context.Context) protocol.Result   *** NOT SUPPORTED ***
// * func(event.Event)
// * func(event.Event) protocol.Result
// * func(context.Context, event.Event)
// * func(context.Context, event.Event) protocol.Result
// * func(event.Event) *event.Event
// * func(event.Event) (*event.Event, protocol.Result)
// * func(context.Context, event.Event) *event.Event
// * func(context.Context, event.Event) (*event.Event, protocol.Result)
// * func(context.Context, event.Event) (*event.Event, error)
var validFunctions = map[string]functionSignature{
	//	"func(context.Context)":                                              functionSignature{in: []paramType{contextType}},
	//	"func(context.Context) protocol.Result":                              functionSignature{in: []paramType{contextType}, out: []paramType{protocolResultType}},
	"func(event.Event)":                                                  functionSignature{in: []paramType{eventType}},
	"func(event.Event) protocol.Result":                                  functionSignature{in: []paramType{eventType}, out: []paramType{protocolResultType}},
	"func(event.Event) error":                                            functionSignature{in: []paramType{eventType}, out: []paramType{errorType}},
	"func(context.Context, event.Event)":                                 functionSignature{in: []paramType{contextType, eventType}},
	"func(context.Context, event.Event) protocol.Result":                 functionSignature{in: []paramType{contextType, eventType}, out: []paramType{protocolResultType}},
	"func(context.Context, event.Event) error":                           functionSignature{in: []paramType{contextType, eventType}, out: []paramType{errorType}},
	"func(event.Event) *event.Event":                                     functionSignature{in: []paramType{eventType}, out: []paramType{ptrEventType}},
	"func(event.Event) (*event.Event, protocol.Result)":                  functionSignature{in: []paramType{eventType}, out: []paramType{ptrEventType, protocolResultType}},
	"func(event.Event) (*event.Event, error)":                            functionSignature{in: []paramType{eventType}, out: []paramType{ptrEventType, errorType}},
	"func(context.Context, event.Event) *event.Event":                    functionSignature{in: []paramType{contextType, eventType}, out: []paramType{ptrEventType}},
	"func(context.Context, event.Event) (*event.Event, protocol.Result)": functionSignature{in: []paramType{contextType, eventType}, out: []paramType{ptrEventType, protocolResultType}},
	"func(context.Context, event.Event) (*event.Event, error)":           functionSignature{in: []paramType{contextType, eventType}, out: []paramType{ptrEventType, errorType}},
}

// imports keeps track of which files that we care about are imported as which
// local imports. For example if you import:
// 	cloudevents "github.com/cloudevents/sdk-go/v2"
// localCEImport would be set to cloudevents
type imports struct {
	localCEImport       string
	localProtocolImport string
	localContextImport  string
}

type FunctionDetails struct {
	Name string
	Package   string
	Signature string
}

func ReadAndCheckFile(filename string) *FunctionDetails {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer file.Close()

	// read the whole file in
	srcbuf, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return nil
	}
	return CheckFile(&Function{File: filename, Source: string(srcbuf)})
}

func CheckFile(f *Function) *FunctionDetails {
	// file set
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, f.File, f.Source, 0)
	if err != nil {
		log.Println(err)
		return nil
	}

	c := imports{}
	var functionName = ""
	var signature = ""

	// main inspection
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch fn := n.(type) {
		// Check if the file imports the cloud events SDK that we're expecting
		case *ast.File:
			for _, i := range fn.Imports {
				if i.Path.Value == ceImport {
					if i.Name == nil {
						c.localCEImport = "v2"
					} else {
						c.localCEImport = i.Name.String()
					}
				}
				if i.Path.Value == ceProtocolImport {
					if i.Name == nil {
						c.localProtocolImport = "protocol"
					} else {
						c.localProtocolImport = i.Name.String()
					}
				}
				if i.Path.Value == contextImport {
					if i.Name == nil {
						c.localContextImport = "context"
					} else {
						c.localContextImport = i.Name.String()
					}
				}
			}

			for _, d := range fn.Decls {
				if f, ok := d.(*ast.FuncDecl); ok {
					functionName = f.Name.Name
					if f.Recv != nil {
						fmt.Println("Found receiver ", f.Recv)
					}
					if sig := checkFunction(c, f.Type); sig != "" {
						signature = sig
					}
				}
			}
		}
		return true
	})
	if signature != "" && functionName != "" {
		return &FunctionDetails{Name: functionName, Signature: signature}
	}
	return nil
}

// checkFunction takes a function signature and returns a friendly (string)
// representation of the supported function or "" if the function signature
// is not supported.
// For example
// func Receive(ctx ctx.Context, event cloudevents.Event) (*cloudevents.Event, protocol.Result) {
// would return:
// func(context.Context, event.Event) (*event.Event, protocol.Result)
func checkFunction(c imports, f *ast.FuncType) string {
	fs := functionSignature{}
	if f == nil {
		return ""
	}
	if f.Params != nil {
		for _, p := range f.Params.List {
			t := typeToParamType(c, p.Type)
			fs.in = append(fs.in, t)
		}
	}
	if f.Results != nil {
		for _, r := range f.Results.List {
			t := typeToParamType(c, r.Type)
			fs.out = append(fs.out, t)
		}
	}

	for k, v := range validFunctions {
		if len(fs.in) == len(v.in) && len(fs.out) == len(v.out) {
			match := true
			for i := range fs.in {
				if fs.in[i] != v.in[i] {
					match = false
					continue
				}
			}
			for i := range fs.out {
				if fs.out[i] != v.out[i] {
					match = false
					continue
				}
			}
			if match {
				return k
			}
		}
	}
	return ""
}

// typeToParamType will take import paths and an expression and try to map
// it to know paramType. If supported paramType is not found, will return
// notSupportedType
// TODO: Consider supporting mapping error. I just don't know how to and those
// functions do not really seem all that important to support.
func typeToParamType(c imports, e ast.Expr) paramType {
	switch e := e.(type) {
	// Check if pointer to Event
	case *ast.StarExpr:
		if s, ok := e.X.(*ast.SelectorExpr); ok {
			// We only support pointer to Event
			if s.Sel.String() == "Event" {
				if im, ok := s.X.(*ast.Ident); ok {
					if im.Name == c.localCEImport {
						return ptrEventType
					}
				}
			}
		}
	case *ast.SelectorExpr:
		if e.Sel.String() == "Event" {
			if im, ok := e.X.(*ast.Ident); ok {
				if c.localCEImport != "" && im.Name == c.localCEImport {
					return eventType
				}
			}
		}
		if e.Sel.String() == "Context" {
			if im, ok := e.X.(*ast.Ident); ok {
				if c.localContextImport != "" && im.Name == "" || im.Name == c.localContextImport {
					return contextType
				}
			}
		}
		if e.Sel.String() == "Result" {
			if im, ok := e.X.(*ast.Ident); ok {
				if c.localProtocolImport != "" && im.Name == c.localProtocolImport {
					return protocolResultType
				}
			}
		}
	case *ast.Ident:
		if e.Name == "error" {
			return errorType
		}
	}
	return notSuportedType
}
