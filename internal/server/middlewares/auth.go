package middlewares

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/shreyner/gophkeeper/internal/server/pgk/stoken"
)

var HeaderAuthorizationKey = "Authorization"
var TokenDataKeyCtx = "auth-token"

func SetTokenDataCtx(ctx context.Context, tokenData *stoken.Data) context.Context {
	return context.WithValue(ctx, TokenDataKeyCtx, tokenData)
}

func GetTokenDataCtx(ctx context.Context) (*stoken.Data, bool) {
	v, ok := ctx.Value(TokenDataKeyCtx).(*stoken.Data)

	return v, ok
}

func Authenticator(log *zap.Logger, stokenService stoken.JWTService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, r *http.Request) {
			authToken := r.Header.Get(HeaderAuthorizationKey)

			if authToken == "" {
				next.ServeHTTP(wr, r)
				return
			}

			tokenData, err := stokenService.ParseToken(authToken)

			if err != nil {
				http.Error(wr, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(wr, r.WithContext(SetTokenDataCtx(r.Context(), tokenData)))
		})
	}
}
