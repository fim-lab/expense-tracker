package middleware

import (
	"context"
	"net/http"
)

type DemoMiddleware struct{}

func NewDemoMiddleware() *DemoMiddleware {
	return &DemoMiddleware{}
}

func (dm *DemoMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		demoUserId := 0

		ctx := context.WithValue(r.Context(), "userID", demoUserId)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
