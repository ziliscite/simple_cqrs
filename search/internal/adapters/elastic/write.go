package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"

	"github.com/ziliscite/cqrs_search/internal/domain/product"
)

func (r *repo) Create(ctx context.Context, p *product.Product) error {
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      r.idx,
		DocumentID: p.ID(),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.c)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	log.Printf("create product: %s %s %s $%.2f", p.ID(), p.Name(), p.Category(), p.Price())
	log.Printf("index response: %s", res.String())

	if res.IsError() {
		return fmt.Errorf("index error: %s", res.String())
	}
	return nil
}

func (r *repo) Update(ctx context.Context, p *product.Product) error {
	body, _ := json.Marshal(map[string]interface{}{
		"doc": p,
	})

	req := esapi.UpdateRequest{
		Index:      r.idx,
		DocumentID: p.ID(),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.c)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating document: %s", res.String())
	}
	return nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	req := esapi.DeleteRequest{
		Index:      r.idx,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.c)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting document: %s", res.String())
	}
	return nil
}
