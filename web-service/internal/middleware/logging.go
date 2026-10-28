package middleware

import (
	"log"
	"net/http"
)

// Logging пишет метод и путь каждого входящего запроса.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
