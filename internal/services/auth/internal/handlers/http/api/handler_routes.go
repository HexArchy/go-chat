// Code generated; DO NOT EDIT.

package api

import (
	"fmt"
	"strings"

	"github.com/gorilla/mux"
)

const version = "1.0.0"

func (h *Handler) AddRoutes(router *mux.Router) {

	router.Handle("/api/v1/users/{userId}", h.handlerFor("DELETE", "/api/v1/users/{userId}")).Methods("DELETE")
	router.Handle("/api/v1/login", h.handlerFor("POST", "/api/v1/login")).Methods("POST")
	router.Handle("/api/v1/refresh", h.handlerFor("POST", "/api/v1/refresh")).Methods("POST")
	router.Handle("/api/v1/register", h.handlerFor("POST", "/api/v1/register")).Methods("POST")
	router.Handle("/api/v1/users/{userId}", h.handlerFor("POST", "/api/v1/users/{userId}")).Methods("POST")

	router.Handle("/swagger.json", h.SwaggerDocJSONHandler()).Methods("GET")
	router.Handle("/docs", h.SwaggerDocUIHandler()).Methods("GET")
}

func (h *Handler) GetVersion() string {
	return fmt.Sprintf("v%s", strings.Split(version, ".")[0])
}
