package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
	"github.com/ziliscite/cqrs_search/internal/ports"
	"log"
)

type CreateProductEvent struct {
	ID       string
	Name     string
	Category string
	Price    float64
}

func NewCreateProduct(id, name, category string, price float64) (CreateProductEvent, Errs) {
	var cp CreateProductEvent

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
		return cp, errs
	}

	cp.ID = id
	cp.Name = name
	cp.Category = category
	cp.Price = price

	return cp, nil
}

type CreateProductHandler interface {
	Handle(ctx context.Context, cmd CreateProductEvent) error
}

type createProductHandler struct {
	repo ports.WriteRepository
	ch   ports.CacheInvalidator
}

func NewCreateProductHandler(repo ports.WriteRepository, ch ports.CacheInvalidator) CreateProductHandler {
	return &createProductHandler{repo: repo, ch: ch}
}

func (h *createProductHandler) Handle(ctx context.Context, cmd CreateProductEvent) error {
	p, err := product.New(cmd.Name, cmd.Category, cmd.Price)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	p.SetID(cmd.ID)

	log.Printf("product: %s %s %s $%.2f", p.ID(), p.Name(), p.Category(), p.Price())
	if err = h.repo.Create(ctx, p); err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Invalidate all search-result caches
	if err = h.ch.InvalidateByKey(ctx, "products:search"); err != nil {
		return fmt.Errorf("failed to invalidate search cache: %w", err)
	}

	// Invalidate all tag caches that may be affected by this product
	tags := p.Tags()
	if err = h.ch.InvalidateByTags(ctx, tags); err != nil {
		return err
	}

	// Invalidate all paging, sort, min, max caches
	patterns := []string{"tag:paging:*", "tag:sort:*", "tag:min:*", "tag:max:*"}
	for _, tag := range patterns {
		if err = h.ch.InvalidateTagsByPattern(ctx, tag); err != nil {
			return err
		}
	}

	return nil
}
