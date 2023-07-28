package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	ErrNotfound = fmt.Errorf("no worker with matching taskID found")
)

type dynamoClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

type Storage struct {
	client    dynamoClient
	tableName string
}

type Worker struct {
	TaskID   string `dynamodbav:"taskID"`
	WorkerID string `dynamodbav:"workerID"`
}

func (s *Storage) Store(ctx context.Context, w Worker) error {
	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"taskID":   &types.AttributeValueMemberS{Value: w.TaskID},
			"workerID": &types.AttributeValueMemberS{Value: w.WorkerID},
		},
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return fmt.Errorf("failed to save worker to dynamodb: %w", err)
	}
	return nil
}

func (s *Storage) GetWorker(ctx context.Context, taskID string) (Worker, error) {
	response, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"taskID": &types.AttributeValueMemberS{
				Value: taskID,
			},
		},
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return Worker{}, fmt.Errorf("failed to save worker to dynamodb: %w", err)
	}
	if len(response.Item) == 0 {
		return Worker{}, fmt.Errorf("%w: %s", ErrNotfound, taskID)
	}
	var gotWorker Worker
	err = attributevalue.UnmarshalMap(response.Item, &gotWorker)
	if err != nil {
		return Worker{}, fmt.Errorf("failed to unmarshal worker: %w", err)
	}
	return gotWorker, nil
}

func (s *Storage) Delete(ctx context.Context, taskID string) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"taskID": &types.AttributeValueMemberS{Value: taskID},
		},
		TableName: aws.String(s.tableName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete item from dynamoDB: %w", err)
	}
	return nil
}

func New(client dynamoClient, tableName string) *Storage {
	return &Storage{
		client:    client,
		tableName: tableName,
	}
}
