package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/ziliscite/cqrs_search/internal/ports"
)

type DeleteProduct struct {
	ID string
}

func NewDeleteProduct(id string) (DeleteProduct, error) {
	var dp DeleteProduct
	if id == "" {
		return dp, errors.New("id is required")
	}

	dp.ID = id
	return dp, nil
}

type DeleteProductHandler interface {
	Handle(ctx context.Context, cmd DeleteProduct) error
}

type deleteProductHandler struct {
	repo ports.WriteRepository
	ch   ports.CacheInvalidator
}

func NewDeleteProductHandler(repo ports.WriteRepository, cache ports.CacheInvalidator) DeleteProductHandler {
	return &deleteProductHandler{repo: repo, ch: cache}
}

func (h *deleteProductHandler) Handle(ctx context.Context, cmd DeleteProduct) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		return err
	}

	if err := h.ch.InvalidateByKey(ctx, fmt.Sprintf("product:%s", cmd.ID)); err != nil {
		return err
	}

	return nil
}
