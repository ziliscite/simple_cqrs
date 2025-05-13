package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/ziliscite/cqrs_search/internal/ports"
)

func NewESClient(host, port string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{host + ":" + port},
	}

	return elasticsearch.NewClient(cfg)
}

type repo struct {
	c   *elasticsearch.Client
	idx string
}

func NewRepository(client *elasticsearch.Client, index string) (ports.Repository, error) {
	r := &repo{c: client, idx: index}
	if err := r.EnsureIndex(context.Background()); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *repo) EnsureIndex(ctx context.Context) error {
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "text",
				},
				"price": map[string]interface{}{
					"type": "double",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal index mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: r.idx,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, r.c)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 400 {
		// If it already exists, you might get a 400; you can ignore or handle specifically
		return fmt.Errorf("index creation failed: %s", res.String())
	}
	return nil
}
