package command

import (
	"context"
	"encoding/json"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"
)

type UpdateProductEvent struct {
	ID       string
	Name     string
	Category string
	Price    float64
}

func NewUpdateProduct(id, name, category string, price float64) (UpdateProductEvent, map[string]string) {
	var up UpdateProductEvent

	errs := make(map[string]string)
	if id == "" {
		errs["id"] = "product id is required"
	}

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
		return up, errs
	}

	up.ID = id
	up.Name = name
	up.Category = category
	up.Price = price

	return up, nil
}

type UpdateProductRequest struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type UpdateProductHandler interface {
	Handle(ctx context.Context, cmd UpdateProductEvent) error
}

type updateProductHandler struct {
	repo ports.Repository
	pub  ports.Publisher
}

func NewUpdateProductHandler(repo ports.Repository, producer ports.Publisher) UpdateProductHandler {
	return &updateProductHandler{repo: repo, pub: producer}
}

func (h *updateProductHandler) Handle(ctx context.Context, cmd UpdateProductEvent) error {
	p, err := product.New(cmd.Name, cmd.Category, cmd.Price)
	if err != nil {
		return err
	}
	p.SetID(cmd.ID)

	if err = h.repo.Update(ctx, p); err != nil {
		return err
	}

	msg, err := json.Marshal(UpdateProductRequest{
		ID:       p.ID(),
		Name:     p.Name(),
		Price:    p.Price(),
		Category: p.Category(),
	})
	if err != nil {
		return err
	}

	return h.pub.Publish(ctx, msg, "update")
}
