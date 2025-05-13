package command

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"
)

type DeleteProduct struct {
	ID product.ID
}

func NewDeleteProduct(id string) (DeleteProduct, error) {
	var dp DeleteProduct

	if id == "" {
		return dp, errors.New("id is required")
	}

	return DeleteProduct{
		ID: product.ID(id),
	}, nil
}

type DeleteProductRequest struct {
	ID string `json:"id"`
}

type DeleteProductHandler interface {
	Handle(ctx context.Context, cmd DeleteProduct) error
}

type deleteProductHandler struct {
	repo ports.Repository
	pub  ports.Publisher
}

func NewDeleteProductHandler(repo ports.Repository, producer ports.Publisher) DeleteProductHandler {
	return &deleteProductHandler{repo: repo, pub: producer}
}

func (h *deleteProductHandler) Handle(ctx context.Context, cmd DeleteProduct) error {
	if err := h.repo.Delete(ctx, cmd.ID.String()); err != nil {
		return err
	}

	msg, err := json.Marshal(DeleteProductRequest{
		ID: cmd.ID.String(),
	})
	if err != nil {
		return err
	}

	return h.pub.Publish(ctx, msg, "delete")
}
