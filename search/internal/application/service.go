package application

import (
	"github.com/ziliscite/cqrs_search/internal/application/command"
	"github.com/ziliscite/cqrs_search/internal/application/query"
	"github.com/ziliscite/cqrs_search/internal/ports"
)

type Command struct {
	Create command.CreateProductHandler
	Update command.UpdateProductHandler
	Delete command.DeleteProductHandler
}

func NewCommand(repo ports.WriteRepository, cacher ports.CacheInvalidator) *Command {
	return &Command{
		Create: command.NewCreateProductHandler(repo),
		Update: command.NewUpdateProductHandler(repo, cacher),
		Delete: command.NewDeleteProductHandler(repo, cacher),
	}
}

type Query struct {
	Get    query.GetProductHandler
	Search query.SearchProductHandler
}

func NewQuery(repo ports.ReadRepository, cacher ports.CacheWriteReader) *Query {
	return &Query{
		Get:    query.NewGetProductHandler(repo),
		Search: query.NewSearchProductHandler(repo, cacher),
	}
}

type Service struct {
	Command *Command
	Query   *Query
}

func NewService(repo ports.Repository, cacher ports.Cacher) Service {
	return Service{
		Command: NewCommand(repo, cacher),
		Query:   NewQuery(repo, cacher),
	}
}
