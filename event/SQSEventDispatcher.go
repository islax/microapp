package event

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

// SQSEventDispatcher is an event dispatcher that sends event to the SQS
type SQSEventDispatcher struct {
	logger                     *zerolog.Logger
	snsSvc                     *sns.SNS
	availableTopicARNs         []string
	sendChannel                chan *queueCommand
	retryMessagePublishChannel chan *retryCommand
	retryTopicCreateChannel    chan *retryCommand
	connectionMutex            sync.Mutex
}

// NewSQSEventDispatcher create and returns a new SQSEventDispatcher
func NewSQSEventDispatcher(logger *zerolog.Logger) (*SQSEventDispatcher, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ctxLogger := logger.With().Str("module", "SQSEventDispatcher").Logger()

	svc := sns.New(sess)

	availableTopics, _ := svc.ListTopics(nil)

	availableTopicARNs := make([]string, len(availableTopics.Topics))

	for i, t := range availableTopics.Topics {
		availableTopicARNs[i] = *t.TopicArn
	}

	dispatcher := &SQSEventDispatcher{
		logger:                     &ctxLogger,
		snsSvc:                     svc,
		availableTopicARNs:         availableTopicARNs,
		sendChannel:                make(chan *queueCommand, 200),
		retryMessagePublishChannel: make(chan *retryCommand, 200),
		retryTopicCreateChannel:    make(chan *retryCommand, 200),
	}

	go dispatcher.start()

	return dispatcher, nil
}

func (eventDispatcher *SQSEventDispatcher) start() {

	for {

		var command *queueCommand
		var retryMessagePublishCount int
		var retryTopicCreateCount int

		// Ensure that connection process is not going on
		eventDispatcher.connectionMutex.Lock()
		eventDispatcher.connectionMutex.Unlock()

		select {
		case commandFromSendChannel := <-eventDispatcher.sendChannel:
			command = commandFromSendChannel
		case commandFromMessagePublishRetryChannel := <-eventDispatcher.retryMessagePublishChannel:
			command = commandFromMessagePublishRetryChannel.command
			retryMessagePublishCount = commandFromMessagePublishRetryChannel.retryCount
		case commandFromTopicCreateRetryChannel := <-eventDispatcher.retryTopicCreateChannel:
			command = commandFromTopicCreateRetryChannel.command
			retryTopicCreateCount = commandFromTopicCreateRetryChannel.retryCount
		}

		routingKey := strings.ReplaceAll(command.topic, ".", "_")

		var body []byte
		var err error

		body, isByteMessage := command.payload.([]byte)
		if !isByteMessage {
			body, err = json.Marshal(command.payload)
			if err != nil {
				eventDispatcher.logger.Error().Msg("Failed to convert payload to JSON" + ": " + err.Error())
				continue
			}
		}

		var requiredTopicArn string

		for _, t := range eventDispatcher.availableTopicARNs {
			if strings.Contains(t, routingKey) {
				requiredTopicArn = t
			}
		}

		if len(requiredTopicArn) == 0 {

			topicOutput, err := eventDispatcher.snsSvc.CreateTopic(&sns.CreateTopicInput{Name: &routingKey})
			if err != nil {
				if retryTopicCreateCount < 3 {
					eventDispatcher.logger.Warn().Msg("Failed to create topic. Trying again ... Error: " + err.Error())
					go eventDispatcher.retryCreateTopic(retryTopicCreateCount+1, command)
				} else {
					eventDispatcher.logger.Error().Msg("Failed to create topic" + ": " + err.Error())
					continue
				}
			}

			requiredTopicArn = *topicOutput.TopicArn

			eventDispatcher.availableTopicARNs = append(eventDispatcher.availableTopicARNs, requiredTopicArn)
		}

		// message attribute values cannot be empty so passing random values when empty
		// TODO: find better solution
		token := "random-string"
		if len(command.token) > 0 {
			token = command.token
		}

		correlationID := "00000000-0000-0000-0000-000000000000"
		if len(command.correlationID) > 0 {
			correlationID = command.correlationID
		}

		_, err = eventDispatcher.snsSvc.Publish(&sns.PublishInput{
			Message:  aws.String(string(body)),
			TopicArn: aws.String(requiredTopicArn),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"X-Authorization": {
					DataType:    aws.String("String"),
					StringValue: aws.String(token),
				},
				"X-Correlation-ID": {
					DataType:    aws.String("String"),
					StringValue: aws.String(correlationID),
				},
			},
		})

		if err != nil {
			if retryMessagePublishCount < 3 {
				eventDispatcher.logger.Warn().Msg("Failed to publish message. Trying again ... Error: " + err.Error())
				go eventDispatcher.retryCreateTopic(retryMessagePublishCount+1, command)
			} else {
				eventDispatcher.logger.Error().Msg("Failed to publish message" + ": " + err.Error())
				continue
			}
		}

		eventDispatcher.logger.Trace().Msgf("message published to topic %s", command.topic)
	}
}

// DispatchEvent dispatches events to the message queue
func (eventDispatcher *SQSEventDispatcher) DispatchEvent(token string, correlationID string, topic string, payload interface{}) {
	eventDispatcher.sendChannel <- &queueCommand{token: token, topic: topic, correlationID: correlationID, payload: payload}
}

func (eventDispatcher *SQSEventDispatcher) retryMessagePublish(retryCount int, command *queueCommand) {
	eventDispatcher.retryMessagePublishChannel <- &retryCommand{retryCount: retryCount, command: command}
}

func (eventDispatcher *SQSEventDispatcher) retryCreateTopic(retryCount int, command *queueCommand) {
	eventDispatcher.retryTopicCreateChannel <- &retryCommand{retryCount: retryCount, command: command}
}
