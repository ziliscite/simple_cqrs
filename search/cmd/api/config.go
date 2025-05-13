package main

import (
	"flag"
	"os"
	"sync"
	"time"
)

type HTTP struct {
	host string
	port string
}

func (h HTTP) addr() string {
	return h.host + ":" + h.port
}

type Redis struct {
	host     string
	port     string
	user     string
	password string
	db       string

	ttl time.Duration // in seconds
}

type Elastic struct {
	host  string
	port  string
	index string
}

type MQ struct {
	host  string
	user  string
	pass  string
	vhost string

	exchange string
	queue    string
	binding  string
}

type Config struct {
	h  HTTP
	r  Redis
	e  Elastic
	mq MQ
}

var (
	instance Config
	once     sync.Once
)

func getConfig() Config {
	once.Do(func() {
		instance = Config{}
		flag.StringVar(&instance.h.host, "http-host", os.Getenv("HTTP_HOST"), "HTTP host")
		flag.StringVar(&instance.h.port, "http-port", os.Getenv("HTTP_PORT"), "HTTP port")

		flag.StringVar(&instance.r.host, "redis-host", os.Getenv("REDIS_HOST"), "Redis host")
		flag.StringVar(&instance.r.port, "redis-port", os.Getenv("REDIS_PORT"), "Redis port")
		flag.StringVar(&instance.r.user, "redis-user", os.Getenv("REDIS_USER"), "Redis user")
		flag.StringVar(&instance.r.password, "redis-password", os.Getenv("REDIS_PASSWORD"), "Redis password")
		flag.StringVar(&instance.r.db, "redis-db", os.Getenv("REDIS_DB"), "Redis db")

		var ttlStr string
		flag.StringVar(&ttlStr, "redis-ttl", os.Getenv("REDIS_TTL"), "Redis db")
		ttl, err := time.ParseDuration(ttlStr)
		if err != nil {
			panic(err)
		}
		instance.r.ttl = ttl

		flag.StringVar(&instance.e.host, "elastic-host", os.Getenv("ELASTIC_HOST"), "Elastic host")
		flag.StringVar(&instance.e.port, "elastic-port", os.Getenv("ELASTIC_PORT"), "Elastic port")
		flag.StringVar(&instance.e.index, "elastic-index", os.Getenv("ELASTIC_INDEX"), "Elastic index")

		flag.StringVar(&instance.mq.host, "mq-host", os.Getenv("RABBITMQ_HOST"), "RabbitMQ host")
		flag.StringVar(&instance.mq.user, "mq-user", os.Getenv("RABBITMQ_USER"), "RabbitMQ user")
		flag.StringVar(&instance.mq.pass, "mq-pass", os.Getenv("RABBITMQ_PASSWORD"), "RabbitMQ password")
		flag.StringVar(&instance.mq.vhost, "mq-vhost", os.Getenv("RABBITMQ_VHOST"), "RabbitMQ vhost")

		flag.StringVar(&instance.mq.exchange, "mq-exchange", os.Getenv("RABBITMQ_EXCHANGE"), "RabbitMQ exchange")
		flag.StringVar(&instance.mq.queue, "mq-queue", os.Getenv("RABBITMQ_QUEUE"), "RabbitMQ queue")
		flag.StringVar(&instance.mq.binding, "mq-binding", os.Getenv("RABBITMQ_BINDING"), "RabbitMQ binding")

		flag.Parse()
	})

	return instance
}
