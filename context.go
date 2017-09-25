package vegeta

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Context struct {
	echo.Context
	DB  *gorm.DB
	Zap *zap.Logger
}

type baseArg struct{ isAuthed, isAdmin bool }

type Args interface {
	IsAuthed() bool
	IsAdmin() bool
	Year() int
}

func (b *baseArg) IsAuthed() bool { return b.isAuthed }
func (b *baseArg) IsAdmin() bool  { return b.isAdmin }
func (baseArg) Year() int         { return time.Now().Year() }

func (v *Vegeta) NewContext(ctx echo.Context) (*Context, error) {
	c := &Context{
		Context: ctx,
		DB:      v.DB,
		Zap:     v.Logger,
	}
	return c, nil
}

func (c *Context) GetUserStatus() *baseArg {
	var isAuthed, isAdmin bool
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
		if isAuthed {
			user, ok := token.Claims.(*jwtVegetaClaims)
			if ok {
				isAdmin = user.Admin
			}
		}
	}
	return &baseArg{
		isAuthed: isAuthed,
		isAdmin:  isAdmin,
	}
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
