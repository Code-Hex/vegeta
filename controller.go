package vegeta

import (
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/pkg/errors"
)

type Controller struct {
	*Vegeta
	*xslate.Xslate
}

func (c *Controller) setupXslate() (err error) {
	c.Xslate, err = xslate.New(xslate.Args{
		"Loader": xslate.Args{
			"LoadPaths": []string{"./templates"},
		},
		"Parser": xslate.Args{"Syntax": "TTerse"},
	})
	if err != nil {
		return errors.Wrap(err, "Failed to construct xslate")
	}
	return // nil
}

func NewController(v *Vegeta) (*Controller, error) {
	c := &Controller{Vegeta: v}
	if err := c.setupXslate(); err != nil {
		return nil, errors.Wrap(err, "Failed to setup xslate")
	}
	return c, nil
}

func (c *Controller) Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		arg := c.Param("arg")
		return c.String(http.StatusOK, arg)
	}
}
