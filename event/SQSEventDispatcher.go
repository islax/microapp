package event

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/rs/zerolog"
)

// SQSEventDispatcher is an event dispatcher that sends event to the RabbitMQ Exchange
type SQSEventDispatcher struct {
	logger          *zerolog.Logger
	sendChannel     chan *queueCommand
	retryChannel    chan *retryCommand
	sqsSvc          *sqs.SQS
	snsSvc          *sns.SNS
	connectionMutex sync.Mutex
}

// NewSQSEventDispatcher create and returns a new SQSEventDispatcher
func NewSQSEventDispatcher(logger *zerolog.Logger) (*SQSEventDispatcher, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("us-west-2")},
		SharedConfigState: session.SharedConfigEnable,
	}))

	ctxLogger := logger.With().Str("module", "SQSEventDispatcher").Logger()

	dispatcher := &SQSEventDispatcher{
		logger:       &ctxLogger,
		sqsSvc:       sqs.New(sess),
		snsSvc:       sns.New(sess),
		sendChannel:  make(chan *queueCommand, 200),
		retryChannel: make(chan *retryCommand, 200),
	}

	go dispatcher.start()

	return dispatcher, nil
}

func (eventDispatcher *SQSEventDispatcher) start() {

	for {
		var command *queueCommand
		var retryCount int

		// Ensure that connection process is not going on
		eventDispatcher.connectionMutex.Lock()
		eventDispatcher.connectionMutex.Unlock()

		select {
		case commandFromSendChannel := <-eventDispatcher.sendChannel:
			command = commandFromSendChannel
		case commandFromRetryChannel := <-eventDispatcher.retryChannel:
			command = commandFromRetryChannel.command
			retryCount = commandFromRetryChannel.retryCount
		}
		fmt.Printf("command: %+v", command)
		routingKey := strings.ReplaceAll(command.topic, "_", ".")
		routingKey = strings.ReplaceAll(routingKey, ".", "")
		var body []byte
		var err error

		body, isByteMessage := command.payload.([]byte)
		if !isByteMessage {
			body, err = json.Marshal(command.payload)
			if err != nil {
				eventDispatcher.logger.Error().Msg("Failed to convert payload to JSON" + ": " + err.Error())
			}
		}

		var queueUrl *string
		fmt.Println("routingKey", routingKey)
		queueUrlOutput, errGetQueueUrl := eventDispatcher.sqsSvc.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: &routingKey,
		})
		fmt.Printf("queueUrlOutput,errGetQueueUrl: %+v %+v\n", queueUrlOutput, errGetQueueUrl)
		// if queue is not exists, create one
		if errGetQueueUrl != nil {
			createQueueOutput, err := eventDispatcher.sqsSvc.CreateQueue(&sqs.CreateQueueInput{
				QueueName: &routingKey,
			})
			fmt.Printf("createQueueOutput, err: %+v %+v\n", createQueueOutput, err)
			queueUrl = createQueueOutput.QueueUrl

			_, err = eventDispatcher.snsSvc.CreateTopic(&sns.CreateTopicInput{Name: &routingKey})
			fmt.Printf("CreateTopic: %+v\n", err)
			errGetQueueUrl = err
		} else {
			queueUrl = queueUrlOutput.QueueUrl
		}
		fmt.Println(errGetQueueUrl)
		if errGetQueueUrl == nil {

			listTopicsRequest := sns.ListTopicsInput{}

			// List all topics and loop through the results until we find a match
			allTopics, _ := eventDispatcher.snsSvc.ListTopics(&listTopicsRequest)

			var topicARN string
			for _, t := range allTopics.Topics {
				if strings.Contains(*t.TopicArn, routingKey) {
					topicARN = *t.TopicArn
					break
				}
			}
			//topicARN = "arn:aws:sns:us-west-2:104722656260:user_updated"
			protocol := "sqs"

			queueAttrs, _ := eventDispatcher.sqsSvc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
				AttributeNames: aws.StringSlice([]string{"QueueArn"}),
				QueueUrl:       queueUrl,
			})

			queueARN, _ := queueAttrs.Attributes["QueueArn"]

			subscribeQueueInput := sns.SubscribeInput{
				TopicArn: &topicARN,
				Protocol: &protocol,
				Endpoint: queueARN,
			}

			_, err := eventDispatcher.snsSvc.Subscribe(&subscribeQueueInput)
			fmt.Println(err)
			if err != nil {
				continue
			}

			_, err = eventDispatcher.sqsSvc.SendMessage(&sqs.SendMessageInput{
				DelaySeconds: aws.Int64(10),
				MessageAttributes: map[string]*sqs.MessageAttributeValue{
					"X-Authorization": {
						DataType:    aws.String("String"),
						StringValue: aws.String(command.token),
					},
					//"X-Correlation-ID": {
					//	DataType:    aws.String("String"),
					//	StringValue: aws.String(command.correlationID),
					//},
				},
				MessageBody: aws.String(string(body)),
				QueueUrl:    queueUrl,
			})

			if err != nil {
				if retryCount < 3 {
					eventDispatcher.logger.Warn().Msg("Publish to queue failed. Trying again ... Error: " + err.Error())

					go func(command *queueCommand, retryCount int) {
						time.Sleep(time.Second)
						eventDispatcher.retryChannel <- &retryCommand{retryCount: retryCount, command: command}
					}(command, retryCount+1)
				} else {
					eventDispatcher.logger.Error().Msg("Failed to publish to an Exchange" + ": " + err.Error())
				}
			} else {
				eventDispatcher.logger.Trace().Msg("Sent message to queue")
			}
		}
	}
}

// DispatchEvent dispatches events to the message queue
func (eventDispatcher *SQSEventDispatcher) DispatchEvent(token string, correlationID string, topic string, payload interface{}) {
	fmt.Println("sending to sendChannel")
	eventDispatcher.sendChannel <- &queueCommand{token: token, topic: topic, payload: payload}
	fmt.Println("sent to sendChannel")
}

func convertQueueURLToARN(inputURL string) string {
	queueARN := strings.Replace(strings.Replace(strings.Replace(inputURL, "https://sqs.", "arn:aws:sqs:", -1), ".amazonaws.com/", ":", -1), "/", ":", -1)
	return queueARN
}
