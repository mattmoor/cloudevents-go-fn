package foo

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func MyCustomReceiver(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	return nil, nil
}
