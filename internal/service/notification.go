package service

import (
	"fmt"
	"integration-test-example/internal/model"
	"integration-test-example/pkg/sqs"
	"time"
)

type NotificationService struct {
	sqsClient *sqs.SQSClient
}

func NewNotificationService(sqsClient *sqs.SQSClient) *NotificationService {
	return &NotificationService{
		sqsClient: sqsClient,
	}
}

func (n *NotificationService) SendTodoCompletedNotification(todo *model.Todo) error {
	message := sqs.NotificationMessage{
		EventType: "todo_completed",
		TodoID:    todo.ID,
		Title:     todo.Title,
		Message:   fmt.Sprintf("🎉 축하합니다! '%s' 할 일을 완료했습니다!", todo.Title),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := n.sqsClient.SendMessage(message); err != nil {
		return fmt.Errorf("failed to send todo completed notification: %w", err)
	}

	return nil
}
