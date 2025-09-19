package main

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid/v5"
)

type contextKey string

const correlationKey contextKey = "correlation_id"

func (s *Server) correlationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = generateCorrelationID()
			w.Header().Set("X-Request-Id", reqID)
		}
		ctx := context.WithValue(r.Context(), correlationKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateCorrelationID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func correlationIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(correlationKey).(string); ok {
		return value
	}
	return ""
}
