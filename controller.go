package vegeta

import (
	"time"

	"github.com/Code-Hex/saltissimo"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}

var secret []byte

func init() {
	var err error
	secret, err = saltissimo.RandomBytes(saltissimo.SaltLength)
	if err != nil {
		panic(err)
	}
}

func (v *Vegeta) registerRoutes() {
	v.GET("/test/:arg", Index())
	v.GET("/login", Login())
	v.POST("/auth", Auth())
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		return ctx.RenderTemplate("login.tt", Vars{})
	}
}

func Auth() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		username := ctx.FormValue("username")
		password := ctx.FormValue("password")
		user, err := BasicAuth(ctx.DB, username, password)
		if err != nil {
			ctx.Zap.Info("Failed to auth user", zap.String("username", username))
			return err
		}
		claims := &jwtCustomClaims{
			Name:  user.Name,
			Admin: false,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		t, err := token.SignedString(secret)
		if err != nil {
			return err
		}
		return ctx.RenderTemplate("index.tt", Vars{"token": t})
	}
}

func Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := c.(*Context)
		cc.Zap.Info("Hello", zap.String("Test", "Hi"))
		arg := cc.Param("arg")
		return cc.RenderTemplate("index.tt", Vars{"arg": arg})
	}
}
