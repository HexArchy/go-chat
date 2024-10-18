// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"encoding/json"
	"net/http"

	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/handlers"
)

func (h *Handler) SwaggerDocJSONHandler() http.Handler {
	specDoc, _ := loads.Analyzed(restapi.SwaggerJSON, "")

	b, _ := json.MarshalIndent(specDoc.Spec(), "", "  ")
	basePath := "/v1"
	handler := http.NotFoundHandler()

	return handlers.CORS()(middleware.Spec(basePath, b, handler))
}
