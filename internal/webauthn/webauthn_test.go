package webauthn

import (
	"context"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel/internal/api/authz"
	caos_errs "github.com/zitadel/zitadel/internal/errors"
)

func TestConfig_serverFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *webauthn.WebAuthn
		wantErr error
	}{
		{
			name:    "webauthn error",
			args:    args{context.Background()},
			wantErr: caos_errs.ThrowInternal(nil, "WEBAU-UX9ta", "Errors.User.WebAuthN.ServerConfig"),
		},
		{
			name: "success",
			args: args{authz.WithRequestedDomain(context.Background(), "example.com")},
			want: &webauthn.WebAuthn{
				Config: &webauthn.Config{
					RPDisplayName: "DisplayName",
					RPID:          "example.com",
					RPOrigins:     []string{"https://example.com"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Config{
				DisplayName:    "DisplayName",
				ExternalSecure: true,
			}
			got, err := w.serverFromContext(tt.args.ctx)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Config.RPDisplayName, got.Config.RPDisplayName)
				assert.Equal(t, tt.want.Config.RPID, got.Config.RPID)
				assert.Equal(t, tt.want.Config.RPOrigins, got.Config.RPOrigins)
			}
		})
	}
}
