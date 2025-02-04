//go:build integration

package user_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/muhlemmer/gu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitadel/zitadel/internal/integration"
	object "github.com/zitadel/zitadel/pkg/grpc/object/v2alpha"
	user "github.com/zitadel/zitadel/pkg/grpc/user/v2alpha"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func createHumanUser(t *testing.T) *user.AddHumanUserResponse {
	resp, err := Client.AddHumanUser(CTX, &user.AddHumanUserRequest{
		Organisation: &object.Organisation{
			Org: &object.Organisation_OrgId{
				OrgId: Tester.Organisation.ID,
			},
		},
		Profile: &user.SetHumanProfile{
			FirstName: "Mickey",
			LastName:  "Mouse",
		},
		Email: &user.SetHumanEmail{
			Email: fmt.Sprintf("%d@mouse.com", time.Now().UnixNano()),
			Verification: &user.SetHumanEmail_ReturnCode{
				ReturnCode: &user.ReturnEmailVerificationCode{},
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetUserId())
	return resp
}

func TestServer_SetEmail(t *testing.T) {
	userID := createHumanUser(t).GetUserId()

	tests := []struct {
		name    string
		req     *user.SetEmailRequest
		want    *user.SetEmailResponse
		wantErr bool
	}{
		{
			name: "default verfication",
			req: &user.SetEmailRequest{
				UserId: userID,
				Email:  "default-verifier@mouse.com",
			},
			want: &user.SetEmailResponse{
				Details: &object.Details{
					Sequence:      1,
					ChangeDate:    timestamppb.Now(),
					ResourceOwner: Tester.Organisation.ID,
				},
			},
		},
		{
			name: "custom url template",
			req: &user.SetEmailRequest{
				UserId: userID,
				Email:  "custom-url@mouse.com",
				Verification: &user.SetEmailRequest_SendCode{
					SendCode: &user.SendEmailVerificationCode{
						UrlTemplate: gu.Ptr("https://example.com/email/verify?userID={{.UserID}}&code={{.Code}}&orgID={{.OrgID}}"),
					},
				},
			},
			want: &user.SetEmailResponse{
				Details: &object.Details{
					Sequence:      1,
					ChangeDate:    timestamppb.Now(),
					ResourceOwner: Tester.Organisation.ID,
				},
			},
		},
		{
			name: "template error",
			req: &user.SetEmailRequest{
				UserId: userID,
				Email:  "custom-url@mouse.com",
				Verification: &user.SetEmailRequest_SendCode{
					SendCode: &user.SendEmailVerificationCode{
						UrlTemplate: gu.Ptr("{{"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "return code",
			req: &user.SetEmailRequest{
				UserId: userID,
				Email:  "return-code@mouse.com",
				Verification: &user.SetEmailRequest_ReturnCode{
					ReturnCode: &user.ReturnEmailVerificationCode{},
				},
			},
			want: &user.SetEmailResponse{
				Details: &object.Details{
					Sequence:      1,
					ChangeDate:    timestamppb.Now(),
					ResourceOwner: Tester.Organisation.ID,
				},
				VerificationCode: gu.Ptr("xxx"),
			},
		},
		{
			name: "is verified true",
			req: &user.SetEmailRequest{
				UserId: userID,
				Email:  "verified-true@mouse.com",
				Verification: &user.SetEmailRequest_IsVerified{
					IsVerified: true,
				},
			},
			want: &user.SetEmailResponse{
				Details: &object.Details{
					Sequence:      1,
					ChangeDate:    timestamppb.Now(),
					ResourceOwner: Tester.Organisation.ID,
				},
			},
		},
		{
			name: "is verified false",
			req: &user.SetEmailRequest{
				UserId: userID,
				Email:  "verified-false@mouse.com",
				Verification: &user.SetEmailRequest_IsVerified{
					IsVerified: false,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Client.SetEmail(CTX, tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			integration.AssertDetails(t, tt.want, got)
			if tt.want.GetVerificationCode() != "" {
				assert.NotEmpty(t, got.GetVerificationCode())
			}
		})
	}
}

func TestServer_VerifyEmail(t *testing.T) {
	userResp := createHumanUser(t)
	tests := []struct {
		name    string
		req     *user.VerifyEmailRequest
		want    *user.VerifyEmailResponse
		wantErr bool
	}{
		{
			name: "wrong code",
			req: &user.VerifyEmailRequest{
				UserId:           userResp.GetUserId(),
				VerificationCode: "xxx",
			},
			wantErr: true,
		},
		{
			name: "wrong user",
			req: &user.VerifyEmailRequest{
				UserId:           "xxx",
				VerificationCode: userResp.GetEmailCode(),
			},
			wantErr: true,
		},
		{
			name: "verify user",
			req: &user.VerifyEmailRequest{
				UserId:           userResp.GetUserId(),
				VerificationCode: userResp.GetEmailCode(),
			},
			want: &user.VerifyEmailResponse{
				Details: &object.Details{
					Sequence:      1,
					ChangeDate:    timestamppb.Now(),
					ResourceOwner: Tester.Organisation.ID,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Client.VerifyEmail(CTX, tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			integration.AssertDetails(t, tt.want, got)
		})
	}
}
