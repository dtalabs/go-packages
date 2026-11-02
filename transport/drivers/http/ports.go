package http

import "net/http"

// Handler is the standard HTTP handler accepted by routers and middleware.
type Handler = http.Handler

// HandlerFunc adapts a function to an HTTP handler.
type HandlerFunc = http.HandlerFunc

// ResponseWriter writes an HTTP response.
type ResponseWriter = http.ResponseWriter

// Request is the standard HTTP request.
type Request = http.Request
