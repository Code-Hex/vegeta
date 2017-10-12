package vegeta

import (
	"net/http"

	"github.com/Code-Hex/saltissimo"
	"github.com/Code-Hex/vegeta/html"
	"github.com/Code-Hex/vegeta/model"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

//go:generate hero -source=template -pkgname=html -dest=html
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

	authAPI := auth.Group("/api")
	authAPI.Use(
		middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:      &apiVegetaClaims{},
			SigningKey:  secret,
			TokenLookup: "header:Authorization",
			ContextKey:  "auth_api",
		}),
	)
	authAPI.PATCH("/regenerate", RegenerateToken())
	authAPI.POST("/reregister_password", ReRegisterPassword())
	authAPI.PUT("/add_tag", AddTag())
	authAPI.POST("/data", JSONTagsData())

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

	adminAPI := admin.Group("/api")
	adminAPI.Use(
		middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:      &apiVegetaClaims{},
			SigningKey:  secret,
			TokenLookup: "header:Authorization",
			ContextKey:  "admin_api",
		}),
	)
	adminAPI.POST("/create", JSONCreateUser())
	adminAPI.POST("/edit", JSONEditUser())
	adminAPI.POST("/delete", JSONDeleteUser())
}

type adminArgs struct {
	html.Args
	token        string
	users        model.Users
	isCreated    bool
	failedReason string
}

func (a *adminArgs) Token() string      { return a.token }
func (a *adminArgs) Users() model.Users { return a.users }
func (a *adminArgs) IsCreated() bool    { return a.isCreated }
func (a *adminArgs) Reason() string     { return a.failedReason }

func Admin() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		users, err := model.GetUsers(ctx.DB)
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
			Args:  ctx.GetUserStatus(),
			token: token,
			users: users,
		}
		html.Admin(args, c.Response())
		return nil
	}
}

type mypageArgs struct {
	html.Args
	user  *model.User
	token string
}

func (m *mypageArgs) Token() string     { return m.token }
func (m *mypageArgs) User() *model.User { return m.user }

func MyPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		token, ok := c.Get("user").(*jwt.Token)
		if !ok {
			ctx.Zap.Info("Failed to check user has a permission")
			return c.Redirect(http.StatusFound, "/login")
		}
		claim := token.Claims.(*jwtVegetaClaims)
		user, err := model.FindUserByName(ctx.DB, claim.Name)
		if err != nil {
			ctx.Zap.Info("Failed to get user via mypage")
			return c.Redirect(http.StatusFound, "/login")
		}
		t, err := ctx.CreateAPIToken(user.Name)
		if err != nil {
			ctx.Zap.Info("Failed to create api token at mypage", zap.Error(err))
			return c.Redirect(http.StatusFound, "/login")
		}
		args := &mypageArgs{
			Args:  ctx.GetUserStatus(),
			user:  user,
			token: t,
		}
		html.MyPage(args, ctx.Response())
		return nil
	}
}

type settingsArgs struct {
	html.Args
	user  *model.User
	token string
}

func (s *settingsArgs) Token() string     { return s.token }
func (s *settingsArgs) User() *model.User { return s.user }

func Settings() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		token, ok := c.Get("user").(*jwt.Token)
		if !ok {
			ctx.Zap.Info("Failed to check user has a permission")
			return c.Redirect(http.StatusFound, "/login")
		}
		claim := token.Claims.(*jwtVegetaClaims)
		user, err := model.FindUserByName(ctx.DB, claim.Name)
		if err != nil {
			ctx.Zap.Info("Failed to get user via mypage")
			return c.Redirect(http.StatusFound, "/login")
		}
		t, err := ctx.CreateAPIToken(user.Name)
		if err != nil {
			ctx.Zap.Info("Failed to create api token at mypage", zap.Error(err))
			return c.Redirect(http.StatusFound, "/login")
		}
		args := &settingsArgs{
			Args:  ctx.GetUserStatus(),
			user:  user,
			token: t,
		}
		html.Settings(args, ctx.Response())
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
		if arg.IsAuthed() {
			return ctx.Redirect(http.StatusFound, "/mypage")
		}
		html.Login(arg, ctx.Response())
		return nil
	}
}

func Auth() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		username := ctx.FormValue("username")
		password := ctx.FormValue("password")
		user, err := model.BasicAuth(ctx.DB, username, password)
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
		html.Index(ctx.GetUserStatus(), ctx.Response())
		return nil
	}
}
