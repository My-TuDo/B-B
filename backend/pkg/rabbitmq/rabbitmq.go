package rabbitmq

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queueTranscode = "video.transcode"
)

// TranscodeMessage is the payload published to the transcode queue.
type TranscodeMessage struct {
	VideoID uint `json:"video_id"`
}

// Client wraps an AMQP connection and channel for publishing.
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
	done    chan struct{}
}

// Config holds RabbitMQ connection parameters.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
}

// Init creates a RabbitMQ connection, declares the transcode queue, and returns a Client.
func Init(cfg *Config) (*Client, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

	var conn *amqp.Connection
	var err error

	// Retry connecting up to 5 times
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(dsn)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.Init: dial failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq.Init: open channel failed: %w", err)
	}

	// Declare a durable queue
	_, err = ch.QueueDeclare(
		queueTranscode,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("rabbitmq.Init: declare queue failed: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: ch,
		done:    make(chan struct{}),
	}, nil
}

// PublishTranscodeTask sends a transcode job to the queue.
func (c *Client) PublishTranscodeTask(videoID uint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := TranscodeMessage{VideoID: videoID}
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("rabbitmq.PublishTranscodeTask: marshal: %w", err)
	}

	err = c.channel.Publish(
		"",               // exchange
		queueTranscode,   // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("rabbitmq.PublishTranscodeTask: publish: %w", err)
	}
	return nil
}

// ConsumeTranscodeTask returns a channel of TranscodeMessage deliveries.
func (c *Client) ConsumeTranscodeTask() (<-chan amqp.Delivery, error) {
	deliveries, err := c.channel.Consume(
		queueTranscode,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.ConsumeTranscodeTask: %w", err)
	}
	return deliveries, nil
}

// Close shuts down the channel and connection.
func (c *Client) Close() {
	close(c.done)
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
