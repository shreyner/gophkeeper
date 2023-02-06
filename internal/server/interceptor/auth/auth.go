package interceptor_auth

import (
	"context"

	"github.com/shreyner/gophkeeper/internal/server/pgk/stoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var headerAuthorizeToken = "token"

const tokenDataKet int = iota

func SetTokenDataCtx(ctx context.Context, tokenData *stoken.Data) context.Context {
	return context.WithValue(ctx, tokenDataKet, tokenData)
}

func GetTokenDataCtx(ctx context.Context) (*stoken.Data, bool) {
	v, ok := ctx.Value(tokenDataKet).(*stoken.Data)

	return v, ok
}

func Interceptor(stokenService *stoken.Service) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)

		if !ok {
			return handler(ctx, req)
		}

		var token string
		if values := md.Get(headerAuthorizeToken); len(values) != 0 {
			token = values[0]
		}

		if token == "" {
			return handler(ctx, req)
		}

		tokenData, err := stokenService.ParseToken(token)

		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "invalid token")
		}

		return handler(SetTokenDataCtx(ctx, tokenData), req)
	}
}
