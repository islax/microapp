package monitor

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
	"strings"
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
	go monitor.connector()

	return nil
}

func (monitor *sqsEventMonitor) connector() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	monitor.sqsSvc = sqs.New(sess)
	monitor.snsSvc = sns.New(sess)

	// Create a new request to list queues, first we will check to see if our required queue already exists
	listQueuesRequest := sqs.ListQueuesInput{}

	listQueueResults, err := monitor.sqsSvc.ListQueues(&listQueuesRequest)
	if err != nil {
		monitor.logger.Error().Err(err).Msg("Failed to get list of queues")
		return
	}

	var queueURL string
	for _, t := range listQueueResults.QueueUrls {
		// If one of the returned queue URL's contains the required name we need then break the loop
		if strings.Contains(*t, monitor.queueName) {
			queueURL = *t
			break
		}
	}

	// If, after checking existing queues, the URL is still empty then create the SQS queue.
	if len(queueURL) == 0 {
		// TODO: find a input attribute to create fifo queue
		createQueueInput := &sqs.CreateQueueInput{
			QueueName: &monitor.queueName,
		}

		createQueueResponse, err := monitor.sqsSvc.CreateQueue(createQueueInput)
		if err != nil {
			monitor.logger.Error().Err(err).Msg("Failed to create new queue")
			return
		}

		if createQueueResponse != nil {
			queueURL = *createQueueResponse.QueueUrl
		}
	}

	queueArn := convertQueueURLToArn(queueURL)

	protocolName := "sqs"
	var topicArn string

	listTopicsRequest := sns.ListTopicsInput{}

	// List all topics and loop through the results until we find a match
	allTopics, _ := monitor.snsSvc.ListTopics(&listTopicsRequest)

	for _, topic := range monitor.eventsToMonitor {

		routingKey := strings.ReplaceAll(topic, ".", "_")

		for _, t := range allTopics.Topics {
			if strings.Contains(*t.TopicArn, routingKey) {
				topicArn = *t.TopicArn
				break
			}
		}

		// If the required topic is found, then create the subscription
		if len(topicArn) > 0 {

			// subscribe SQS queue to a specific SNS topic by passing queue ARN
			_, err := monitor.snsSvc.Subscribe(&sns.SubscribeInput{
				TopicArn: &topicArn,
				Protocol: &protocolName,
				Endpoint: &queueArn,
			})
			if err != nil {
				monitor.logger.Error().Err(err).Msgf("Failed to subscribe queue")
				return
			}
		}

		// SNS cannot publish to an SQS topic unless it is given permission to do so.
		// TODO: find a better solution
		policyContent := "{\"Version\": \"2012-10-17\",  \"Id\": \"" + queueArn + "/SQSDefaultPolicy\",  \"Statement\": [    {     \"Sid\": \"Sid1580665629194\",      \"Effect\": \"Allow\",      \"Principal\": {        \"AWS\": \"*\"      },      \"Action\": \"SQS:SendMessage\",      \"Resource\": \"" + queueArn + "\",      \"Condition\": {        \"ArnEquals\": {         \"aws:SourceArn\": \"" + topicArn + "\"        }      }    }  ]}"

		attr := make(map[string]*string, 1)
		attr["Policy"] = &policyContent

		setQueueAttrInput := sqs.SetQueueAttributesInput{
			QueueUrl:   &queueURL,
			Attributes: attr,
		}

		_, err := monitor.sqsSvc.SetQueueAttributes(&setQueueAttrInput)
		if err != nil {
			monitor.logger.Error().Err(err).Msg("Failed to set queue attributes")
			return
		}

		go monitor.monitorQueueAndProcessMessages(queueURL, topic)
	}
}

// Awfully bad string replace code to convert a SQS queue URL to an Arn
// TODO: find better solution
func convertQueueURLToArn(inputURL string) string {
	queueArn := strings.Replace(strings.Replace(strings.Replace(inputURL, "https://sqs.", "arn:aws:sqs:", -1), ".amazonaws.com/", ":", -1), "/", ":", -1)
	return queueArn
}

func (monitor *sqsEventMonitor) monitorQueueAndProcessMessages(queueURL, topic string) {
	for {

		retrieveMessageRequest := sqs.ReceiveMessageInput{
			QueueUrl: &queueURL,
		}

		retrieveMessageResponse, _ := monitor.sqsSvc.ReceiveMessage(&retrieveMessageRequest)

		for _, message := range retrieveMessageResponse.Messages {

			monitor.logger.Debug().Msg("event received")

			snsMessage := snsMessage{}
			_ = json.Unmarshal([]byte(*message.Body), &snsMessage)

			command := &EventInfo{
				Payload:      snsMessage.Message,
				Name:         topic,
				CorelationID: snsMessage.MessageAttributes.CorrelationID.Value,
				RawToken:     snsMessage.MessageAttributes.XAuth.Value,
			}

			monitor.eventSignal <- command

			// delete the message to avoid duplication
			go monitor.deleteMessage(message, queueURL)
		}
	}
}

// message attribute
type attribute struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

// sns message
type snsMessage struct {
	Type              string `json:"Type"`
	MessageId         string `json:"MessageId"`
	TopicArn          string `json:"TopicArn"`
	Message           string `json:"Message"`
	Timestamp         string `json:"Timestamp"`
	SignatureVersion  string `json:"SignatureVersion"`
	Signature         string `json:"Signature"`
	MessageAttributes struct {
		XAuth         attribute `json:"X-Authorization"`
		CorrelationID attribute `json:"X-Correlation-ID"`
	} `json:"MessageAttributes"`
}

func (monitor *sqsEventMonitor) Start() {}

func (monitor *sqsEventMonitor) Stop() {}

func (monitor *sqsEventMonitor) deleteMessage(message *sqs.Message, queue string) {
	_, err := monitor.sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &queue,
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		monitor.logger.Error().Err(err).Msg("Failed to delete message")
	}
}
