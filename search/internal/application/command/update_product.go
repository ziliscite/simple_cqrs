package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
	"github.com/ziliscite/cqrs_search/internal/ports"
)

type UpdateProduct struct {
	Product product.Product
}

func (cmd UpdateProduct) Validate() Errs {
	errs := make(map[string]error)
	if cmd.Product.ID() == "" {
		errs["id"] = errors.New("product id is required")
	}

	if cmd.Product.Name() == "" {
		errs["name"] = errors.New("product name is required")
	}

	if cmd.Product.Category() == "" {
		errs["category"] = errors.New("product category is required")
	}

	if cmd.Product.Price() == 0 {
		errs["price"] = errors.New("product price is required")
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateProductHandler interface {
	Handle(ctx context.Context, cmd UpdateProduct) error
}

type updateProductHandler struct {
	repo ports.WriteRepository
	ch   ports.CacheInvalidator
}

func NewUpdateProductHandler(repo ports.WriteRepository, cache ports.CacheInvalidator) UpdateProductHandler {
	return &updateProductHandler{
		repo: repo,
		ch:   cache,
	}
}

func (h *updateProductHandler) Handle(ctx context.Context, cmd UpdateProduct) error {
	if err := h.ch.InvalidateByKey(ctx, fmt.Sprintf("product:%s", cmd.Product.ID())); err != nil {
		return err
	}

	return h.repo.Update(ctx, &cmd.Product)
}
