package ports

import (
	"context"
)

type Publisher interface {
	Publish(ctx context.Context, payload []byte, event string) error
}
