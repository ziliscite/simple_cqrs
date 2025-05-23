package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/ziliscite/cqrs_search/internal/domain/product"

	"github.com/ziliscite/cqrs_search/internal/ports"
)

type UpdateProductEvent struct {
	ID       string
	Name     string
	Category string
	Price    float64
}

func NewUpdateProduct(id, name, category string, price float64) (UpdateProductEvent, Errs) {
	var up UpdateProductEvent

	errs := make(map[string]error)
	if id == "" {
		errs["id"] = errors.New("product id is required")
	}

	if name == "" {
		errs["name"] = errors.New("product name is required")
	}

	if category == "" {
		errs["category"] = errors.New("product category is required")
	}

	if price <= 0 {
		errs["price"] = errors.New("product price must be greater than zero")
	}

	if len(errs) > 0 {
		return up, errs
	}

	up.ID = id
	up.Name = name
	up.Category = category
	up.Price = price

	return up, nil
}

type UpdateProductHandler interface {
	Handle(ctx context.Context, cmd UpdateProductEvent) error
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

func (h *updateProductHandler) Handle(ctx context.Context, cmd UpdateProductEvent) error {
	p, err := product.New(cmd.Name, cmd.Category, cmd.Price)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	p.SetID(cmd.ID)

	if err = h.repo.Update(ctx, p); err != nil {
		return err
	}

	if err = h.ch.InvalidateByKey(ctx, fmt.Sprintf("product:%s", p.ID())); err != nil {
		return err
	}

	return nil
}
