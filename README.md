# cloudevents-go-fn

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/mattmoor/cloudevents-go-fn)
[![Go Report Card](https://goreportcard.com/badge/mattmoor/cloudevents-go-fn)](https://goreportcard.com/report/mattmoor/cloudevents-go-fn)
[![Releases](https://img.shields.io/github/release-pre/mattmoor/cloudevents-go-fn.svg?sort=semver)](https://github.com/mattmoor/cloudevents-go-fn/releases)
[![LICENSE](https://img.shields.io/github/license/mattmoor/cloudevents-go-fn.svg)](https://github.com/mattmoor/cloudevents-go-fn/blob/master/LICENSE)
[![codecov](https://codecov.io/gh/mattmoor/cloudevents-go-fn/branch/master/graph/badge.svg)](https://codecov.io/gh/mattmoor/cloudevents-go-fn)

This repository implements a Go function buildpack for wrapping functions matching one of the
[`cloudevents/sdk-go`](https://github.com/cloudevents/sdk-go) signatures with scaffolding for
the appropriate protocol binding.

This buildpack is not standalone, it should be composed with the Paketo Go buildpacks.

# Build this buildpack

This buildpack can be built (from the root of the repo) with:

```shell
pack package-buildpack my-buildpack --config ./package.toml
```


# Use this buildpack

```shell
# This runs the cloudevents-go-fn buildpack at HEAD within the Paketo Go order.
# You can pin to a release by replacing ":main" below with a release tag
# e.g. ":v0.0.1"
pack build -v test-container \
  --pull-policy if-not-present \
  --buildpack ghcr.io/mattmoor/cloudevents-go-fn:main \
  --buildpack gcr.io/paketo-buildpacks/go:0.2.7
```


# Sample function

With this buildpack, users can define a Go function that implements one of the
supported [`cloudevents/sdk-go`](https://github.com/cloudevents/sdk-go) signatures
For example, the following function:

```go
package fn

import (
       cloudevents "github.com/cloudevents/sdk-go/v2"
)

func Receiver(ce cloudevents.Event) (*cloudevents.Event, error) {
        r := cloudevents.NewEvent(cloudevents.VersionV1)
        r.SetType("io.mattmoor.cloudevents-go-fn")
        r.SetSource("https://github.com/mattmoor/cloudevents-go-fn")

        if err := r.SetData("application/json", struct {
                A string `json:"a"`
                B string `json:"b"`
        }{
                A: "hello",
                B: "world",
        }); err != nil {
                return nil, cloudevents.NewHTTPResult(500, "failed to set response data: %s", err)
        }

        return &r, nil
}
```


# Configuration

You can configure aspects of the generated function scaffolding via the
following configuration options in `project.toml`:

```toml
[[build.env]]
name = "CE_GO_PACKAGE"
value = "./blah"         # default is the root package: "."

[[build.env]]
name = "CE_GO_FUNCTION"
value = "MyReceiverName" # default is "Receiver"

[[build.env]]
name = "CE_PROTOCOL"
value = "http"           # default is "http"
```

Depending on the protocol, you can further customize the behavior of that
protocol at runtime via environment variables prefixed with: `CE_{protocol}`.
