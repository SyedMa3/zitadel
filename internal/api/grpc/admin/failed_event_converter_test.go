package admin

import (
	"context"
	"testing"
	"time"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/test"
	"github.com/zitadel/zitadel/internal/view/model"
	admin_pb "github.com/zitadel/zitadel/pkg/grpc/admin"
)

func TestFailedEventsToPbFields(t *testing.T) {
	type args struct {
		failedEvents []*model.FailedEvent
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "all fields",
			args: args{
				failedEvents: []*model.FailedEvent{
					{
						Database:       "admin",
						ViewName:       "users",
						FailedSequence: 456,
						FailureCount:   5,
						LastFailed:     time.Now(),
						ErrMsg:         "some error",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FailedEventsViewToPb(tt.args.failedEvents)
			for _, g := range got {
				test.AssertFieldsMapped(t, g)
			}
		})
	}
}

func TestFailedEventToPbFields(t *testing.T) {
	type args struct {
		failedEvent *model.FailedEvent
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"all fields",
			args{
				failedEvent: &model.FailedEvent{
					Database:       "admin",
					ViewName:       "users",
					FailedSequence: 456,
					FailureCount:   5,
					LastFailed:     time.Now(),
					ErrMsg:         "some error",
				},
			},
		},
	}
	for _, tt := range tests {
		converted := FailedEventViewToPb(tt.args.failedEvent)
		test.AssertFieldsMapped(t, converted)
	}
}

func TestRemoveFailedEventRequestToModelFields(t *testing.T) {
	type args struct {
		ctx context.Context
		req *admin_pb.RemoveFailedEventRequest
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"all fields",
			args{
				ctx: authz.WithInstanceID(context.Background(), "instanceID"),
				req: &admin_pb.RemoveFailedEventRequest{
					Database:       "admin",
					ViewName:       "users",
					FailedSequence: 456,
				},
			},
		},
	}
	for _, tt := range tests {
		converted := RemoveFailedEventRequestToModel(tt.args.ctx, tt.args.req)
		test.AssertFieldsMapped(t, converted, "FailureCount", "LastFailed", "ErrMsg")
	}
}
