package projection

import (
	"context"
	"testing"

	"golang.org/x/text/language"

	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/handler"
	"github.com/zitadel/zitadel/internal/eventstore/repository"
	"github.com/zitadel/zitadel/internal/repository/instance"
	"github.com/zitadel/zitadel/internal/repository/org"
	"github.com/zitadel/zitadel/internal/repository/user"
)

func TestOrgMemberProjection_reduces(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "org MemberAddedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.MemberAddedEventType),
					org.AggregateType,
					[]byte(`{
					"userId": "user-id",
					"roles": ["role"]
				}`),
				), org.MemberAddedEventMapper),
			},
			reduce: (&orgMemberProjection{
				StatementHandler: getStatementHandlerWithFilters(
					user.NewHumanAddedEvent(context.Background(),
						&user.NewAggregate("user-id", "org1").Aggregate,
						"username1",
						"firstname1",
						"lastname1",
						"nickname1",
						"displayname1",
						language.German,
						domain.GenderMale,
						"email1",
						true,
					),
				)(t)}).reduceAdded,
			want: wantReduce{
				aggregateType:    org.AggregateType,
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.org_members3 (user_id, user_resource_owner, user_owner_removed, roles, creation_date, change_date, sequence, resource_owner, instance_id, owner_removed, org_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
							expectedArgs: []interface{}{
								"user-id",
								"org1",
								false,
								database.StringArray{"role"},
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								false,
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org MemberChangedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.MemberChangedEventType),
					org.AggregateType,
					[]byte(`{
					"userId": "user-id",
					"roles": ["role", "changed"]
				}`),
				), org.MemberChangedEventMapper),
			},
			reduce: (&orgMemberProjection{}).reduceChanged,
			want: wantReduce{
				aggregateType:    org.AggregateType,
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.org_members3 SET (roles, change_date, sequence) = ($1, $2, $3) WHERE (instance_id = $4) AND (user_id = $5) AND (org_id = $6)",
							expectedArgs: []interface{}{
								database.StringArray{"role", "changed"},
								anyArg{},
								uint64(15),
								"instance-id",
								"user-id",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org MemberCascadeRemovedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.MemberCascadeRemovedEventType),
					org.AggregateType,
					[]byte(`{
					"userId": "user-id"
				}`),
				), org.MemberCascadeRemovedEventMapper),
			},
			reduce: (&orgMemberProjection{}).reduceCascadeRemoved,
			want: wantReduce{
				aggregateType:    org.AggregateType,
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.org_members3 WHERE (instance_id = $1) AND (user_id = $2) AND (org_id = $3)",
							expectedArgs: []interface{}{
								"instance-id",
								"user-id",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org MemberRemovedType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.MemberRemovedEventType),
					org.AggregateType,
					[]byte(`{
					"userId": "user-id"
				}`),
				), org.MemberRemovedEventMapper),
			},
			reduce: (&orgMemberProjection{}).reduceRemoved,
			want: wantReduce{
				aggregateType:    org.AggregateType,
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.org_members3 WHERE (instance_id = $1) AND (user_id = $2) AND (org_id = $3)",
							expectedArgs: []interface{}{
								"instance-id",
								"user-id",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "user UserRemovedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(user.UserRemovedType),
					user.AggregateType,
					[]byte(`{}`),
				), user.UserRemovedEventMapper),
			},
			reduce: (&orgMemberProjection{}).reduceUserRemoved,
			want: wantReduce{
				aggregateType:    user.AggregateType,
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.org_members3 WHERE (instance_id = $1) AND (user_id = $2)",
							expectedArgs: []interface{}{
								"instance-id",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "org OrgRemovedEventType",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OrgRemovedEventType),
					org.AggregateType,
					[]byte(`{}`),
				), org.OrgRemovedEventMapper),
			},
			reduce: (&orgMemberProjection{}).reduceOrgRemoved,
			want: wantReduce{
				aggregateType:    org.AggregateType,
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.org_members3 SET (change_date, sequence, owner_removed) = ($1, $2, $3) WHERE (instance_id = $4) AND (resource_owner = $5)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								true,
								"instance-id",
								"agg-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.org_members3 SET (change_date, sequence, user_owner_removed) = ($1, $2, $3) WHERE (instance_id = $4) AND (user_resource_owner = $5)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								true,
								"instance-id",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceInstanceRemoved",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.InstanceRemovedEventType),
					instance.AggregateType,
					nil,
				), instance.InstanceRemovedEventMapper),
			},
			reduce: reduceInstanceRemovedHelper(MemberInstanceID),
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.org_members3 WHERE (instance_id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, OrgMemberProjectionTable, tt.want)
		})
	}
}
