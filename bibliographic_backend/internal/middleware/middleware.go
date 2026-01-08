package middleware

import (
	"bibliographic_litriture_gigachat/internal/config"
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	requestIDKey  requestIDctxKey = 0
	logMiddleware                 = "request logger middleware"
)

type requestIDctxKey int

type responseWriterWithStatus struct {
	http.ResponseWriter
	Status  int
	URLPath string
}

func createRequestID() string {
	return uuid.New().String()
}

// RequestWithLoggerMiddleware places logger inside request context
func RequestWithLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = createRequestID()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		logger := log.With().Str("request_id", requestID).Caller().Logger()

		ctxtWithLogger := logger.WithContext(ctx)

		customResponseWriter := NewResponseWriterWithStatus(w, r.URL.Path)

		requestStartTime := time.Now()

		if strings.HasPrefix(r.URL.Path, "/ws") {
			logger.Info().Msg(fmt.Sprintf("logged websocket connection from %s", r.RequestURI))
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(customResponseWriter, r.WithContext(ctxtWithLogger))
		status := customResponseWriter.Status
		logRequestData(r.WithContext(ctxtWithLogger), requestStartTime, logMiddleware, requestID, status, requestURLPath(w, r))
	})
}

// this WriteHeader method captures the status code and calls the original WriteHeader.
func (rws *responseWriterWithStatus) WriteHeader(statusCode int) {
	rws.Status = statusCode
	rws.ResponseWriter.WriteHeader(statusCode)
}

// wrap response
func NewResponseWriterWithStatus(w http.ResponseWriter, path string) *responseWriterWithStatus {
	return &responseWriterWithStatus{
		ResponseWriter: w,
		Status:         http.StatusOK,
		URLPath:        path,
	}
}

func requestURLPath(w http.ResponseWriter, r *http.Request) string {
	urlPath := mux.CurrentRoute(r)
	if urlPath == nil {
		http.Error(w, "Route not found", http.StatusNotFound)
		return ""
	}

	return urlPath.GetName()
}

func logRequestData(r *http.Request, start time.Time, msg string, requestID string, status int, path string) {
	logger := log.Ctx(r.Context())

	duration := time.Since(start)

	logger.Info().
		Str("method", r.Method).
		Str("remote_addr", r.RemoteAddr).
		Str("url", path).
		Str("request_id", requestID).
		Dur("work_time", duration).
		Int("status", status).
		Str("user_agent", r.UserAgent()).
		Str("host", r.Host).
		Str("real_ip", getRealIPAddr(r)).
		Int64("content_length", r.ContentLength).
		Str("start_time", start.Format(time.RFC3339)).
		Str("duration_human_readable", duration.String()).
		Int64("duration_ms", duration.Milliseconds()).
		Msg(msg)
}

func getRealIPAddr(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			realIP := strings.TrimSpace(parts[0])
			if net.ParseIP(realIP) != nil {
				return realIP
			}
		}
	}

	hostIPAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return hostIPAddr
}

// Middleware for enabling needed CORS
func MiddlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.Ctx(r.Context())

		w.Header().Set("Access-Control-Allow-Origin", viper.GetString(config.FrontendHostEnv))
		w.Header().Set("Access-Control-Allow-Methods", viper.GetString(config.AllowedMethodsEnv))
		w.Header().Set("Access-Control-Allow-Credentials", viper.GetString(config.AllowCredentialsEnv))
		w.Header().Set("Access-Control-Allow-Headers", viper.GetString(config.AllowedHeadersEnv))
		w.Header().Set("Access-Control-Expose-Headers", viper.GetString(config.AllowedHeadersEnv))

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			logger.Info().Msg(fmt.Sprintf("options asked from %s", r.RequestURI))
			return
		}

		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

// Middleware for preventing any panic in server, so it won't instantly crash
func PreventPanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			logger := log.Ctx(r.Context())

			if err := recover(); err != nil {
				logger.Error().Msgf("Catched by middleware: panic happend: %v", err)

				stackTrace := make([]byte, 4<<10)
				n := runtime.Stack(stackTrace, true)
				stackTrace = stackTrace[:n]

				logger.Error().Msg(fmt.Sprintf("Panic stack trace: %s ", strings.TrimSpace(string(stackTrace))))

				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
