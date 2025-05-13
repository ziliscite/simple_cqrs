package main

import (
	"github.com/ziliscite/cqrs_search/internal/adapters/elastic"
	handler "github.com/ziliscite/cqrs_search/internal/adapters/http_handler"
	"github.com/ziliscite/cqrs_search/internal/adapters/rabbitmq"
	cache "github.com/ziliscite/cqrs_search/internal/adapters/redis_cache"
	"github.com/ziliscite/cqrs_search/internal/application"
	"github.com/ziliscite/cqrs_search/pkg/rabbit"
)

func main() {
	cfg := getConfig()

	// initialize repo
	ESClient, err := elastic.NewESClient(cfg.e.host, cfg.e.port)
	if err != nil {
		panic(err)
	}

	repo, err := elastic.NewRepository(ESClient, cfg.e.index) // "products"
	if err != nil {
		panic(err)
	}

	redisClient, err := cache.NewRedisClient(cfg.r.user, cfg.r.password, cfg.r.host, cfg.r.port, cfg.r.db)
	if err != nil {
		panic(err)
	}
	cacher := cache.NewCacher(redisClient, cfg.r.ttl)

	// initialize services
	app := application.NewService(repo, cacher)

	// initialize drivers
	rabbitConn, err := rabbit.Dial(cfg.mq.user, cfg.mq.pass, cfg.mq.host, cfg.mq.port, cfg.mq.vhost)
	if err != nil {
		panic(err)
	}
	rabbitClient := rabbit.NewClient(rabbitConn)

	consumer, err := rabbitmq.NewConsumer(rabbitClient, cfg.mq.exchange, cfg.mq.queue, cfg.mq.binding, app.Command)
	if err != nil {
		panic(err)
	}

	srv := handler.NewHandler(app.Query)

	// start server
	go func() {
		if err = consumer.Consume(); err != nil {
			panic(err)
		}
	}()

	if err = srv.Run(cfg.h.addr()); err != nil {
		panic(err)
	}
}
