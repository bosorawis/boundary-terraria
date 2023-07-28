package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/stretchr/testify/require"
)

func Test_parseEvent(t *testing.T) {
	type args struct {
		event events.ECSContainerInstanceEvent
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "happy_path",
			args: args{
				event: events.ECSContainerInstanceEvent{
					Resources: []string{
						"arn:aws:ecs:us-west-2:123456789012:task/boundary-workers-cluster/7348f796ddc54fab8013b3cac54d9f76",
						"arn:aws:ecs:us-west-2:123456789012:task/boundary-workers-cluster/dd48f796ddc54fab8013b3cac54d9f76",
						"arn:aws:ecs:us-west-2:123456789012:task/boundary-workers-cluster/3128f796ddc54fab8013b3cac54d9f76",
					},
				},
			},
			want: []string{"7348f796ddc54fab8013b3cac54d9f76", "dd48f796ddc54fab8013b3cac54d9f76", "3128f796ddc54fab8013b3cac54d9f76"},
		},
		{
			name: "invalid_arn",
			args: args{
				event: events.ECSContainerInstanceEvent{
					Resources: []string{
						"hello",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_resource",
			args: args{
				event: events.ECSContainerInstanceEvent{
					Resources: []string{
						"arn:aws:ecs:us-west-2:123456789012:task/boundary-workers-cluster/",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "resource_too_short",
			args: args{
				event: events.ECSContainerInstanceEvent{
					Resources: []string{
						"arn:aws:ecs:us-west-2:123456789012:task",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseForTaskID(tt.args.event)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestHello(t *testing.T) {
	s := "arn:aws:ecs:us-west-2:123456789012:task/boundary-workers-cluster/3128f796ddc54fab8013b3cac54d9f76"
	str, _ := arn.Parse(s)

	splitted := strings.Split(str.Resource, "/")
	fmt.Println(splitted[len(splitted)-1])
}
