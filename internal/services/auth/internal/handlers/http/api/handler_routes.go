// Code generated; DO NOT EDIT.

package api

import (
	"fmt"
	"strings"

	"github.com/gorilla/mux"
)

const version = "1.0.0"

func (h *Handler) AddRoutes(router *mux.Router) {
	router.Handle("/users/{userId}", h.handlerFor("DELETE", "/users/{userId}")).Methods("DELETE")
	router.Handle("/login", h.handlerFor("POST", "/login")).Methods("POST")
	router.Handle("/refresh", h.handlerFor("POST", "/refresh")).Methods("POST")
	router.Handle("/register", h.handlerFor("POST", "/register")).Methods("POST")
	router.Handle("/users/{userId}", h.handlerFor("POST", "/users/{userId}")).Methods("POST")

	router.Handle("/swagger.json", h.SwaggerDocJSONHandler()).Methods("GET")
	router.Handle("/docs", h.SwaggerDocUIHandler()).Methods("GET")
}

func (h *Handler) GetVersion() string {
	return fmt.Sprintf("v%s", strings.Split(version, ".")[0])
}
