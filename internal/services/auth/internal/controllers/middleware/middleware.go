package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/controllers/cache"
	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/metrics"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/validatetoken"
	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthMiddleware struct {
	logger          *zap.Logger
	metrics         *metrics.AuthMetrics
	tokenCache      *cache.TokenCache
	validateTokenUC *validatetoken.UseCase
	serviceToken    string
	circuitBreaker  *gobreaker.CircuitBreaker
}

func NewAuthMiddleware(
	logger *zap.Logger,
	metrics *metrics.AuthMetrics,
	tokenCache *cache.TokenCache,
	validateTokenUC *validatetoken.UseCase,
	serviceToken string,
) *AuthMiddleware {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "auth-validation",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	})

	return &AuthMiddleware{
		logger:          logger,
		metrics:         metrics,
		tokenCache:      tokenCache,
		validateTokenUC: validateTokenUC,
		serviceToken:    serviceToken,
		circuitBreaker:  cb,
	}
}

func (m *AuthMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()
		m.metrics.ActiveRequests.Inc()
		defer m.metrics.ActiveRequests.Dec()

		// Skip auth for public methods.
		if isPublicMethod(info.FullMethod) {
			return m.handleRequest(ctx, req, info, handler, startTime)
		}

		token, err := m.extractToken(ctx)
		if err != nil {
			m.recordError("token_extraction")
			return nil, err
		}

		if token == m.serviceToken {
			return m.handleRequest(ctx, req, info, handler, startTime)
		}

		userID, err := m.validateAndCacheToken(ctx, token)
		if err != nil {
			m.recordError("token_validation")
			return nil, err
		}

		newCtx := context.WithValue(ctx, UserIDKey, userID)
		return m.handleRequest(newCtx, req, info, handler, startTime)
	}
}

func (m *AuthMiddleware) handleRequest(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
	startTime time.Time,
) (interface{}, error) {
	resp, err := handler(ctx, req)

	status := "success"
	if err != nil {
		status = "error"
	}

	m.metrics.RequestDuration.WithLabelValues(
		info.FullMethod,
		status,
	).Observe(time.Since(startTime).Seconds())

	return resp, err
}

func (m *AuthMiddleware) validateAndCacheToken(ctx context.Context, token string) (uuid.UUID, error) {
	// Check cache first.
	if userID, found := m.tokenCache.Get(token); found {
		m.metrics.CacheHits.WithLabelValues("token").Inc()
		return userID, nil
	}

	// Validate through circuit breaker.
	result, err := m.circuitBreaker.Execute(func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		return m.validateTokenUC.Execute(ctx, token)
	})

	if err != nil {
		return uuid.Nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	user := result.(*entities.User)
	m.tokenCache.Set(token, user.ID)

	return user.ID, nil
}

func (m *AuthMiddleware) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no metadata provided")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "no authorization header")
	}

	token := values[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(token, "Bearer "), nil
}

func (m *AuthMiddleware) recordError(errorType string) {
	m.metrics.ErrorsTotal.WithLabelValues(errorType).Inc()
}

func isPublicMethod(method string) bool {
	switch method {
	case "/auth.AuthService/Login",
		"/auth.AuthService/RegisterUser",
		"/auth.AuthService/RefreshToken",
		"/auth.AuthService/ValidateToken":
		return true
	default:
		return false
	}
}

func SetUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return uuid.Nil, status.Error(codes.Unauthenticated, "user ID not found in context")
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, status.Error(codes.Internal, "invalid user ID type in context")
	}

	return userID, nil
}

func HasUserID(ctx context.Context) bool {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return false
	}
	_, ok := value.(uuid.UUID)
	return ok
}
