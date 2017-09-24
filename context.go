package vegeta

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Context struct {
	echo.Context
	DB     *gorm.DB
	Zap    *zap.Logger
	Xslate *xslate.Xslate
}

type Vars = xslate.Vars

func CurrentYear() int {
	return time.Now().Year()
}

func (v *Vegeta) NewContext(ctx echo.Context) (*Context, error) {
	c := &Context{
		Context: ctx,
		DB:      v.DB,
		Zap:     v.Logger,
	}
	if err := c.setupXslate(); err != nil {
		return nil, errors.Wrap(err, "Failed to setup xslate")
	}
	return c, nil
}

func (c *Context) setupXslate() (err error) {
	c.Xslate, err = xslate.New(xslate.Args{
		"Loader": xslate.Args{
			"LoadPaths": []string{"./templates"},
		},
		"Parser": xslate.Args{"Syntax": "TTerse"},
		"Functions": xslate.Args{
			"year": CurrentYear,
		},
	})
	if err != nil {
		return errors.Wrap(err, "Failed to construct xslate")
	}
	return // nil
}

func (c *Context) RenderTemplate(tmpl string, vars Vars) error {
	return c.Xslate.RenderInto(c.Response(), tmpl, vars)
}

func (c *Context) RedirectWithJWT(code int, token, url string) error {
	if code < 300 || code > 308 {
		return echo.ErrInvalidRedirectCode
	}
	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	})
	resp := c.Response()
	resp.Header().Set(echo.HeaderLocation, url)
	resp.WriteHeader(code)
	return nil
}

var expiredAt = time.Now()

func (c *Context) ExpiredCookie() {
	for _, cookie := range c.Cookies() {
		cookie.Expires = expiredAt
		c.SetCookie(cookie)
	}
}
