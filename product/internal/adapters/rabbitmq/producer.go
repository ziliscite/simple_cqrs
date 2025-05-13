package rabbitmq

import (
	"context"
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/cqrs_product/internal/domain/product"
	"github.com/ziliscite/cqrs_product/internal/ports"
	"github.com/ziliscite/cqrs_product/pkg/rabbit"
)

type producer[M product.Message] struct {
	cfg struct {
		exchange string
		queue    string
		binding  string
	}

	c *rabbit.Client
}

func NewProducer[M product.Message](c *rabbit.Client, exchange, queue, binding string) (ports.Publisher[M], error) {
	ch, err := c.Channel()
	if err != nil {
		return nil, err
	}
	defer c.Put(ch)

	// product
	if err = c.CreateExchange(ch, exchange, rabbit.ExchangeDirect, true, false); err != nil {
		return nil, err
	}

	// product_queue
	if err = c.CreateQueue(ch, queue, true, false); err != nil {
		return nil, err
	}

	// product_event
	if err = c.CreateBinding(ch, queue, binding, exchange); err != nil {
		return nil, err
	}

	return &producer[M]{
		c: c,
		cfg: struct {
			exchange string
			queue    string
			binding  string
		}{
			exchange: exchange,
			queue:    queue,
			binding:  binding,
		},
	}, nil
}

func (p *producer[M]) Publish(ctx context.Context, event product.Event[M]) error {
	ch, err := p.c.Channel()
	if err != nil {
		return err
	}
	defer p.c.Put(ch)

	msg, err := json.Marshal(event.Message)
	if err != nil {
		return err
	}

	return p.c.SendDeferredJSON(ctx, ch, p.cfg.exchange, p.cfg.binding, msg, amqp091.Table{
		"event_type": event.Action,
	})
}
