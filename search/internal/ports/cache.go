package ports

import (
	"context"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
)

type CacheReader interface {
	Get(ctx context.Context, key string) ([]product.Product, error)
}

type CacheWriter interface {
	Set(ctx context.Context, key string, products []product.Product, tags ...string) error
}

type CacheWriteReader interface {
	CacheReader
	CacheWriter
}

type CacheInvalidator interface {
	InvalidateByKey(ctx context.Context, key string) error
	InvalidateByTags(ctx context.Context, tags []string) error
	InvalidateTagsByPattern(ctx context.Context, pattern string) error
}

type Cacher interface {
	CacheWriteReader
	CacheInvalidator
}
