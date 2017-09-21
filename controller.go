package vegeta

import (
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
)

type Controller struct {
	*Vegeta
}

func NewController(v *Vegeta) (*Controller, error) {
	return &Controller{Vegeta: v}, nil
}

func (c *Controller) Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		arg := c.Param("arg")
		return c.String(http.StatusOK, arg)
	}
}
