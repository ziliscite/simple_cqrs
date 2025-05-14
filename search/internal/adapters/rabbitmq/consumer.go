package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/cqrs_search/internal/application"
	"github.com/ziliscite/cqrs_search/internal/application/command"
	"github.com/ziliscite/cqrs_search/internal/ports"
	"github.com/ziliscite/cqrs_search/pkg/rabbit"
	"log"
	"sync"
)

type consumer struct {
	c   *rabbit.Client
	cmd *application.Command
	q   string
}

func NewConsumer(c *rabbit.Client, exchange, queue, binding string, cmd *application.Command) (ports.Consumer, error) {
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

	return &consumer{
		c:   c,
		cmd: cmd,
		q:   queue,
	}, nil
}

func (c *consumer) Consume() error {
	ch, err := c.c.Channel()
	if err != nil {
		return err
	}
	defer c.c.Put(ch)

	if err = ch.Qos(1, 0, false); err != nil {
		return err
	}

	bus, err := c.c.Consume(context.Background(), ch, "product_queue", "product_event", false)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for msg := range bus {
		m := msg
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("New Message: %v", m)

			if err = c.process(context.Background(), &m); err != nil {
				m.Nack(false, !m.Redelivered)
				return
			}
			
			m.Ack(false)
		}()
	}

	wg.Wait()
	return nil
}

func (c *consumer) process(ctx context.Context, msg *amqp091.Delivery) error {
	event := msg.Headers["event_type"].(string)

	switch event {
	case "create":
		if err := c.CreateProduct(ctx, msg.Body); err != nil {
			return err
		}
	case "update":
		if err := c.UpdateProduct(ctx, msg.Body); err != nil {
			return err
		}
	case "delete":
		if err := c.DeleteProduct(ctx, msg.Body); err != nil {
			return err
		}
	default:
		msg.Redelivered = true // don't re-queue
		return errors.New("unknown event type")
	}

	return nil
}

func (c *consumer) CreateProduct(ctx context.Context, payload []byte) error {
	var request struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Category string  `json:"category"`
		Price    float64 `json:"price"`
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		return errors.New("unable to unmarshal payload")
	}

	cmd, errs := command.NewCreateProduct(request.ID, request.Name, request.Category, request.Price)
	if errs != nil {
		return errs
	}

	log.Printf("CreateProduct: %v %v %v %v", cmd.ID, cmd.Name, cmd.Category, cmd.Price)
	return c.cmd.Create.Handle(ctx, cmd)
}

func (c *consumer) UpdateProduct(ctx context.Context, payload []byte) error {
	var request struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Category string  `json:"category"`
		Price    float64 `json:"price"`
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		return errors.New("unable to unmarshal payload")
	}

	cmd, errs := command.NewUpdateProduct(request.ID, request.Name, request.Category, request.Price)
	if errs != nil {
		return errs
	}

	return c.cmd.Update.Handle(ctx, cmd)
}

func (c *consumer) DeleteProduct(ctx context.Context, payload []byte) error {
	var request struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		return errors.New("unable to unmarshal payload")
	}

	cmd, errs := command.NewDeleteProduct(request.ID)
	if errs != nil {
		return errs
	}

	return c.cmd.Delete.Handle(ctx, cmd)
}
