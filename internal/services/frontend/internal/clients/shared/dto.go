package shared

import (
	"net/http"

	"github.com/gorilla/sessions"
	"google.golang.org/grpc"
)

type authStream struct {
	grpc.ClientStream
	interceptor *AuthInterceptor
	session     *sessions.Session
	request     *http.Request
	writer      http.ResponseWriter
}
