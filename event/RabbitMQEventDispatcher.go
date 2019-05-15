package event

import (
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// RabbitMQEventDispatcher is an event dispatcher that sends event to the RabbitMQ Exchange
type RabbitMQEventDispatcher struct {
	logger       *log.Logger
	exchangeName string
}

// NewRabbitMQEventDispatcher create and returns a new RabbitMQEventDispatcher
func NewRabbitMQEventDispatcher(logger *log.Logger) *RabbitMQEventDispatcher {
	return &RabbitMQEventDispatcher{logger: logger, exchangeName: "isla_Exchange"}
}

// func logError (message string, err error, contextLogger *log.Logger) {
// 	if err != nil {
// 		contextLogger.Error(message + ": " + err.Error())
// 		return
// 	}
// }

// DispatchEvent dispatches events to the message queue
func (eventDispatcher *RabbitMQEventDispatcher) DispatchEvent(token string, topic string, payload interface{}) {
	contextLogger := eventDispatcher.logger.WithFields(log.Fields{
		"module": "RabbitMQEventDispatcher",
	})

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		contextLogger.Error("Failed to conntect to RabbitMQ" + ": " + err.Error())
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		contextLogger.Error("Failed to open a Channel" + ": " + err.Error())
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(eventDispatcher.exchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		contextLogger.Error("ailed to declare an exchange" + ": " + err.Error())
		return
	}

	routingKey := strings.ReplaceAll(topic, "_", ".")
	body, err := json.Marshal(payload)
	if err != nil {
		contextLogger.Error("Failed to convert payload to JSON" + ": " + err.Error())
		return
	}

	err = ch.Publish(
		eventDispatcher.exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	if err != nil {
		contextLogger.Error("Failed to publish to an Exchange" + ": " + err.Error())
		return
	}

	contextLogger.Info("Sent message to queue")
}
