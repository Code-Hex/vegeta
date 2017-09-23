package vegeta

import (
	"net/http"

	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := c.(*Context)
		cc.Zap.Info("Hello", zap.String("Test", "Hi"))
		arg := c.Param("arg")
		return c.String(http.StatusOK, arg)
	}
}
