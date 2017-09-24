package vegeta

import (
	"net/http"
	"time"

	"github.com/Code-Hex/saltissimo"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

type jwtVegetaClaims struct {
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
	v.GET("/", Index())
	v.GET("/login", Login())
	v.POST("/auth", Auth())

	auth := v.Group("/mypage")
	auth.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				err := next(c)
				if err != nil {
					v.Error("Error via restricted group", zap.Error(err))
					return c.Redirect(http.StatusFound, "/login")
				}
				return nil
			}
		},
		middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:      &jwtVegetaClaims{},
			SigningKey:  secret,
			TokenLookup: "cookie:token",
			ContextKey:  "user",
		}),
	)
	auth.GET("", MyPage())
	auth.GET("/settings", Settings())
}

func MyPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		token, ok := ctx.Get("user").(*jwt.Token)
		if !ok {
			ctx.Zap.Info("Failed to check user is authed")
			return ctx.Redirect(http.StatusFound, "/login")
		}
		user := token.Claims.(*jwtVegetaClaims)
		return ctx.RenderTemplate("index.tt", Vars{"name": user.Name})
	}
}

func Settings() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		user, ok := ctx.Get("user").(*jwtVegetaClaims)
		if !ok {
			return ctx.RenderTemplate("index.tt", Vars{"name": ""})
		}
		return ctx.RenderTemplate("index.tt", Vars{"name": user.Name})
	}
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		return ctx.RenderTemplate("login.tt", Vars{})
	}
}

func Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		ctx.ExpiredCookie()
		return ctx.Redirect(http.StatusFound, "/login")
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
			return ctx.Redirect(http.StatusFound, "/login")
		}
		claims := &jwtVegetaClaims{
			Name:  user.Name,
			Admin: false,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		t, err := token.SignedString(secret)
		if err != nil {
			ctx.Zap.Info("Failed to get jwt", zap.String("username", username))
			return ctx.Redirect(http.StatusFound, "/login")
		}
		return ctx.RedirectWithJWT(http.StatusFound, t, "/mypage")
	}
}

func Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		return ctx.RenderTemplate("index.tt", Vars{"name": ""})
	}
}
