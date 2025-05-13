package main

import (
	"context"
	"github.com/ziliscite/cqrs_search/internal/adapters/elastic"
	handler "github.com/ziliscite/cqrs_search/internal/adapters/http_handler"
	"github.com/ziliscite/cqrs_search/internal/adapters/rabbitmq"
	cache "github.com/ziliscite/cqrs_search/internal/adapters/redis_cache"
	"github.com/ziliscite/cqrs_search/internal/application"
	"github.com/ziliscite/cqrs_search/pkg/rabbit"
	"time"
)

func main() {
	cfg := getConfig()

	// initialize repo
	ESClient, err := elastic.NewESClient(cfg.e.host, cfg.e.port)
	if err != nil {
		return
	}

	repo, err := elastic.NewRepository(ESClient, cfg.e.index) // "products"
	if err != nil {
		return
	}

	redisClient, err := cache.NewRedisClient(cfg.r.user, cfg.r.password, cfg.r.host, cfg.r.port, cfg.r.db)
	if err != nil {
		return
	}
	cacher := cache.NewCacher(redisClient, cfg.r.ttl)

	// initialize services
	app := application.NewService(repo, cacher)

	// initialize drivers
	rabbitConn, err := rabbit.Dial(cfg.mq.user, cfg.mq.pass, cfg.mq.host, cfg.mq.vhost)
	if err != nil {
		return
	}
	rabbitClient := rabbit.NewClient(rabbitConn)

	consumer := rabbitmq.NewConsumer(rabbitClient, app.Command)
	srv := handler.NewHandler(app.Query)

	// start server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		if err = consumer.Consume(ctx, cfg.mq.queue); err != nil {
			return
		}
	}()

	if err = srv.Run(cfg.h.addr()); err != nil {
		return
	}
}
