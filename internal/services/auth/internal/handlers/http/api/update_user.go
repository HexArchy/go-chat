// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"github.com/go-openapi/runtime/middleware"

	apiUserManagement "github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations/user_management"
)

func (h *Handler) UpdateUserHandler(params apiUserManagement.UpdateUserParams, principal interface{}) middleware.Responder {
	return middleware.NotImplemented("operation user_management UpdateUser has not yet been implemented")
}
