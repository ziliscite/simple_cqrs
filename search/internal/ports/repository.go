package ports

import (
	"context"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
)

type ReadRepository interface {
	Search(ctx context.Context, opts *product.Search) ([]product.Product, error)
	GetByID(ctx context.Context, id string) (*product.Product, error)
}

type WriteRepository interface {
	Create(ctx context.Context, product *product.Product) error
	Update(ctx context.Context, product *product.Product) error
	Delete(ctx context.Context, id string) error
}

type Repository interface {
	ReadRepository
	WriteRepository
}
