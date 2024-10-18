// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"

	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations"
)

//go:generate swagger generate server --target ../../api --name Sso --spec ../../../../../swagger.yaml --template-dir ./swagger/templates --principal interface{}

//lint:ignore U1000 example
func configureFlags(api *operations.SsoAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.SsoAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Authorization" header is set
	if api.BearerAuthAuth == nil {
		api.BearerAuthAuth = func(token string) (interface{}, error) {
			return nil, errors.NotImplemented("api key auth (BearerAuth) Authorization from header param [Authorization] has not yet been implemented")
		}
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return api.Serve(func(handler http.Handler) http.Handler {
		return handler
	})
}

// The TLS configuration before HTTPS server starts.
//
//lint:ignore U1000 example
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
//
//lint:ignore U1000 example
func configureServer(s *http.Server, scheme, addr string) {
}
