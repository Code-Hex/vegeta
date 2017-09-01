package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

type AccessLogger struct {
	*zap.Logger
}

func AccessLog(logger *zap.Logger) *AccessLogger {
	return &AccessLogger{logger}
}

func (a *AccessLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	a.Info("Accessed",
		zap.String("URI", r.RequestURI),
		zap.String("From", r.RemoteAddr),
	)
	next(rw, r)
}
