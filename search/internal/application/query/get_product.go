package query

import (
	"context"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
	"github.com/ziliscite/cqrs_search/internal/ports"
)

type GetProduct struct {
	ID string
}

func NewGetProduct(id string) GetProduct {
	return GetProduct{
		ID: id,
	}
}

type GetProductHandler interface {
	Handle(ctx context.Context, query GetProduct) (*product.Product, error)
}

type getProductHandler struct {
	repo ports.ReadRepository
}

func NewGetProductHandler(repo ports.ReadRepository) GetProductHandler {
	return &getProductHandler{
		repo: repo,
	}
}

func (h *getProductHandler) Handle(ctx context.Context, query GetProduct) (*product.Product, error) {
	return h.repo.GetByID(ctx, query.ID)
}
