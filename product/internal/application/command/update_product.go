package command

import (
	"context"
	"errors"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"
)

type UpdateProduct struct {
	Product product.Product
}

func (cmd UpdateProduct) Validate() map[string]error {
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
	repo     ports.Repository
	producer ports.Publisher[*product.Product]
}

func NewUpdateProductHandler(repo ports.Repository, producer ports.Publisher[*product.Product]) UpdateProductHandler {
	return &updateProductHandler{repo: repo, producer: producer}
}

func (h *updateProductHandler) Handle(ctx context.Context, cmd UpdateProduct) error {
	e := product.NewEvent(product.Update, &cmd.Product)
	if err := h.producer.Publish(ctx, e); err != nil {
		return err
	}

	return h.repo.Update(ctx, &cmd.Product)
}
