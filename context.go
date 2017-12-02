package vegeta

import (
	"net/http"
	"time"

	"github.com/Code-Hex/vegeta/html"
	"github.com/Code-Hex/vegeta/internal/session"
	"github.com/Code-Hex/vegeta/model"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Context struct {
	echo.Context
	DB  *gorm.DB
	Zap *zap.Logger
}

type baseArg struct{ Authed, Admin bool }

func (b *baseArg) IsAuthed() bool { return b.Authed }
func (b *baseArg) IsAdmin() bool  { return b.Admin }
func (baseArg) Year() int         { return time.Now().Year() }

func (v *Vegeta) NewContext(ctx echo.Context) (*Context, error) {
	c := &Context{
		Context: ctx,
		DB:      v.DB,
		Zap:     v.Logger,
	}
	return c, nil
}

type callFunc func(c *Context) error

// a tiny hack
func call(h callFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return h(c.(*Context))
	}
}

func (c *Context) GetUserStatus() html.Args {
	var isAuthed, isAdmin bool
	s := session.Get(c)
	u, ok := s.Get("user").(*model.User)
	if ok {
		isAuthed = true
		isAdmin = u.Admin
	}
	return &baseArg{
		Authed: isAuthed,
		Admin:  isAdmin,
	}
}

func (c *Context) CreateAPIToken(username string) (string, error) {
	tm := time.Now().Add(time.Hour * 24)
	claims := &apiVegetaClaims{
		Name: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tm.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(secret)
	if err != nil {
		return "", errors.New("Failed to get api jwt")
	}
	return t, nil
}

func (c *Context) BindValidate(i interface{}) error {
	if err := c.Bind(i); err != nil {
		c.Zap.Error("Failed to bind from json", zap.Error(err))
		return c.JSON(http.StatusOK, &resultJSON{
			Reason: "リクエスト内容を取得できませんでした: " + err.Error(),
		})
	}
	if err := c.Validate(i); err != nil {
		c.Zap.Error("Failed to validate json", zap.Error(err))
		return c.JSON(http.StatusOK, &resultJSON{
			Reason: "入力に誤りがあります: " + err.Error(),
		})
	}
	return nil
}
