package interceptors

import (
	"context"
	"errors"

	"github.com/Julia-ivv/shortener-url.git/internal/authorizer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// HandlerWithAuth adds user authentication to the handler.
func HandlerWithAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(authorizer.AccessToken)
		if len(values) > 0 {
			token = values[0]
		}
	}
	if len(token) == 0 {
		return nil, status.Error(codes.Internal, "missing token")
	}
	userID, err := authorizer.GetUserIDFromToken(token)
	if err != nil {
		var tokenErr *authorizer.TokenErr
		isTokenError := errors.As(err, &tokenErr)
		if isTokenError && (tokenErr.ErrType == authorizer.ParseError) {
			return nil, status.Error(codes.Unauthenticated, "parse token error")
		}
		if isTokenError && (tokenErr.ErrType == authorizer.NotValidToken) {
			return nil, status.Error(codes.Unauthenticated, "not valid token error")
		}
	}
	ctx = context.WithValue(ctx, authorizer.UserContextKey, userID)

	return handler(ctx, req)
}
