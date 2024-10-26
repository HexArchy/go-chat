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

func (i *AuthInterceptor) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Получаем request и writer из контекста
		r, ok := ctx.Value("http_request").(*http.Request)
		if !ok {
			return nil, status.Error(codes.Internal, "no http request in context")
		}

		w, ok := ctx.Value("http_response").(http.ResponseWriter)
		if !ok {
			return nil, status.Error(codes.Internal, "no http response in context")
		}

		// Получаем сессию для текущего запроса
		session, err := i.sessionStore.Get(r, i.sessionName)
		if err != nil {
			i.logger.Error("Failed to get session", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to get session")
		}

		token, err := i.tokenManager.GetAccessToken(ctx, session)
		if err != nil {
			i.logger.Error("Failed to get access token",
				zap.Error(err),
				zap.String("method", method))
			return nil, err
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			if isTokenExpiredError(err) {
				i.logger.Debug("Token expired, attempting refresh",
					zap.String("method", method))

				token, err = i.tokenManager.RefreshAccessToken(ctx, session, r, w)
				if err != nil {
					i.logger.Error("Failed to refresh token",
						zap.Error(err),
						zap.String("method", method))
					return nil, err
				}

				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
				stream, err = streamer(ctx, desc, cc, method, opts...)
				if err != nil {
					i.logger.Error("Failed to create stream after token refresh",
						zap.Error(err),
						zap.String("method", method))
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		return newAuthStream(stream, i, session, r, w), nil
	}
}

type authStream struct {
	grpc.ClientStream
	interceptor *AuthInterceptor
	session     *sessions.Session
	request     *http.Request
	writer      http.ResponseWriter
}

func newAuthStream(
	stream grpc.ClientStream,
	interceptor *AuthInterceptor,
	session *sessions.Session,
	r *http.Request,
	w http.ResponseWriter,
) grpc.ClientStream {
	return &authStream{
		ClientStream: stream,
		interceptor:  interceptor,
		session:      session,
		request:      r,
		writer:       w,
	}
}

func (s *authStream) SendMsg(m interface{}) error {
	err := s.ClientStream.SendMsg(m)
	if err != nil && isTokenExpiredError(err) {
		ctx := s.Context()

		token, refreshErr := s.interceptor.tokenManager.RefreshAccessToken(ctx, s.session, s.request, s.writer)
		if refreshErr != nil {
			s.interceptor.logger.Error("Failed to refresh token during stream send",
				zap.Error(refreshErr))
			return err
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return s.ClientStream.SendMsg(m)
	}
	return err
}

func (s *authStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if err != nil && isTokenExpiredError(err) {
		ctx := s.Context()

		token, refreshErr := s.interceptor.tokenManager.RefreshAccessToken(ctx, s.session, s.request, s.writer)
		if refreshErr != nil {
			s.interceptor.logger.Error("Failed to refresh token during stream receive",
				zap.Error(refreshErr))
			return err
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return s.ClientStream.RecvMsg(m)
	}
	return err
}

func (i *AuthInterceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Получаем request и writer из контекста
		r, ok := ctx.Value("http_request").(*http.Request)
		if !ok {
			return status.Error(codes.Internal, "no http request in context")
		}

		w, ok := ctx.Value("http_response").(http.ResponseWriter)
		if !ok {
			return status.Error(codes.Internal, "no http response in context")
		}

		// Получаем сессию для текущего запроса
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
