package middleware

import (
	"expvar"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/justinas/nosurf"
)

// Handles user facing messages. Mimics http.Error()
type ErrorResponseFunc func(w http.ResponseWriter, message string, statusCode int)

func WithTimeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "")
	}
}

func StripSlashes(next http.Handler) http.Handler {
	return middleware.StripSlashes(next)
}

func Profiler() http.Handler {
	return middleware.Profiler()
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

func RateLimit(rps int, errResponse ErrorResponseFunc) func(next http.Handler) http.Handler {
	return httprate.Limit(
		rps,
		time.Second,
		httprate.WithKeyByRealIP(),
		httprate.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			errResponse(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		}),
	)
}

func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; frame-ancestors 'self'; form-action 'self';")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

// Calls failureHandler in case CSRF check fails
func NoSurf(failureHandler http.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)
		csrfHandler.SetBaseCookie(http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		csrfHandler.SetFailureHandler(failureHandler)

		return csrfHandler
	}
}
