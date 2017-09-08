package middleware

import (
	"fmt"

	"github.com/Code-Hex/vegeta"
	"go.uber.org/zap"
)

func AccessLog(next vegeta.HandlerFunc) vegeta.HandlerFunc {
	return func(c vegeta.Context) error {
		r := c.Request()
		c.Logger().Info("Accessed",
			zap.String("URI", r.RequestURI),
			zap.String("From", r.RemoteAddr),
		)
		return next(c)
	}
}

func Recover(next vegeta.HandlerFunc) vegeta.HandlerFunc {
	return func(c vegeta.Context) error {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch r := r.(type) {
				case error:
					err = r
				default:
					err = fmt.Errorf("%v", r)
				}
				c.Error(err)
			}
		}()
		return next(c)
	}
}
