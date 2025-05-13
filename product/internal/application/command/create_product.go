package command

import (
	"context"
	"errors"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"
)

type CreateProduct struct {
	Name     string
	Category string
	Price    float64
}

func (p CreateProduct) Validate() map[string]error {
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
	repo ports.Repository
	pub  ports.Publisher[*product.Product]
}

func NewCreateProductHandler(repo ports.Repository, producer ports.Publisher[*product.Product]) CreateProductHandler {
	return &createProductHandler{repo: repo, pub: producer}
}

func (h *createProductHandler) Handle(ctx context.Context, cmd CreateProduct) error {
	p, err := product.New(cmd.Name, cmd.Category, cmd.Price)
	if err != nil {
		return err
	}

	e := product.NewEvent(product.Create, p)
	if err = h.pub.Publish(ctx, e); err != nil {
		return err
	}

	return h.repo.Create(ctx, p)
}
