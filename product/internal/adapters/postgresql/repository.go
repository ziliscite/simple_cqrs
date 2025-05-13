package postgresql

import (
	"context"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) ports.Repository {
	return &repo{
		db: db,
	}
}

func (r *repo) Create(ctx context.Context, product *product.Product) error {
	if _, err := r.db.Exec(ctx, `
		INSERT INTO products (id, name, price, category) VALUES ($1, $2, $3, $4)
	`, product.ID(), product.Name(), product.Price(), product.Category(),
	); err != nil {
		return err
	}

	return nil
}

func (r *repo) Update(ctx context.Context, product *product.Product) error {
	if _, err := r.db.Exec(ctx, `
		UPDATE products SET name = $2, price = $3, category = $4 WHERE id = $1
	`, product.ID(), product.Name(), product.Price(), product.Category(),
	); err != nil {
		return err
	}

	return nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	if _, err := r.db.Exec(ctx, `
		DELETE FROM products WHERE id = $1
	`, id,
	); err != nil {
		return err
	}

	return nil
}
