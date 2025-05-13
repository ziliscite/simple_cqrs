package product

import (
	"encoding/json"
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

func (p *Product) ID() string {
	return p.id.String()
}

func (p *Product) SetID(id string) {
	p.id = ID(id)
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

func (p *Product) Tags() []string {
	return []string{
		"tag:product:" + p.ID(),
		"tag:category:" + p.Category(),
		"tag:name:" + p.Name(),
	}
}

func (p *Product) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID       ID      `json:"id"`
		Name     string  `json:"name"`
		Price    float64 `json:"price"`
		Category string  `json:"category"`
	}{
		ID:       p.id,
		Name:     p.name,
		Price:    p.price,
		Category: p.category,
	})
}

func (p *Product) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with exported fields
	var temp struct {
		ID       ID      `json:"id"`
		Name     string  `json:"name"`
		Price    float64 `json:"price"`
		Category string  `json:"category"`
	}

	// Unmarshal into the temporary struct
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Set the unexported fields from the temporary struct
	p.id = temp.ID
	p.name = temp.Name
	p.price = temp.Price
	p.category = temp.Category

	return nil
}
