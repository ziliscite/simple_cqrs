package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
	"github.com/ziliscite/cqrs_search/internal/ports"
)

type CreateProduct struct {
	Name     string
	Category string
	Price    float64
}

func (p CreateProduct) Validate() Errs {
	errs := make(map[string]error)
	if p.Name == "" {
		errs["name"] = errors.New("product name is required")
	}

	if p.Category == "" {
		errs["category"] = errors.New("product category is required")
	}

	if p.Price <= 0 {
		errs["price"] = errors.New("product price must be greater than zero")
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type CreateProductHandler interface {
	Handle(ctx context.Context, cmd CreateProduct) error
}

type createProductHandler struct {
	repo ports.WriteRepository
	ch   ports.CacheInvalidator
}

func NewCreateProductHandler(repo ports.WriteRepository, ch ports.CacheInvalidator) CreateProductHandler {
	return &createProductHandler{repo: repo, ch: ch}
}

func (h *createProductHandler) Handle(ctx context.Context, cmd CreateProduct) error {
	p, err := product.New(cmd.Name, cmd.Category, cmd.Price)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

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
