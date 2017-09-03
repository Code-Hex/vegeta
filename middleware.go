package vegeta

import (
	"fmt"

	"go.uber.org/zap"
)

func AccessLog(next HandlerFunc) HandlerFunc {
	return func(c Context) error {
		r := c.Request()
		c.Logger().Info("Accessed",
			zap.String("URI", r.RequestURI),
			zap.String("From", r.RemoteAddr),
		)
		return next(c)
	}
}

func Recover(next HandlerFunc) HandlerFunc {
	return func(c Context) error {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch r := r.(type) {
				case error:
					err = r
				default:
					err = fmt.Errorf("%v", r)
				}
				c.Logger().Panic(err.Error())
			}
		}()
		return next(c)
	}
}
