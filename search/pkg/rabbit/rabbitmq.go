// Package rabbit provides a client for interacting with RabbitMQ.
//
// It provides a Client which can be used to create channels, queues,
// exchanges, and bindings. The Client is safe for concurrent use and
// provides a pool of channels so that a new channel does not have to be
// created for every operation.
package rabbit

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

// Dial connects to RabbitMQ and returns a new connection.
func Dial(username, password, host, port, vhost string) (*amqp.Connection, error) {
	// Set up the Connection to RabbitMQ host using AMQP
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/%s", username, password, host, port, vhost))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Client is a client for interacting with RabbitMQ.
type Client struct {
	conn *amqp.Connection
	pool *sync.Pool
}

// NewClient returns a new Client.
func NewClient(conn *amqp.Connection) *Client {
	return &Client{
		conn: conn,
		pool: &sync.Pool{
			New: func() interface{} {
				ch, err := conn.Channel()
				if err != nil {
					return nil
				}

				// Puts the Channel in confirmation mode, which will allow waiting for ACK or NACK from the receiver
				if err = ch.Confirm(false); err != nil {
					return nil
				}

				return ch
			},
		},
	}
}

// Channel returns a channel from the pool.
func (p *Client) Channel() (*amqp.Channel, error) {
	ch, ok := p.pool.Get().(*amqp.Channel)
	if !ok {
		return nil, fmt.Errorf("could not get channel from pool")
	}

	if ch == nil {
		return nil, fmt.Errorf("channel from pool is nil")
	}

	return ch, nil
}

// ChannelWithConfirm returns a channel from the pool and puts it in confirmation mode.
func (p *Client) ChannelWithConfirm() (*amqp.Channel, error) {
	ch, ok := p.pool.Get().(*amqp.Channel)
	if !ok {
		return nil, fmt.Errorf("could not get channel from pool")
	}

	if ch == nil {
		return nil, fmt.Errorf("channel from pool is nil")
	}

	if err := ch.Confirm(false); err != nil {
		return nil, err
	}

	return ch, nil
}

// Put puts a channel back into the pool.
func (p *Client) Put(ch *amqp.Channel) {
	p.pool.Put(ch)
}

// CreateQueue creates a new queue.
func (p *Client) CreateQueue(ch *amqp.Channel, queueName string, durable, autoDelete bool) error {
	if _, err := ch.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	return nil
}

// ExchangeType is the type of exchange.
type ExchangeType int

const (
	ExchangeDirect  ExchangeType = iota // direct exchange
	ExchangeFanout                      // fanout exchange
	ExchangeTopic                       // topic exchange
	ExchangeHeaders                     // headers exchange
)

func (e ExchangeType) String() string {
	return map[ExchangeType]string{
		ExchangeDirect:  "direct",
		ExchangeFanout:  "fanout",
		ExchangeTopic:   "topic",
		ExchangeHeaders: "headers",
	}[e]
}

// CreateExchange creates a new exchange.
func (p *Client) CreateExchange(ch *amqp.Channel, exchangeName string, exchangeType ExchangeType, durable, autoDelete bool) error {
	return ch.ExchangeDeclare(
		exchangeName,
		exchangeType.String(),
		durable,
		autoDelete,
		false,
		false,
		nil,
	)
}

// CreateBinding creates a new binding between a queue and an exchange.
func (p *Client) CreateBinding(ch *amqp.Channel, name, binding, exchange string) error {
	// having nowait set to false will cause the channel to return an error and close if it cannot bind.
	// the final argument is the extra headers, but we won't be doing that now
	return ch.QueueBind(
		name,
		binding,
		exchange,
		false,
		nil,
	)
}

// Send is used to publish a payload onto an exchange with a given routingkey
func (p *Client) Send(ctx context.Context, ch *amqp.Channel, exchange, routingKey string, options amqp.Publishing) error {
	return ch.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		// 'Mandatory' is used when we HAVE to have the message return an error, if there is no route or queue then
		// setting this to true will make the message bounce back
		// If this is False, and the message fails to deliver, it will be dropped
		true, // mandatory
		// 'immediate' Removed in MQ 3 or up https://blog.rabbitmq.com/posts/2012/11/breaking-things-with-rabbitmq-3-0§
		false,   // immediate
		options, // amqp publishing struct
	)
}

func (p *Client) SendJSON(ctx context.Context, ch *amqp.Channel, exchange, routingKey string, payload []byte, headers map[string]interface{}) error {
	return p.Send(ctx, ch, exchange, routingKey, amqp.Publishing{
		Headers:      headers,
		ContentType:  "application/json",
		Body:         payload,
		DeliveryMode: amqp.Persistent,
	})
}

// SendDeferred is used to publish a payload onto an exchange with a given routingkey
func (p *Client) SendDeferred(ctx context.Context, ch *amqp.Channel, exchange, routingKey string, options amqp.Publishing) error {
	confirmation, err := ch.PublishWithDeferredConfirmWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		// Mandatory is used when we HAVE to have the message return an error, if there is no route or queue then
		// setting this to true will make the message bounce back
		// If this is False, and the message fails to deliver, it will be dropped
		true,    // mandatory
		false,   // immediate
		options, // amqp publishing struct
	)
	if err != nil {
		return err
	}

	// Blocks until ACK from Server is received
	confirmation.Wait()
	return nil
}

func (p *Client) SendDeferredJSON(ctx context.Context, ch *amqp.Channel, exchange, routingKey string, payload []byte, headers map[string]interface{}) error {
	return p.SendDeferred(ctx, ch, exchange, routingKey, amqp.Publishing{
		Headers:      headers,
		ContentType:  "application/json",
		Body:         payload,
		DeliveryMode: amqp.Persistent,
	})
}

func (p *Client) Consume(ctx context.Context, ch *amqp.Channel, queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	return ch.ConsumeWithContext(ctx, queue, consumer, autoAck, false, false, false, nil)
}
