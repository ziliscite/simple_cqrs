package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/cqrs_search/internal/application"
	"github.com/ziliscite/cqrs_search/internal/application/command"
	"github.com/ziliscite/cqrs_search/internal/ports"
	"github.com/ziliscite/cqrs_search/pkg/rabbit"
	"golang.org/x/sync/errgroup"
)

type consumer struct {
	c   *rabbit.Client
	cmd *application.Command
}

func NewConsumer(c *rabbit.Client, cmd *application.Command) ports.Consumer {
	return &consumer{
		c:   c,
		cmd: cmd,
	}
}

func (c *consumer) Consume(ctx context.Context, queue string) error {
	const (
		consumerTag = "product_event_consumer"
		prefetch    = 12 // max un-acked msgs per consumer
		numChannels = 2  // # of CPU cores or desired parallel channels
	)

	// Use an `errgroup` to capture any channel-level errors
	gChannels, ctx := errgroup.WithContext(ctx)

	for i := 0; i < numChannels; i++ {
		gChannels.Go(func() error {
			// Open a dedicated channel per goroutine
			ch, err := c.c.Channel()
			if err != nil {
				return fmt.Errorf("open AMQP channel: %w", err)
			}
			defer ch.Close()

			// Set per-channel prefetch
			if err = ch.Qos(prefetch, 0, false); err != nil {
				return fmt.Errorf("qos(%d): %w", prefetch, err)
			}

			// Start consuming
			msgs, err := c.c.Consume(ctx, ch, queue, consumerTag, false)
			if err != nil {
				return fmt.Errorf("start consume: %w", err)
			}

			// Use a per-channel errgroup for message‐level concurrency
			gMsgs, msgCtx := errgroup.WithContext(ctx)
			gMsgs.SetLimit(prefetch)

			for {
				select {
				case <-msgCtx.Done():
					return msgCtx.Err()
				case msg, ok := <-msgs:
					if !ok {
						// broker closed the channel
						return nil
					}
					// capture `msg` for this closure
					m := msg

					gMsgs.Go(func() error {
						if err = c.process(msgCtx, &m); err != nil {
							// retry if not redelivered, else drop
							return m.Nack(false, !m.Redelivered)
						}

						return m.Ack(false)
					})
				}
			}
		})
	}

	// wait for all channel‐goroutines (and their msg‐`errgroups`) to finish
	return gChannels.Wait()
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
	var cmd command.CreateProduct
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return errors.New("unable to unmarshal payload")
	}

	errs := cmd.Validate()
	if errs != nil {
		return errs
	}

	return c.cmd.Create.Handle(ctx, cmd)
}

func (c *consumer) UpdateProduct(ctx context.Context, payload []byte) error {
	var cmd command.UpdateProduct
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return errors.New("unable to unmarshal payload")
	}

	errs := cmd.Validate()
	if errs != nil {
		return errs
	}

	return c.cmd.Update.Handle(ctx, cmd)
}

func (c *consumer) DeleteProduct(ctx context.Context, payload []byte) error {
	var cmd command.DeleteProduct
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return errors.New("unable to unmarshal payload")
	}

	errs := cmd.Validate()
	if errs != nil {
		return errs
	}

	return c.cmd.Delete.Handle(ctx, cmd)
}
