package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/dihmuzikien/boundary-sandbox/pkg/storage"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/authmethods"
	"github.com/hashicorp/boundary/api/workers"
)

type app struct {
	boundaryWorkerClient *workers.Client
	storage              *storage.Storage
}

func (a *app) HandleRequest(ctx context.Context, event events.ECSContainerInstanceEvent) (string, error) {
	taskIDs, err := parseForTaskID(event)
	if err != nil {
		return "", fmt.Errorf("failed to parse for taskIDs: %w", err)
	}
	for _, id := range taskIDs {
		w, err := a.storage.GetWorker(ctx, id)
		if err != nil {
			if errors.Is(err, storage.ErrNotfound) {
				continue
			}
			return "", fmt.Errorf("failed to fetch worker info with taskID %s: %w", id, err)
		}
		_, err = a.boundaryWorkerClient.Delete(ctx, w.WorkerID)
		if err != nil {
			return "", fmt.Errorf("failed to delete worker with ID %s from boundary: %w", w.WorkerID, err)
		}
		err = a.storage.Delete(ctx, w.TaskID)
		if err != nil {
			return "", fmt.Errorf("failed to delete worker %v from storage: %w", w, err)
		}
		fmt.Printf("successfully clean up worker %v\n", w)
	}
	return "SUCCEED", nil
}

func parseForTaskID(event events.ECSContainerInstanceEvent) ([]string, error) {
	var results []string
	for _, r := range event.Resources {
		taskARN, err := arn.Parse(r)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ARN %s: %w", r, err)
		}
		splitted := strings.Split(taskARN.Resource, "/")
		if len(splitted) < 3 {
			return nil, fmt.Errorf("")
		}
		results = append(results, splitted[len(splitted)-1])
	}

	return results, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
