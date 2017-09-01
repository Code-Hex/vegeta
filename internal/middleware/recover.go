package middleware

import (
	"net/http"

	"github.com/Code-Hex/vegeta/internal/header"
	"github.com/Code-Hex/vegeta/internal/mime"
	"go.uber.org/zap"
)

type Recovery struct {
	*zap.Logger
}

func Recover(logger *zap.Logger) *Recovery {
	return &Recovery{logger}
}

func (rec *Recovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			if rw.Header().Get(header.ContentType) == "" {
				rw.Header().Set(header.ContentType, mime.TextPlainCharsetUTF8)
			}

			rw.WriteHeader(http.StatusInternalServerError)

			rec.Panic("Panic is detected", zap.String("error", err.(string)))
			/*
				if rec.ErrorHandlerFunc != nil {
					func() {
						defer func() {
							if err := recover(); err != nil {
								rec.Logger.Printf("provided ErrorHandlerFunc panic'd: %s, trace:\n%s", err, debug.Stack())
								rec.Logger.Printf("%s\n", debug.Stack())
							}
						}()
						rec.ErrorHandlerFunc(err)
					}()
				}
			*/
		}
	}()

	next(rw, r)
}
