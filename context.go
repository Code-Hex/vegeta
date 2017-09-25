package vegeta

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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
	var isAuthed bool
	cookie, err := c.Cookie(keyName)
	if err == nil && cookie.Value != "" {
		token, err := jwt.ParseWithClaims(
			cookie.Value, // token
			&jwtVegetaClaims{},
			func(t *jwt.Token) (interface{}, error) {
				// Check the signing method
				if t.Method.Alg() != middleware.AlgorithmHS256 {
					return nil, fmt.Errorf("Unexpected jwt signing method=%v", t.Header["alg"])
				}
				return secret, nil
			},
		)
		isAuthed = err == nil && token.Valid
	}
	vars["isAuthed"] = isAuthed
	return c.Xslate.RenderInto(c.Response(), tmpl, vars)
}

const keyName = "token"

var expiredAt = time.Now()

func (c *Context) SetToken2Cookie(user *User) error {
	tm := time.Now().Add(time.Hour * 24)
	claims := &jwtVegetaClaims{
		Name:  user.Name,
		Admin: user.Admin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tm.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(secret)
	if err != nil {
		return errors.New("Failed to get jwt")
	}
	c.SetCookie(&http.Cookie{
		Path:     "/",
		Name:     keyName,
		Value:    t,
		Expires:  tm,
		HttpOnly: true,
	})
	return nil
}

func (c *Context) ExpiredCookie() {
	c.SetCookie(&http.Cookie{
		Path:     "/",
		Name:     keyName,
		Value:    "",
		MaxAge:   -1,
		Expires:  expiredAt,
		HttpOnly: true,
	})
}
