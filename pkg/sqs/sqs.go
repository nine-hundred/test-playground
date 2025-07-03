package sqs

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSConfig struct {
	Region      string `json:"region"`
	EndpointURL string `json:"endpoint_url"` // LocalStack용
	QueueName   string `json:"queue_name"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
}

type SQSClient struct {
	client   *sqs.SQS
	queueURL string
}

type NotificationMessage struct {
	EventType string `json:"event_type"`
	TodoID    int64  `json:"todo_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func NewSQSClient(config SQSConfig) (*SQSClient, error) {
	// AWS 세션 생성
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(config.Region),
		Endpoint:         aws.String(config.EndpointURL), // LocalStack용
		Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		S3ForcePathStyle: aws.Bool(true), // LocalStack 호환성
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// SQS 클라이언트 생성
	sqsClient := sqs.New(sess)

	// 큐 URL 가져오기 (큐가 없으면 생성)
	queueURL, err := getOrCreateQueue(sqsClient, config.QueueName)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create queue: %w", err)
	}

	return &SQSClient{
		client:   sqsClient,
		queueURL: queueURL,
	}, nil
}

func getOrCreateQueue(client *sqs.SQS, queueName string) (string, error) {
	// 기존 큐 URL 가져오기 시도
	getQueueURLInput := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}

	result, err := client.GetQueueUrl(getQueueURLInput)
	if err == nil {
		return *result.QueueUrl, nil
	}

	// 큐가 없으면 생성
	createQueueInput := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]*string{
			"VisibilityTimeoutSeconds": aws.String("300"),     // 5분
			"MessageRetentionPeriod":   aws.String("1209600"), // 14일
		},
	}

	createResult, err := client.CreateQueue(createQueueInput)
	if err != nil {
		return "", fmt.Errorf("failed to create queue: %w", err)
	}

	return *createResult.QueueUrl, nil
}

func (s *SQSClient) SendMessage(message NotificationMessage) error {
	// JSON으로 직렬화
	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// SQS 메시지 전송
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(messageBody)),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"EventType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(message.EventType),
			},
			"TodoID": {
				DataType:    aws.String("Number"),
				StringValue: aws.String(fmt.Sprintf("%d", message.TodoID)),
			},
		},
	}

	_, err = s.client.SendMessage(input)
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	return nil
}

func (s *SQSClient) GetQueueAttributes() (map[string]*string, error) {
	input := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(s.queueURL),
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
			aws.String("ApproximateNumberOfMessagesNotVisible"),
		},
	}

	result, err := s.client.GetQueueAttributes(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	return result.Attributes, nil
}
