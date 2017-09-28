package vegeta

import (
	"net/http"

	"github.com/Code-Hex/saltissimo"
	jwt "github.com/dgrijalva/jwt-go"
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
					if herr, ok := err.(*echo.HTTPError); ok && herr.Code == 404 {
						return err
					}
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
	auth.GET("/mypage", MyPage())
	auth.GET("/settings", Settings())

	// only admin
	admin := auth.Group("/admin")
	admin.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				token, ok := c.Get("user").(*jwt.Token)
				if !ok {
					v.Info("Failed to check user is admin")
					return c.Redirect(http.StatusFound, "/login")
				}
				user := token.Claims.(*jwtVegetaClaims)
				if !user.Admin {
					v.Info("Failed to access admin page", zap.String("username", user.Name))
					return c.Redirect(http.StatusFound, "/login")
				}
				c.Set("username", user.Name)
				return next(c)
			}
		},
	)
	admin.GET("", Admin())

	api := admin.Group("/api")
	api.Use(
		middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:      &apiVegetaClaims{},
			SigningKey:  secret,
			TokenLookup: "header:Authorization",
			ContextKey:  "token",
		}),
	)
	api.POST("/create", JSONCreateUser())
	api.POST("/edit", JSONEditUser())
	api.POST("/delete", JSONDeleteUser())
}

type adminArgs struct {
	*baseArg
	token        string
	users        []*User
	isCreated    bool
	failedReason string
}

func Admin() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		users, err := GetUsers(ctx.DB)
		if err != nil {
			ctx.Zap.Info("Failed to get user list", zap.Error(err))
			return ctx.Redirect(http.StatusFound, "/mypage")
		}
		token, err := ctx.CreateAPIToken(ctx.Get("username").(string))
		if err != nil {
			ctx.Zap.Info("Failed to create api token", zap.Error(err))
			return ctx.Redirect(http.StatusFound, "/mypage")
		}
		args := &adminArgs{
			baseArg: ctx.GetUserStatus(),
			token:   token,
			users:   users,
		}
		AdminHTML(args, c.Response())
		return nil
	}
}

func MyPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		MyPageHTML(ctx.GetUserStatus(), ctx.Response())
		return nil
	}
}

func Settings() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		SettingsHTML(ctx.GetUserStatus(), ctx.Response())
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
		arg := ctx.GetUserStatus()
		if arg.isAuthed {
			return ctx.Redirect(http.StatusFound, "/mypage")
		}
		LoginHTML(arg, ctx.Response())
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
