package vegeta

import (
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
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

func (c *Controller) ServeAPI(s *grpc.Server) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		r, w := ctx.Request(), ctx.Response()
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			s.ServeHTTP(w, r)
			return nil
		}
		return echo.ErrUnsupportedMediaType
	}
}
