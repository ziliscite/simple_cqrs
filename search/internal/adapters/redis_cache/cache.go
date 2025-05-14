package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
	"github.com/ziliscite/cqrs_search/internal/ports"
	"log"
	"strconv"
	"time"
)

func NewRedisClient(user, password, host, port, db string) (*redis.Client, error) {
	d, err := strconv.Atoi(db)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "",
		DB:       d,
	}), nil
}

// cacher bundles read, write, and invalidate methods
// as defined by the user interfaces.
type cacher struct {
	client *redis.Client
	ttl    time.Duration
}

// NewCacher creates a new cacher with given Redis options and default TTL.
func NewCacher(client *redis.Client, timeToLive time.Duration) ports.Cacher {
	return &cacher{
		client: client,
		ttl:    timeToLive,
	}
}

// Get retrieves products from cache by key.
func (c *cacher) Get(ctx context.Context, key string) ([]product.Product, error) {
	raw, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}
	if raw == "" || raw == "{}" || raw == "[]" {
		return nil, nil // cache miss
	}

	log.Printf("cache hit: %s, raw: %s", key, raw)
	var products []product.Product
	if err = json.Unmarshal([]byte(raw), &products); err != nil {
		return nil, err
	}
	return products, nil
}

// Set stores products under key and registers tags (product IDs) for invalidation.
func (c *cacher) Set(ctx context.Context, key string, products []product.Product, tags ...string) error {
	// Serialize products
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}

	log.Printf("cache set: %s, raw: %s", key, data)

	pipe := c.client.TxPipeline()
	for _, p := range tags {
		pipe.SAdd(ctx, p, key)
		pipe.Expire(ctx, p, c.ttl)
	}

	// Set the cache entry
	pipe.Set(ctx, key, data, c.ttl)
	if _, err = pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

// InvalidateByKey invalidates a tag or key.
// If the key is a tag set, it deletes all member keys and the set itself.
// If the key is a simple cache key, it deletes just that key.
func (c *cacher) InvalidateByKey(ctx context.Context, key string) error {
	// Check if key exists and is a set
	typ, err := c.client.Type(ctx, key).Result()
	if err != nil {
		return err
	}

	switch typ {
	case "set":
		// Fetch all members (cache keys)
		members, err := c.client.SMembers(ctx, key).Result()
		if err != nil {
			return err
		}

		// Include tag set itself for deletion
		all := append(members, key)
		if len(all) > 0 {
			return c.client.Del(ctx, all...).Err()
		}
		return nil
	case "string":
		// Simple cache key
		return c.client.Del(ctx, key).Err()
	default:
		// Key exists but isn't a set or string; delete it anyway
		return c.client.Del(ctx, key).Err()
	}
}

// InvalidateByTags invalidates all cache entries registered under tags.
func (c *cacher) InvalidateByTags(ctx context.Context, tags []string) error {
	keys := make([]string, 0)
	for _, tag := range tags {
		// Fetch all members (cache keys)
		k, err := c.client.SMembers(ctx, tag).Result()
		if err != nil {
			return err
		}

		keys = append(keys, tag)
		keys = append(keys, k...)
	}

	return c.client.Del(ctx, keys...).Err()
}

// InvalidateTagsByPattern deletes all cache entries registered under tags matching pattern.
// Example: DeleteTagsByPattern("tag:product:*")
func (c *cacher) InvalidateTagsByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var tagKeys []string
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		tagKeys = append(tagKeys, keys...)
		cursor = next
		if cursor == 0 {
			break
		}
	}

	// Collect all keys to delete
	all := make([]string, 0, len(tagKeys))
	for _, tag := range tagKeys {
		members, err := c.client.SMembers(ctx, tag).Result()
		if err != nil {
			return err
		}

		all = append(all, members...) // each is a cache key
		all = append(all, tag)        // the tag set itself
	}

	// Deduplicate
	uniq := make(map[string]struct{}, len(all))
	for _, k := range all {
		uniq[k] = struct{}{}
	}

	// Flatten back to slice
	toDelete := make([]string, 0, len(uniq))
	for k := range uniq {
		toDelete = append(toDelete, k)
	}

	// Delete it in one go
	if len(toDelete) > 0 {
		if err := c.client.Del(ctx, toDelete...).Err(); err != nil {
			return err
		}
	}

	return nil
}
