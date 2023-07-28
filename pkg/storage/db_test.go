package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestStorage_SAGA(t *testing.T) {
	tests := []struct {
		name    string
		worker  Worker
		wantErr bool
	}{
		{
			name: "happy_path",
			worker: Worker{
				TaskID:   "my-task-id",
				WorkerID: "my-worker-id",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, table, cleanup := setup(t, ctx)
			t.Cleanup(cleanup)
			s := &Storage{
				client:    client,
				tableName: table,
			}
			err := s.Store(ctx, tt.worker)
			require.NoErrorf(t, err, "Store() error = %v", err)
			got, err := s.GetWorker(ctx, tt.worker.TaskID)
			require.NoErrorf(t, err, "failed to get item from dynamodb: %v", err)
			require.Equal(t, tt.worker, got)
			err = s.Delete(ctx, tt.worker.TaskID)
			require.NoErrorf(t, err, "failed to delete item from dynamodb: %v", err)
			_, err = s.GetWorker(ctx, tt.worker.TaskID)
			require.Error(t, err)
			require.ErrorContains(t, err, "no worker with matching taskID found")
		})
	}
}

func setup(t *testing.T, ctx context.Context) (*dynamodb.Client, string, func()) {
	dynamo, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:latest",
			ExposedPorts: []string{"8000/tcp"},
			Cmd:          []string{"-jar", "DynamoDBLocal.jar", "-inMemory"},
			WaitingFor:   wait.NewHostPortStrategy("8000"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	port, err := dynamo.MappedPort(ctx, "8000")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           fmt.Sprintf("http://localhost:%s", port),
			SigningRegion: "us-west-2",
		}, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	client := dynamodb.NewFromConfig(cfg)
	tableName := uuid.NewString()
	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("taskID"),
				AttributeType: "S",
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("taskID"),
				KeyType:       types.KeyTypeHash,
			},
		},
		TableName:   aws.String(tableName),
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatal(err)
	}
	return client, tableName, func() {
		if err := dynamo.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	}

}
