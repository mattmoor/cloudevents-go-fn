package function

import "text/template"

const packageMain = `
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

        p "{{.Package}}"
)

func main() {
	// When we get a SIGTERM, cancel the context.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		cancel()
	}()

	client, err := newClient(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := client.StartReceiver(ctx, p.{{.Function}}); err != nil {
		log.Fatal(err)
	}
}
`

const protocolHTTP = `
// +build http

package main

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func newClient(ctx context.Context) (cloudevents.Client, error) {
	return cloudevents.NewDefaultClient()
}
`

var templates = map[string]*template.Template{
	"main": template.Must(template.New("ce-go-function-main").Parse(packageMain)),
	"http": template.Must(template.New("ce-go-function-main").Parse(protocolHTTP)),
}
