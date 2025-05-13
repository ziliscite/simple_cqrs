package application

import (
	"github.com/ziliscite/cqrs_product/internal/application/command"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"
)

type Service struct {
	Create command.CreateProductHandler
	Update command.UpdateProductHandler
	Delete command.DeleteProductHandler
}

func NewService(repo ports.Repository, cu ports.Publisher[*product.Product], d ports.Publisher[product.ID]) Service {
	return Service{
		Create: command.NewCreateProductHandler(repo, cu),
		Update: command.NewUpdateProductHandler(repo, cu),
		Delete: command.NewDeleteProductHandler(repo, d),
	}
}
