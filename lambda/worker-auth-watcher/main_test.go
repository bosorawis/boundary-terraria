package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func Test_convert(t *testing.T) {
	type args struct {
		e events.CloudwatchLogsData
	}
	tests := []struct {
		name    string
		args    args
		want    []registrationEvent
		wantErr bool
	}{
		{
			name: "happy_path",
			args: args{e: events.CloudwatchLogsData{
				Owner:     "609542363224",
				LogGroup:  "/fargate/service/boundary-workers",
				LogStream: "ecs/boundary-worker/649d1ff930e743e29f10c840f74947af",
				SubscriptionFilters: []string{
					"test",
				},
				MessageType: "DATA_MESSAGE",
				LogEvents: []events.CloudwatchLogsLogEvent{
					{
						ID:        "37579893129332945012746422377953780515727171655121502216",
						Timestamp: 1685140689012,
						Message:   "  Worker Auth Registration Request: pdZ5SAAebKa9DmnokkNu5EuBMK73v9mZUioiadcgEqFVk2hT9hjaFVvr2weyJnkSVj4Y3nneSwiDogmByuagSp2WJuE8Bq9gqHxi1PyK4fox4gWdSLwUNacYcBZFTzsWCVxVwqak9zt7U3sxyonmH2AXYgUiTSgkPrW55tN83Sd96WpWwDvhbJGb7DQq7ShfLCjTtJRmxLHUzqe9BqChMdMfpFqsfsMGSKBzq47BXpYnV3RMMPCo8xWYKHg6nbHivbb5mRXmdQCi11AnZeNqrd1cujffSSCeEStUKfV",
					},
				},
			}},
			want: []registrationEvent{
				{
					taskID: "649d1ff930e743e29f10c840f74947af",
					token:  "pdZ5SAAebKa9DmnokkNu5EuBMK73v9mZUioiadcgEqFVk2hT9hjaFVvr2weyJnkSVj4Y3nneSwiDogmByuagSp2WJuE8Bq9gqHxi1PyK4fox4gWdSLwUNacYcBZFTzsWCVxVwqak9zt7U3sxyonmH2AXYgUiTSgkPrW55tN83Sd96WpWwDvhbJGb7DQq7ShfLCjTtJRmxLHUzqe9BqChMdMfpFqsfsMGSKBzq47BXpYnV3RMMPCo8xWYKHg6nbHivbb5mRXmdQCi11AnZeNqrd1cujffSSCeEStUKfV",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convert(tt.args.e)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}
