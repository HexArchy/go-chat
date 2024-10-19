// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"net/http"

	"github.com/go-openapi/loads"
	"go.uber.org/zap"

	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations"
	usecases "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases"

	apiAuthentication "github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations/authentication"

	apiUserManagement "github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations/user_management"
)

type Handler struct {
	ops      *operations.SsoAPI
	logger   *zap.Logger
	useCases *usecases.UseCases
}

func NewHandler(logger *zap.Logger, useCases *usecases.UseCases) (*Handler, error) {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, err
	}

	r := &Handler{
		ops:      operations.NewSsoAPI(swaggerSpec),
		logger:   logger,
		useCases: useCases,
	}
	r.setUpHandlers()

	return r, nil
}

func (h *Handler) handlerFor(method, path string) http.Handler {
	r, _ := h.ops.HandlerFor(method, path)

	return r
}

func (h *Handler) setUpHandlers() {

	h.ops.UserManagementDeleteUserHandler = apiUserManagement.DeleteUserHandlerFunc(h.DeleteUserHandler)
	h.ops.AuthenticationLoginUserHandler = apiAuthentication.LoginUserHandlerFunc(h.LoginUserHandler)
	h.ops.AuthenticationRefreshTokenHandler = apiAuthentication.RefreshTokenHandlerFunc(h.RefreshTokenHandler)
	h.ops.AuthenticationRegisterUserHandler = apiAuthentication.RegisterUserHandlerFunc(h.RegisterUserHandler)
	h.ops.UserManagementUpdateUserHandler = apiUserManagement.UpdateUserHandlerFunc(h.UpdateUserHandler)

	// You can add your middleware to concrete route
	// h.ops.AddMiddlewareFor("%method%", "%route%", %middlewareBuilder%)

	// You can add your global middleware
	// h.ops.AddGlobalMiddleware(%middlewareBuilder%)

	configureAPI(h.ops)
}
