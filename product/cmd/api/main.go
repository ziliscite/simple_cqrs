package main

import (
	"context"
	"github.com/ziliscite/cqrs_product/internal/adapters/http_handler"
	"github.com/ziliscite/cqrs_product/internal/adapters/postgresql"
	"github.com/ziliscite/cqrs_product/internal/adapters/rabbitmq"
	"github.com/ziliscite/cqrs_product/internal/application"
	"github.com/ziliscite/cqrs_product/pkg/postgres"
	"github.com/ziliscite/cqrs_product/pkg/rabbit"
	"time"
)

func main() {
	cfg := getConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.Open(ctx, cfg.db.dsn())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = postgres.AutoMigrate(cfg.db.dsn()); err != nil {
		return
	}

	mq, err := rabbit.Dial(cfg.mq.user, cfg.mq.pass, cfg.mq.host, cfg.mq.port, cfg.mq.vhost)
	if err != nil {
		panic(err)
	}
	defer mq.Close()

	pub := rabbit.NewClient(mq)
	repo := postgresql.NewRepository(db)

	cu, err := rabbitmq.NewProducer(pub, cfg.mq.exchange, cfg.mq.queue, cfg.mq.binding)
	if err != nil {
		panic(err)
	}

	app := application.NewService(repo, cu)
	srv := handler.NewHandler(app)

	if err = srv.Run(cfg.h.addr()); err != nil {
		panic(err)
	}
}
