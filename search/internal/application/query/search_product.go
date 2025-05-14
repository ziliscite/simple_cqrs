package query

import (
	"context"
	"fmt"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
	"github.com/ziliscite/cqrs_search/internal/ports"
	"log"
)

type SearchProduct struct {
	search *product.Search
}

func NewSearchProduct(search *product.Search) SearchProduct {
	return SearchProduct{
		search: search,
	}
}

type SearchProductHandler interface {
	Handle(ctx context.Context, query SearchProduct) ([]product.Product, error)
}

type searchProductHandler struct {
	repo ports.ReadRepository
	ch   ports.CacheWriteReader
}

func NewSearchProductHandler(repo ports.ReadRepository, cacher ports.CacheWriteReader) SearchProductHandler {
	return &searchProductHandler{
		repo: repo,
		ch:   cacher,
	}
}

func (h *searchProductHandler) Handle(ctx context.Context, query SearchProduct) ([]product.Product, error) {
	key, tags := query.search.Key()

	// check cache
	cached, err := h.ch.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached products: %w", err)
	}

	// if hit, return
	if cached != nil && len(cached) > 0 {
		log.Printf("hit cache: %s", key)
		return cached, nil
	}

	// cache miss, search
	log.Printf("miss cache: %s", key)
	products, err := h.repo.Search(ctx, query.search)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// add more tags
	for _, p := range products {
		log.Printf("product: %s %s %s $%.2f", p.ID(), p.Name(), p.Category(), p.Price())
		tags = append(tags, fmt.Sprintf("tag:product:%s", p.ID()))
	}

	// set cache
	if err = h.ch.Set(ctx, key, products, tags...); err != nil {
		return nil, fmt.Errorf("failed to cache products: %w", err)
	}

	return products, nil
}
