package command

import (
	"context"
	"errors"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"
)

type DeleteProduct struct {
	ID product.ID
}

func (cmd DeleteProduct) Validate() error {
	if cmd.ID == "" {
		return errors.New("id is required")
	}

	return nil
}

type DeleteProductHandler interface {
	Handle(ctx context.Context, cmd DeleteProduct) error
}

type deleteProductHandler struct {
	repo     ports.Repository
	producer ports.Publisher[product.ID]
}

func NewDeleteProductHandler(repo ports.Repository, producer ports.Publisher[product.ID]) DeleteProductHandler {
	return &deleteProductHandler{repo: repo, producer: producer}
}

func (h *deleteProductHandler) Handle(ctx context.Context, cmd DeleteProduct) error {
	e := product.NewEvent(product.Delete, cmd.ID)
	if err := h.producer.Publish(ctx, e); err != nil {
		return err
	}

	return h.repo.Delete(ctx, cmd.ID.String())
}
