package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/dihmuzikien/boundary-sandbox/pkg/storage"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authmethods"
	"github.com/hashicorp/boundary/api/workers"
)

func (a *app) HandleRequest(ctx context.Context, event events.CloudwatchLogsEvent) (string, error) {
	parsed, err := event.AWSLogs.Parse()
	if err != nil {
		return "", fmt.Errorf("failed to parse CloudwatchLogsEvent: %w", err)
	}
	registEvents, err := convert(parsed)
	if err != nil {
		return "", fmt.Errorf("failed to convert input to events: %w", err)
	}
	err = a.register(ctx, registEvents)
	if err != nil {
		return "", fmt.Errorf("failed to register workers: %w", err)
	}
	return "SUCCEED", nil
}

type registrationEvent struct {
	taskID string
	token  string
}

func convert(e events.CloudwatchLogsData) ([]registrationEvent, error) {
	splitStream := strings.Split(e.LogStream, "/")
	if len(splitStream) != 3 {
		return []registrationEvent{}, fmt.Errorf("invalid logstream format. expects 2 '/' but got: %s", e.LogStream)
	}
	taskID := splitStream[2]
	var result []registrationEvent
	for _, event := range e.LogEvents {
		items := strings.Split(event.Message, ": ")
		if len(items) != 2 {
			continue
		}
		result = append(result, registrationEvent{
			taskID: taskID,
			token:  items[1],
		})
	}

	return result, nil
}

func (a *app) register(ctx context.Context, events []registrationEvent) error {
	for _, e := range events {
		w, err := a.boundaryWorkerClient.CreateWorkerLed(ctx, e.token, "global")
		if err != nil {
			return fmt.Errorf("failed to register worker %s: %w", e.taskID, err)
		}
		err = a.storage.Store(ctx, storage.Worker{
			TaskID:   e.taskID,
			WorkerID: w.GetItem().Id,
		})
		if err != nil {
			return fmt.Errorf("failed to store worker information task: %s worker: %s error: %w", e.taskID, w.GetItem().Id, err)
		}
	}
	return nil
}

type app struct {
	boundaryWorkerClient *workers.Client
	storage              *storage.Storage
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	client, err := api.NewClient(nil)
	if err != nil {
		log.Fatalf("cannot instantiate boundary client: %v", err)
	}
	err = client.SetAddr(os.Getenv("CLUSTER_URL"))
	if err != nil {
		log.Fatalf("failed to set CLUSTER_URL to %s: %v", os.Getenv("CLUSTER_URL"), err)
	}
	credentials := map[string]interface{}{
		"login_name": os.Getenv("BOUNDARY_USERNAME"),
		"password":   os.Getenv("BOUNDARY_PASSWORD"),
	}
	amClient := authmethods.NewClient(client)
	defer cancel()
	authResult, err := amClient.Authenticate(ctx, os.Getenv("BOUNDARY_AUTH_MATHOD_ID"), "login", credentials)
	if err != nil {
		log.Fatalf("failed to authenticate: %v", err)
	}
	client.SetToken(fmt.Sprint(authResult.Attributes["token"]))

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load aws config: %v", err)
	}
	dynamoClient := dynamodb.NewFromConfig(cfg)
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		log.Fatalf("TABLE_NAME environment var must be set")
	}
	sge := storage.New(dynamoClient, tableName)
	a := app{
		storage:              sge,
		boundaryWorkerClient: workers.NewClient(client),
	}
	lambda.Start(a.HandleRequest)
}
