package testutils

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// WithChiURLParams подставляет параметры пути в контекст chi запроса для тестов.
func WithChiURLParams(req *http.Request, params map[string]string) *http.Request {
	chiCtx := chi.NewRouteContext()
	for k, v := range params {
		chiCtx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
}
