package shared

import (
	"context"
	"net/http"

	tokenmanager "github.com/HexArch/go-chat/internal/services/frontend/internal/services/token-manager"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	logger       *zap.Logger
	tokenManager tokenmanager.TokenManager
	sessionStore sessions.Store
	sessionName  string
}

func NewAuthInterceptor(
	logger *zap.Logger,
	tokenManager tokenmanager.TokenManager,
	sessionStore sessions.Store,
	sessionName string,
) *AuthInterceptor {
	return &AuthInterceptor{
		logger:       logger,
		tokenManager: tokenManager,
		sessionStore: sessionStore,
		sessionName:  sessionName,
	}
}

func (i *AuthInterceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		r, ok := ctx.Value("http_request").(*http.Request)
		if !ok {
			return status.Error(codes.Internal, "no http request in context")
		}

		w, ok := ctx.Value("http_response").(http.ResponseWriter)
		if !ok {
			return status.Error(codes.Internal, "no http response in context")
		}

		session, err := i.sessionStore.Get(r, i.sessionName)
		if err != nil {
			i.logger.Error("Failed to get session", zap.Error(err))
			return status.Error(codes.Internal, "failed to get session")
		}

		token, err := i.tokenManager.GetAccessToken(ctx, session)
		if err != nil {
			i.logger.Error("Failed to get access token",
				zap.Error(err),
				zap.String("method", method))
			return err
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil && isTokenExpiredError(err) {
			i.logger.Debug("Token expired, attempting refresh",
				zap.String("method", method))

			token, err = i.tokenManager.RefreshAccessToken(ctx, session, r, w)
			if err != nil {
				i.logger.Error("Failed to refresh token",
					zap.Error(err),
					zap.String("method", method))
				return err
			}

			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
			err = invoker(ctx, method, req, reply, cc, opts...)
		}
		return err
	}
}

func isTokenExpiredError(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.Unauthenticated &&
			(st.Message() == "token has expired" ||
				st.Message() == "invalid token")
	}
	return false
}
