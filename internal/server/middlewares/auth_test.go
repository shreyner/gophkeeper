package middlewares

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	stoken2 "github.com/shreyner/gophkeeper/internal/server/pgk/stoken/mock"
	"go.uber.org/zap"
)

func TestAuthenticator(t *testing.T) {
	type args struct {
		log *zap.Logger
	}
	tests := []struct {
		name string
		args args
		want func(next http.Handler) http.Handler
	}{
		{
			name: "Success check authentification user",
			args: args{
				log: zap.NewNop(),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			stokenMock := stoken2.NewMockJWTService(ctl)

			if got := Authenticator(tt.args.log, stokenMock); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Authenticator() = %v, want %v", got, tt.want)
			}
		})
	}
}
