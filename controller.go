package vegeta

import (
	"net/http"
	"strconv"

	"github.com/Code-Hex/saltissimo"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/k0kubun/pp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

//go:generate hero -source=template -pkgname=vegeta -dest=.

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
	auth.GET("/logout", Logout())
	auth.GET("/settings", Settings())
	// only admin
	auth.GET("/admin", Admin())
}

func Admin() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		token, ok := ctx.Get("user").(*jwt.Token)
		if !ok {
			ctx.Zap.Info("Failed to check user is admin")
			return ctx.Redirect(http.StatusFound, "/login")
		}
		user := token.Claims.(*jwtVegetaClaims)
		if !user.Admin {
			ctx.Zap.Info("Failed to access admin page", zap.String("username", user.Name))
			return ctx.Redirect(http.StatusFound, "/login")
		}
		page := ctx.QueryParam("page")
		p, err := strconv.Atoi(page)
		if err != nil {
			p = 1
		}
		users, err := GetUsers(ctx.DB, 20, p)
		if err != nil {
			ctx.Zap.Info("Failed to get user list", zap.Error(err))
			return ctx.Redirect(http.StatusFound, "/mypage")
		}
		pp.Println(users)
		AdminHTML(users, ctx.GetUserStatus(), c.Response())
		return nil
	}
}

func MyPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		/*
			ctx := c.(*Context)
			token, ok := ctx.Get("user").(*jwt.Token)
			if !ok {
				ctx.Zap.Info("Failed to check user is authed")
				return ctx.Redirect(http.StatusFound, "/login")
			}
			user := token.Claims.(*jwtVegetaClaims)
			return ctx.RenderTemplate("mypage.tt", Vars{"user": user})
		*/
		return nil
	}
}

func Settings() echo.HandlerFunc {
	return func(c echo.Context) error {
		/*
			ctx := c.(*Context)
			token, ok := ctx.Get("user").(*jwt.Token)
			if !ok {
				ctx.Zap.Info("Failed to check user is authed")
				return ctx.Redirect(http.StatusFound, "/login")
			}
			user := token.Claims.(*jwtVegetaClaims)
			return ctx.RenderTemplate("settings.tt", Vars{"user": user})
		*/
		return nil
	}
}

func Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		ctx.ExpiredCookie()
		return ctx.Redirect(http.StatusFound, "/login")
	}
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		LoginHTML(ctx.GetUserStatus(), ctx.Response())
		return nil
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
		if err := ctx.SetToken2Cookie(user); err != nil {
			ctx.Zap.Error(
				"Failed to set token to cookie",
				zap.Error(err),
				zap.String("username", username),
			)
			return ctx.Redirect(http.StatusFound, "/login")
		}
		return ctx.Redirect(http.StatusFound, "/mypage")
	}
}

func Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		IndexHTML(ctx.GetUserStatus(), ctx.Response())
		return nil
	}
}
