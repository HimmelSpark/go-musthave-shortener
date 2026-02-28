package middleware

import (
	"fmt"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			fmt.Println("Error parsing request data", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		for k, v := range r.Form {
			fmt.Println("k:", k, "v:", v)
		}
		next.ServeHTTP(w, r)
	})
}
