package middleware

import (
	"log"
	"net/http"
	"time"
)

func Chain(next http.Handler) http.Handler {
	return Logging(Recovery(next))
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		defer func() {
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started))
		}()

		next.ServeHTTP(w, r)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				
				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}

				log.Printf("request panic: %v", recovered)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
