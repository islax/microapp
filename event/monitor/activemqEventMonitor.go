package monitor

import (
	"net"
	"os"
	"strings"

	"github.com/go-stomp/stomp"
	"github.com/rs/zerolog"
)

type activeMQEventMonitor struct {
	logger          *zerolog.Logger
	queueName       string
	eventSignal     chan *EventInfo
	eventsToMonitor []string
	connection      *stomp.Conn
}

func (monitor *activeMQEventMonitor) initialize(eventsToMonitor []string) error {
	monitor.eventsToMonitor = eventsToMonitor
	go monitor.activemqConnector()

	return nil
}

func (monitor *activeMQEventMonitor) activemqConnector() {
	conn, err := stomp.Dial("tcp", getQueueHostPort(), stomp.Options{HeartBeat: "1000,0"}) //TODO remove hardcoded values
	if err != nil {
		monitor.logger.Error().Err(err).Msg("Failed to connect to activemq.")
		return
	}

	monitor.connection = conn
	for _, queue := range monitor.eventsToMonitor {
		sub, err := conn.Subscribe(queue, stomp.AckClient)
		if err != nil {
			monitor.logger.Error().Err(err).Msg("Failed to subscribe to queue")
			return
		}
		go monitor.monitorQueueAndProcessMessages(sub, queue)
	}
}

func (monitor *activeMQEventMonitor) monitorQueueAndProcessMessages(sub *stomp.Subscription, queue string) {
	for message := range sub.C {
		monitor.logger.Debug().Msg("event received")

		if message.Err != nil {
			monitor.logger.Error().Err(message.Err).Msg("Error in receiving message")
			continue
		}

		authorizationHeader := message.Header.Get("X-Authorization")
		correlationIDHeader := message.Header.Get("X-Correlation-ID")
		payload := string(message.Body)

		command := &EventInfo{
			Payload:      payload,
			Name:         strings.TrimPrefix(message.Destination, "/queue/"),
			CorelationID: correlationIDHeader,
			RawToken:     authorizationHeader,
		}

		monitor.eventSignal <- command

		// acknowledge the message
		err := monitor.connection.Ack(message)
		if err != nil {
			monitor.logger.Error().Err(message.Err).Msg("Error in message ack")
			continue
		}
	}
}

func (monitor *activeMQEventMonitor) Start() {
}

func (monitor *activeMQEventMonitor) Stop() {
}

func getQueueHostPort() string {
	var queueHost string
	queueHost = os.Getenv("ISLA_QUEUE_HOST")
	if len(queueHost) == 0 {
		queueHost = "localhost"
	}

	var queuePort string
	queuePort = os.Getenv("ISLA_QUEUE_PORT")
	if len(queueHost) == 0 {
		queuePort = "61616"
	}

	return net.JoinHostPort(queueHost, queuePort)
}