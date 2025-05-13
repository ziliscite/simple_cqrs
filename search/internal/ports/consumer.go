package ports

import "context"

type Consumer interface {
	Consume(ctx context.Context, queue string) error
	CreateProduct(ctx context.Context, payload []byte) error
	UpdateProduct(ctx context.Context, payload []byte) error
	DeleteProduct(ctx context.Context, payload []byte) error
}
