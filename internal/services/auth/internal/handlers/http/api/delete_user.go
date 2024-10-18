// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"github.com/go-openapi/runtime/middleware"

	apiUserManagement "github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations/user_management"
)

func (h *Handler) DeleteUserHandler(params apiUserManagement.DeleteUserParams, principal interface{}) middleware.Responder {
	return middleware.NotImplemented("operation user_management DeleteUser has not yet been implemented")
}
