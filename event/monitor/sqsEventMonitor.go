package monitor

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
)

type sqsEventMonitor struct {
	logger          *zerolog.Logger
	queueName       string
	eventSignal     chan *EventInfo
	eventsToMonitor []string
	sqsSvc          *sqs.SQS
	snsSvc          *sns.SNS
}

func (monitor *sqsEventMonitor) initialize(eventsToMonitor []string) error {
	eventsToMonitorforsqs := make([]string, len(eventsToMonitor))
	for idx, em := range eventsToMonitor {
		eventsToMonitorforsqs[idx] = strings.ReplaceAll(em, ".", "")
	}
	monitor.eventsToMonitor = eventsToMonitor
	go monitor.sqsConnector()

	return nil
}

func (monitor *sqsEventMonitor) sqsConnector() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	monitor.sqsSvc = sqs.New(sess)
	monitor.snsSvc = sns.New(sess)

	for _, queue := range monitor.eventsToMonitor {

		queueUrlOutput, err := monitor.sqsSvc.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: &queue,
		})

		if err != nil {
			continue
		}

		result, err := monitor.sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
			MaxNumberOfMessages: aws.Int64(1),
			QueueUrl:            queueUrlOutput.QueueUrl,
			MessageAttributeNames: []*string{
				aws.String("X-Authorization"),
				aws.String("X-Correlation-ID"),
			},
		})
		if err != nil {
			monitor.logger.Error().Err(err).Msg("Failed to subscribe to queue")
			continue
		}

		go monitor.monitorQueueAndProcessMessages(result, *queueUrlOutput.QueueUrl)
	}
}

func (monitor *sqsEventMonitor) monitorQueueAndProcessMessages(sub *sqs.ReceiveMessageOutput, queue string) {
	for _, message := range sub.Messages {

		monitor.logger.Debug().Msg("event received")

		var token string
		authorizationAttr, ok := message.Attributes["X-Authorization"]
		if ok && authorizationAttr != nil {
			token = *authorizationAttr
		}

		var correlationID string
		correlationIDAttr, ok := message.Attributes["X-Correlation-ID"]
		if ok && correlationIDAttr != nil {
			correlationID = *correlationIDAttr
		}

		command := &EventInfo{
			Payload:      *message.Body,
			Name:         queue,
			CorelationID: correlationID,
			RawToken:     token,
		}

		monitor.eventSignal <- command

		// delete the message to avoid duplication
		go monitor.deleteMessage(message, queue)
	}
}

func (monitor *sqsEventMonitor) Start() {
}

func (monitor *sqsEventMonitor) Stop() {
}

func (monitor *sqsEventMonitor) deleteMessage(message *sqs.Message, queue string) {
	_, err := monitor.sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &queue,
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		monitor.logger.Error().Err(err).Msg("Error in message ack")
	}
}
