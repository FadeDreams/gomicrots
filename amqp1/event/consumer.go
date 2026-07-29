// consumer.go
package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// "time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

const (
	mainExchange  = "logs_topic"
	dlx           = "logs_topic.dlx"
	retryQueue    = "logs_topic.retry"
	deadQueue     = "logs_topic.dead"
	maxRetries    = 5
	retryTTLMs    = 30000 // 30s backoff between retries
)

func logEvent(entry Payload) error {
	logHandlerURL := "http://loghandler:8083/log"

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	resp, err := http.Post(logHandlerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	fmt.Println("Log event posted successfully")
	return nil
}

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}
	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}
	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := declareExchange(channel); err != nil {
		return err
	}
	return declareDLXTopology(channel)
}

// declareDLXTopology sets up the retry queue (with TTL that requeues back to
// the main exchange on expiry) and the terminal dead queue.
func declareDLXTopology(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(dlx, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// Messages land here after a nack, sit for retryTTLMs, then get
	// dead-lettered *back* onto the main exchange for redelivery.
	_, err := channel.QueueDeclare(retryQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    mainExchange,
		"x-message-ttl":             int32(retryTTLMs),
	})
	if err != nil {
		return err
	}
	if err := channel.QueueBind(retryQueue, retryQueue, dlx, false, nil); err != nil {
		return err
	}

	// Terminal parking lot — nothing consumes this automatically.
	_, err = channel.QueueDeclare(deadQueue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	return channel.QueueBind(deadQueue, deadQueue, dlx, false, nil)
}

// Listen will listen to the queue and handle the messages
func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Qos(10, 0, false); err != nil { // cap in-flight msgs per consumer
		return err
	}

	q, err := declareQueue(ch) // now durable + DLX-bound, see below
	if err != nil {
		return err
	}

	for _, s := range topics {
		if err := ch.QueueBind(q.Name, s, mainExchange, false, nil); err != nil {
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", false, false, false, false, nil) // autoAck=false
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			d := d
			var payload Payload
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				log.Printf("malformed payload, dead-lettering: %v", err)
				// requeue=false with no retry queue route → basic.reject would
				// just drop it; instead route straight to deadQueue via a
				// direct publish so it's inspectable.
				_ = ch.Publish(dlx, deadQueue, false, false, amqp.Publishing{
					ContentType: "application/json",
					Body:        d.Body,
				})
				d.Ack(false) // remove from original queue; it's preserved in deadQueue
				continue
			}

			if err := handlePayload(payload); err != nil {
				if retryCount(d) >= maxRetries {
					log.Printf("giving up after %d retries, dead-lettering: %v", maxRetries, err)
					_ = ch.Publish(dlx, deadQueue, false, false, amqp.Publishing{
						ContentType: "application/json",
						Body:        d.Body,
					})
					d.Ack(false)
					continue
				}
				log.Printf("handlePayload failed, sending to retry queue: %v", err)
				d.Nack(false, false) // false,false → routed to DLX (retryQueue) via queue's x-dead-letter-exchange
				continue
			}

			d.Ack(false)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [%s, %s]\n", mainExchange, q.Name)
	<-forever
	return nil
}

// retryCount reads the x-death header RabbitMQ stamps on messages that have
// bounced through a DLX before, so we know how many attempts have happened.
func retryCount(d amqp.Delivery) int {
	xDeath, ok := d.Headers["x-death"].([]interface{})
	if !ok {
		return 0
	}
	count := 0
	for _, entry := range xDeath {
		if m, ok := entry.(amqp.Table); ok {
			if c, ok := m["count"].(int64); ok {
				count += int(c)
			}
		}
	}
	return count
}

func handlePayload(payload Payload) error {
	switch payload.Name {
	case "log", "event":
		return logEvent(payload)
	case "auth":
		// authenticate — TODO: no logic existed here originally
		return nil
	default:
		return logEvent(payload)
	}
}

// declareQueue replaces declareRandomQueue: an anonymous/exclusive queue
// can't carry retry semantics because it's deleted when the connection
// drops (any in-flight retry route breaks with it), and RabbitMQ won't let
// an existing queue change its DLX args without redeclaration. Use a
// stable, durable, named queue instead.
func declareQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		consumerQueueName(),
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		amqp.Table{
			"x-dead-letter-exchange": dlx,
		},
	)
}

func consumerQueueName() string {
	return "logs_topic.consumer"
}

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		mainExchange,
		"topic",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,
	)
}
