package product

import (
	"errors"
	"github.com/google/uuid"
)

type ID string

func NewID() ID {
	return ID(uuid.New().String())
}

func (i ID) String() string {
	return string(i)
}

type Product struct {
	id       ID
	name     string
	price    float64
	category string
}

func New(name, category string, price float64) (*Product, error) {
	if name == "" {
		return nil, errors.New("product name is required")
	}

	if category == "" {
		return nil, errors.New("product category is required")
	}

	if price <= 0 {
		return nil, errors.New("product price must be greater than zero")
	}

	return &Product{
		id:       NewID(),
		name:     name,
		price:    price,
		category: category,
	}, nil
}

func (p *Product) SetID(id string) {
	p.id = ID(id)
}

func (p *Product) ID() string {
	return p.id.String()
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Price() float64 {
	return p.price
}

func (p *Product) Category() string {
	return p.category
}
