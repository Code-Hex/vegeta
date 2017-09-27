package vegeta

import (
	"crypto/subtle"
	"net/http"

	"github.com/Code-Hex/saltissimo"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/k0kubun/pp"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

//go:generate hero -source=template -pkgname=vegeta -dest=.

type apiVegetaClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

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

type resultJSON struct {
	IsSuccess bool   `json:"is_success"`
	Reason    string `json:"reason"`
}

func (v *Vegeta) registerRoutes() {
	v.GET("/", Index())
	v.GET("/login", Login())
	v.POST("/auth", Auth())

	api := v.Group("/api")
	api.Use(
		middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:      &apiVegetaClaims{},
			SigningKey:  secret,
			TokenLookup: "header:Authorization",
			ContextKey:  "token",
		}),
	)
	api.POST("/create", JSONCreateUser())

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

type createUser struct {
	Name           string `json:"name" validate:"required"`
	Password       string `json:"password" validate:"required"`
	VerifyPassword string `json:"verify_password" validate:"required"`
	IsAdmin        bool   `json:"is_admin"`
}

func JSONCreateUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		createUser := new(createUser)
		if err := c.Bind(createUser); err != nil {
			pp.Println(err)
			ctx.Zap.Error("Failed to bind from form", zap.Error(err))
			return ctx.JSON(http.StatusOK, &resultJSON{
				Reason: "フォーム内容を取得できませんでした: " + err.Error(),
			})
		}
		if err := c.Validate(createUser); err != nil {
			pp.Println(err)
			ctx.Zap.Error("Failed to validate form", zap.Error(err))
			return ctx.JSON(http.StatusOK, &resultJSON{
				Reason: "入力に誤りがあります: " + err.Error(),
			})
		}

		password := createUser.Password
		verifyPassword := createUser.VerifyPassword
		if subtle.ConstantTimeCompare([]byte(password), []byte(verifyPassword)) != 1 {
			ctx.Zap.Error("Invalid password")
			return ctx.JSON(http.StatusOK, &resultJSON{
				Reason: "入力したパスワードと確認用のパスワードが一致しませんでした。",
			})
		}
		username := createUser.Name
		isAdmin := createUser.IsAdmin
		if _, err := CreateUser(ctx.DB, username, password, isAdmin); err != nil {
			ctx.Zap.Error("Failed to create user", zap.Error(err))
			return ctx.JSON(http.StatusOK, &resultJSON{
				Reason: "ユーザー作成時にエラーが発生しました。",
			})
		}
		return ctx.JSON(http.StatusOK, &resultJSON{
			IsSuccess: true,
		})
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
		arg := ctx.GetUserStatus()
		if arg.isAuthed {
			return ctx.Redirect(http.StatusFound, "/mypage")
		}
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
