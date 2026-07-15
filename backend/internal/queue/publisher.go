package queue

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const TaskQueue = "agent.tasks"

type TaskMessage struct {
	TaskID      string `json:"task_id"`
	VehicleID   uint64 `json:"vehicle_id"`
	Command     string `json:"command"`
	CallbackURL string `json:"callback_url"`
}

type Publisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewPublisher 声明持久化主队列与死信队列，并启用发布确认。
func NewPublisher(url string) (*Publisher, error) {
	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ: %w", err)
	}
	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("创建 RabbitMQ channel: %w", err)
	}
	if err := channel.ExchangeDeclare("agent.tasks.dlx", "fanout", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if _, err := channel.QueueDeclare("agent.tasks.dead", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if err := channel.QueueBind("agent.tasks.dead", "", "agent.tasks.dlx", false, nil); err != nil {
		return nil, err
	}
	if _, err := channel.QueueDeclare(TaskQueue, true, false, false, false, amqp.Table{"x-dead-letter-exchange": "agent.tasks.dlx"}); err != nil {
		return nil, err
	}
	if err := channel.Confirm(false); err != nil {
		return nil, err
	}
	return &Publisher{connection: connection, channel: channel}, nil
}

func (p *Publisher) Publish(ctx context.Context, message TaskMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	confirmation, err := p.channel.PublishWithDeferredConfirmWithContext(ctx, "", TaskQueue, false, false, amqp.Publishing{ContentType: "application/json", DeliveryMode: amqp.Persistent, Body: body})
	if err != nil {
		return err
	}
	if ok, err := confirmation.WaitContext(ctx); err != nil || !ok {
		return fmt.Errorf("RabbitMQ 未确认消息: %w", err)
	}
	return nil
}

func (p *Publisher) Close() { _ = p.channel.Close(); _ = p.connection.Close() }
