package vegeta

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func (v *Vegeta) LogHandler() echo.MiddlewareFunc {
	return func(before echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := before(c)
			stop := time.Now()

			w, r := c.Response(), c.Request()
			v.Logger.Info(
				"Detected access",
				zap.String("status", fmt.Sprintf("%d: %s", w.Status, http.StatusText(w.Status))),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("useragent", r.UserAgent()),
				zap.String("remote_ip", r.RemoteAddr),
				zap.Int64("latency", stop.Sub(start).Nanoseconds()/int64(time.Microsecond)),
			)
			return err
		}
	}
}

func (v *Vegeta) ErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg}
	}

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD { // Issue #608
			if err := c.NoContent(code); err != nil {
				goto ERROR
			}
		} else {
			if err := c.JSON(code, msg); err != nil {
				goto ERROR
			}
		}
	}
ERROR:
	v.Logger.Error("Error", zap.String("reason", err.Error()))
}
