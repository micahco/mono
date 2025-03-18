package middleware

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// Handles user facing messages
type ErrorResponseFunc func(w http.ResponseWriter, message string, statusCode int)

// Handles internal server errors. Does NOT reveal errors to user.
type ServerErrorFunc func(w http.ResponseWriter, logMsg string, err error)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func WithErrorHandling(
	handler HandlerFunc,
	errResponse ErrorResponseFunc,
	serverErr ServerErrorFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			var validationError validation.Errors
			switch {
			case errors.As(err, &validationError):
				errResponse(w, validationError.Error(), http.StatusUnprocessableEntity)
			default:
				serverErr(w, "middlware: handled unexpected error", err)
			}
		}
	}
}

func WithTimeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "")
	}
}

func StripSlashes(next http.Handler) http.Handler {
	return middleware.StripSlashes(next)
}

func Metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		totalRequestsReceived.Add(1)
		next.ServeHTTP(w, r)
		totalResponsesSent.Add(1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}

func Recoverer(serverErr ServerErrorFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "close")

					serverErr(w, "middleware: recoverer", fmt.Errorf("%s", err))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func EnableCORS(trustedOrigins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")

			origin := r.Header.Get("Origin")

			if origin != "" && len(trustedOrigins) != 0 {
				for i := range trustedOrigins {
					if origin == trustedOrigins[i] {
						w.Header().Set("Access-Control-Allow-Origin", origin)

						// Respond to preflight request
						if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
							w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
							w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

							w.WriteHeader(http.StatusOK)
							return
						}
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimit(errResponse ErrorResponseFunc, rps, burst int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		type client struct {
			limiter  *rate.Limiter
			lastSeen time.Time
		}

		var (
			mu      sync.Mutex
			clients = make(map[string]*client)
			maxAge  = 3 * time.Minute
		)

		go func() {
			for {
				time.Sleep(time.Minute)

				// Lock the mutex to prevent any rate limiter checks from happening while
				// the cleanup is taking place.
				mu.Lock()

				for ip, client := range clients {
					if time.Since(client.lastSeen) > maxAge {
						delete(clients, ip)
					}
				}

				mu.Unlock()
			}
		}()

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realip.FromRequest(r)

			// Lock the mutex to prevent this code from being executed concurrently.
			mu.Lock()

			// Check to see if the IP address already exists in the map. If it doesn't, then
			// initialize a new rate limiter and add the IP address and limiter to the map.
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(rps), burst),
				}
			}

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				errResponse(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)

				return
			}

			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
