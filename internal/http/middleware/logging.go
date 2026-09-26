package middleware

import (
	"fmt"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Logging records one completion event per HTTP request.
func Logging(log *zap.Logger, development ...bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			writer := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(writer, r)

			status := writer.Status()
			if status == 0 {
				status = http.StatusOK
			}
			if len(development) > 0 && development[0] {
				log.Info(fmt.Sprintf("%s %s  %d  %s", r.Method, r.URL.Path, status, time.Since(started).Round(time.Microsecond)))
				return
			}
			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", status),
				zap.Duration("duration", time.Since(started)),
			}
			if requestID := chimiddleware.GetReqID(r.Context()); requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}
			log.Info("request completed", fields...)
		})
	}
}
