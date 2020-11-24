# cloudevents-go-fn

Playground experimenting with Go functions for CloudEvents

# Build this buildpack

This buildpack can be built (from the root of the repo) with:

```shell
pack package-buildpack my-buildpack --config ./package.toml
```

# Use this buildpack

```shell
pack build blah --buildpack ghcr.io/mattmoor/cloudevents-go-fn:main
```

# Sample function

With this buildpack, users can define a trivial Go function that supports one of the supported cloudevent/sdk-go signatures.  For example, the following function:

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

You can configure both the package containing the function and the name of
the function via the following configuration options in `project.toml`:

```toml
[[build.env]]
name = "CE_GO_PACKAGE"
value = "./blah"

[[build.env]]
name = "CE_GO_FUNCTION"
value = "MyCustomHandlerName"
```
