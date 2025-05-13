package ports

import (
	"context"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
)

type Repository interface {
	Create(ctx context.Context, product *product.Product) error
	Update(ctx context.Context, product *product.Product) error
	Delete(ctx context.Context, id string) error
}
