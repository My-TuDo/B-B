// Package rabbitmq 提供 RabbitMQ 消息队列的连接管理和发布/消费操作。
// 用于将视频转码任务异步投递到队列，由 Worker 消费处理。
// 支持断线重连（最多 5 次），队列声明为持久化。
package rabbitmq

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// queueTranscode 视频转码队列名称。
const (
	queueTranscode = "video.transcode"
)

// TranscodeMessage 转码任务消息体。
type TranscodeMessage struct {
	VideoID uint `json:"video_id"` // 待转码的视频 ID
}

// Client 封装 AMQP 连接和通道，用于发布和消费消息。
type Client struct {
	conn    *amqp.Connection // AMQP 连接
	channel *amqp.Channel    // AMQP 通道
	mu      sync.Mutex       // 发布操作互斥锁
	done    chan struct{}    // 关闭信号通道
}

// Config RabbitMQ 连接参数。
type Config struct {
	Host     string // 主机地址
	Port     string // 端口
	User     string // 用户名
	Password string // 密码
}

// Init 连接 RabbitMQ，声明转码队列，返回 Client。
// 连接失败会重试最多 5 次，每次间隔 2 秒。
func Init(cfg *Config) (*Client, error) {
	// 构造 AMQP DSN
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

	var conn *amqp.Connection
	var err error

	// 重试连接，最多 5 次
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

	// 打开通道
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq.Init: open channel failed: %w", err)
	}

	// 声明持久化队列（durable=true，服务重启后队列不丢失）
	_, err = ch.QueueDeclare(
		queueTranscode, // 队列名
		true,           // durable — 持久化
		false,          // auto-delete — 不自动删除
		false,          // exclusive — 非独占
		false,          // no-wait — 等待服务器确认
		nil,            // 额外参数
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

// PublishTranscodeTask 将转码任务发布到队列。
// 消息体为 JSON 序列化的 TranscodeMessage，投递模式为持久化。
func (c *Client) PublishTranscodeTask(videoID uint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 构造消息体
	msg := TranscodeMessage{VideoID: videoID}
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("rabbitmq.PublishTranscodeTask: marshal: %w", err)
	}

	// 发布到默认 exchange，routing key 为队列名
	err = c.channel.Publish(
		"",             // exchange — 使用默认交换机
		queueTranscode, // routing key — 队列名
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // 消息持久化
		},
	)
	if err != nil {
		return fmt.Errorf("rabbitmq.PublishTranscodeTask: publish: %w", err)
	}
	return nil
}

// ConsumeTranscodeTask 返回转码队列的消费通道。
// 使用手动确认模式（auto-ack=false），消费方需显式 Ack/Nack。
func (c *Client) ConsumeTranscodeTask() (<-chan amqp.Delivery, error) {
	deliveries, err := c.channel.Consume(
		queueTranscode, // 队列名
		"",             // consumer tag — 自动生成
		false,          // auto-ack — 手动确认
		false,          // exclusive — 非独占
		false,          // no-local
		false,          // no-wait
		nil,            // 额外参数
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.ConsumeTranscodeTask: %w", err)
	}
	return deliveries, nil
}

// Close 关闭通道和连接，发送 done 信号。
func (c *Client) Close() {
	close(c.done)
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
