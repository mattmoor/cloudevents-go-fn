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
	"time"

        p "{{.Package}}"
)

func main() {
	// When we get a SIGTERM, cancel the first context and start to fail
	// readiness probes.  After a suitable grace period for the container
	// runtime to redirect network traffic (due to the failing probes),
	// cancel the second context so we drain outstanding requests and exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	go func() {
		<-c
		cancel()
		time.Sleep(30 * time.Second)
		cancel2()
	}()

	client, err := newClient(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := client.StartReceiver(ctx2, p.{{.Function}}); err != nil {
		log.Fatal(err)
	}
}
`

const protocolHTTP = `
// +build http

package main

import (
	"context"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
)

func probe(ctx context.Context) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		// If we get requests from the kubelet's prober logic, handle it so
		// the user function doesn't have to.
		if strings.HasPrefix(r.Header.Get("User-Agent"), "kube-probe/") {
			select {
				// If we've received the termination signal, then start
				// to fail readiness probes to drain traffic away from
				// this replica.
				case <-ctx.Done():
					http.Error(w, "shutting down", http.StatusServiceUnavailable)

				// Otherwise, start to pass readiness probes as soon as
				// we are able to invoke the user function.
				default:
					w.WriteHeader(http.StatusOK)
			}
		} else {
			// If there is no kubelet probe header, then don't accept GET requests.
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func newClient(ctx context.Context) (cloudevents.Client, error) {
	p, err := cehttp.New(cehttp.WithGetHandlerFunc(probe(ctx)))
	if err != nil {
		return nil, err
	}
	return ceclient.NewObserved(p, ceclient.WithTimeNow(), ceclient.WithUUIDs())
}
`

var templates = map[string]*template.Template{
	"main": template.Must(template.New("ce-go-function-main").Parse(packageMain)),
	"http": template.Must(template.New("ce-go-function-main").Parse(protocolHTTP)),
}
