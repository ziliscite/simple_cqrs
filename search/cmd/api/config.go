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
	port  string
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
		flag.StringVar(&instance.r.db, "redis-db", os.Getenv("REDIS_DATABASE"), "Redis db")

		var ttlStr string
		flag.StringVar(&ttlStr, "redis-ttl", os.Getenv("REDIS_TTL"), "Redis TTL in seconds")
		ttl, err := time.ParseDuration(ttlStr + "s")
		if err != nil {
			panic(err)
		}
		instance.r.ttl = ttl

		flag.StringVar(&instance.e.host, "elastic-host", os.Getenv("ELASTICSEARCH_HOST"), "Elastic host")
		flag.StringVar(&instance.e.port, "elastic-port", os.Getenv("ELASTICSEARCH_PORT"), "Elastic port")
		flag.StringVar(&instance.e.index, "elastic-index", os.Getenv("ELASTICSEARCH_INDEX"), "Elastic index")

		flag.StringVar(&instance.mq.host, "mq-host", os.Getenv("RABBITMQ_HOST"), "RabbitMQ host")
		flag.StringVar(&instance.mq.port, "mq-port", os.Getenv("RABBITMQ_PORT"), "RabbitMQ port")
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
