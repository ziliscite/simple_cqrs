package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"io"
	"log"
	"strings"

	"github.com/ziliscite/cqrs_search/internal/domain/product"
)

func (r *repo) Search(ctx context.Context, opts *product.Search) ([]product.Product, error) {
	log.Printf("query: %s, %s", opts.Name(), opts.Category())

	// Build bool query
	boolQuery := map[string]interface{}{"bool": map[string]interface{}{
		"must":   []interface{}{},
		"filter": []interface{}{},
	}}

	// match name
	if opts.Name() != "" {
		boolQuery["bool"].(map[string]interface{})["must"] = append(
			boolQuery["bool"].(map[string]interface{})["must"].([]interface{}),
			map[string]interface{}{"match": map[string]interface{}{"name": opts.Name()}},
		)
	}

	// term category
	if opts.Category() != "" {
		boolQuery["bool"].(map[string]interface{})["filter"] = append(
			boolQuery["bool"].(map[string]interface{})["filter"].([]interface{}),
			map[string]interface{}{"term": map[string]interface{}{"category": opts.Category()}},
		)
	}

	// price range
	minPrice, maxPrice := opts.PriceRange()
	if minPrice != nil || maxPrice != nil {
		rangeQ := map[string]interface{}{"range": map[string]interface{}{"price": map[string]interface{}{}}}
		if minPrice != nil {
			rangeQ["range"].(map[string]interface{})["price"].(map[string]interface{})["gte"] = *minPrice
		}
		if maxPrice != nil {
			rangeQ["range"].(map[string]interface{})["price"].(map[string]interface{})["lte"] = *maxPrice
		}
		boolQuery["bool"].(map[string]interface{})["filter"] = append(
			boolQuery["bool"].(map[string]interface{})["filter"].([]interface{}),
			rangeQ,
		)
	}

	// marshal query
	query, err := json.Marshal(map[string]interface{}{"query": boolQuery})
	if err != nil {
		return nil, err
	}

	// pagination
	size := opts.PageSize()
	from := opts.Offset()

	// sort
	field, by := opts.SortBy()
	var sortOpts []string
	if field != "" {
		sortOpts = []string{fmt.Sprintf("%s:%s", field, by)}
	}

	log.Printf("query: %s", string(query))
	req := esapi.SearchRequest{
		Index:          []string{r.idx},
		Body:           io.NopCloser(strings.NewReader(string(query))),
		From:           &from,
		Size:           &size,
		Sort:           sortOpts, // will only be non-empty if field != ""
		TrackTotalHits: true,
	}

	// execute search
	res, err := req.Do(ctx, r.c)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error getting document: %s", res.String())
	}

	log.Printf("response: %s", res.String())

	// parse hits
	var esResp struct {
		Hits struct {
			Hits []struct {
				Source product.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err = json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	prods := make([]product.Product, len(esResp.Hits.Hits))
	for i, h := range esResp.Hits.Hits {
		log.Printf("product: %v", h.Source)
		prods[i] = h.Source
	}
	return prods, nil
}

func (r *repo) GetByID(ctx context.Context, id string) (*product.Product, error) {
	req := esapi.GetRequest{
		Index:      r.idx,
		DocumentID: id,
	}

	res, err := req.Do(ctx, r.c)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error getting document: %s", res.String())
	}

	var response struct {
		Source product.Product `json:"_source"`
	}
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Source, nil
}
