package ports

import (
	"context"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
)

type Publisher[M product.Message] interface {
	Publish(ctx context.Context, event product.Event[M]) error
}
