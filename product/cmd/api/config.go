package main

import (
	"flag"
	"os"
	"sync"
)

type HTTP struct {
	host string
	port string
}

func (h HTTP) addr() string {
	return h.host + ":" + h.port
}

type DB struct {
	host string
	port string
	user string
	pass string
	db   string
	ssl  bool
}

func (d DB) dsn() string {
	dsn := "postgres://" + d.user + ":" + d.pass + "@" + d.host + ":" + d.port + "/" + d.db
	if !d.ssl {
		return dsn + "?sslmode=disable"
	}
	return dsn
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
	db DB
	mq MQ
	h  HTTP
}

var (
	instance Config
	once     sync.Once
)

func getConfig() Config {
	once.Do(func() {
		instance = Config{}

		flag.StringVar(&instance.db.host, "db-host", os.Getenv("POSTGRES_HOST"), "Database host")
		flag.StringVar(&instance.db.port, "db-port", os.Getenv("POSTGRES_PORT"), "Database port")
		flag.StringVar(&instance.db.user, "db-user", os.Getenv("POSTGRES_USER"), "Database user")
		flag.StringVar(&instance.db.pass, "db-pass", os.Getenv("POSTGRES_PASSWORD"), "Database password")
		flag.StringVar(&instance.db.db, "db-db", os.Getenv("POSTGRES_DB"), "Database name")

		var ssl bool
		sslStr := os.Getenv("POSTGRES_SSL")
		if sslStr == "TRUE" {
			ssl = true
		} else {
			ssl = false
		}
		flag.BoolVar(&instance.db.ssl, "db-ssl", ssl, "Database ssl")

		flag.StringVar(&instance.mq.host, "mq-host", os.Getenv("RABBITMQ_HOST"), "RabbitMQ host")
		flag.StringVar(&instance.mq.user, "mq-user", os.Getenv("RABBITMQ_USER"), "RabbitMQ user")
		flag.StringVar(&instance.mq.pass, "mq-pass", os.Getenv("RABBITMQ_PASSWORD"), "RabbitMQ password")
		flag.StringVar(&instance.mq.vhost, "mq-vhost", os.Getenv("RABBITMQ_VHOST"), "RabbitMQ vhost")

		flag.StringVar(&instance.mq.exchange, "mq-exchange", os.Getenv("RABBITMQ_EXCHANGE"), "RabbitMQ exchange")
		flag.StringVar(&instance.mq.queue, "mq-queue", os.Getenv("RABBITMQ_QUEUE"), "RabbitMQ queue")
		flag.StringVar(&instance.mq.binding, "mq-binding", os.Getenv("RABBITMQ_BINDING"), "RabbitMQ binding")

		flag.StringVar(&instance.h.host, "http-host", os.Getenv("HTTP_HOST"), "HTTP host")
		flag.StringVar(&instance.h.port, "http-port", os.Getenv("HTTP_PORT"), "HTTP port")

		flag.Parse()
	})

	return instance
}
