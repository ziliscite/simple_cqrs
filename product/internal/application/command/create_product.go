package command

import (
	"context"
	"encoding/json"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"
)

type CreateProductEvent struct {
	Name     string
	Category string
	Price    float64
}

func NewCreateProduct(name, category string, price float64) (CreateProductEvent, map[string]string) {
	var cp CreateProductEvent

	errs := make(map[string]string)
	if name == "" {
		errs["name"] = "product name is required"
	}

	if category == "" {
		errs["category"] = "product category is required"
	}

	if price <= 0 {
		errs["price"] = "product price must be greater than zero"
	}

	if len(errs) > 0 {
		return cp, errs
	}

	cp.Name = name
	cp.Category = category
	cp.Price = price

	return cp, nil
}

type CreateProductRequest struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type CreateProductHandler interface {
	Handle(ctx context.Context, cmd CreateProductEvent) error
}

type createProductHandler struct {
	repo ports.Repository
	pub  ports.Publisher
}

func NewCreateProductHandler(repo ports.Repository, producer ports.Publisher) CreateProductHandler {
	return &createProductHandler{repo: repo, pub: producer}
}

func (h *createProductHandler) Handle(ctx context.Context, cmd CreateProductEvent) error {
	p, err := product.New(cmd.Name, cmd.Category, cmd.Price)
	if err != nil {
		return err
	}

	if err = h.repo.Create(ctx, p); err != nil {
		return err
	}

	msg, err := json.Marshal(CreateProductRequest{
		ID:       p.ID(),
		Name:     p.Name(),
		Price:    p.Price(),
		Category: p.Category(),
	})
	if err != nil {
		return err
	}

	return h.pub.Publish(ctx, msg, "create")
}
