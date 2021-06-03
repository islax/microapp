package monitor

import (
	"fmt"
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
		routingKey := strings.ReplaceAll(queue, ".", "_")
		queueUrlOutput, err := monitor.sqsSvc.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: &routingKey,
		})
		fmt.Println("queueUrlOutput", err, routingKey, queueUrlOutput)
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
		//fmt.Println("result", result, queue, result.Messages)
		//fmt.Println("result.Messages", result.Messages)
		go monitor.monitorQueueAndProcessMessages(result, queue)
	}
}

func (monitor *sqsEventMonitor) monitorQueueAndProcessMessages(sub *sqs.ReceiveMessageOutput, queue string) {
	for _, message := range sub.Messages {
		fmt.Println("message", message)
		monitor.logger.Debug().Msg("event received")

		var token string
		authorizationAttr, ok := message.MessageAttributes["X-Authorization"]
		fmt.Println("authorizationAttr", authorizationAttr, ok)
		if ok && authorizationAttr != nil {
			token = *authorizationAttr.StringValue
		}

		var correlationID string
		correlationIDAttr, ok := message.MessageAttributes["X-Correlation-ID"]
		if ok && correlationIDAttr != nil {
			correlationID = *correlationIDAttr.StringValue
		}

		command := &EventInfo{
			Payload:      *message.Body,
			Name:         queue,
			CorelationID: correlationID,
			RawToken:     token,
		}

		monitor.eventSignal <- command

		// delete the message to avoid duplication
		//go monitor.deleteMessage(message, queue)
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
