package vegeta

import (
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/pkg/errors"
)

type Controller struct {
	DB     *gorm.DB
	Xslate *xslate.Xslate
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

func (v *Vegeta) NewController() (*Controller, error) {
	c := &Controller{DB: v.DB}
	if err := c.setupXslate(); err != nil {
		return nil, errors.Wrap(err, "Failed to setup xslate")
	}
	return c, nil
}

func (c *Controller) Index() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		arg := ctx.Param("arg")
		return ctx.String(http.StatusOK, arg)
	}
}
