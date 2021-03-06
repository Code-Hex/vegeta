package vegeta

import (
	"encoding/gob"
	"net/http"
	"os"

	"github.com/Code-Hex/vegeta/html"
	"github.com/Code-Hex/vegeta/internal/model"
	"github.com/Code-Hex/vegeta/internal/session"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//go:generate hero -source=template -pkgname=html -dest=html
type jwtVegetaClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}

const authScheme = "Bearer"

var secret []byte

func init() {
	secret = []byte(os.Getenv("VEGETA_SECRET"))
	gob.Register(&model.User{}) // Register for session
}

func (v *Vegeta) registerRoutes() {
	store := sessions.NewCookieStore(secret)
	store.Options.HttpOnly = true
	v.Use(session.Middleware("vegeta-session", store))
	v.GET("/", Index())
	v.GET("/login", Login())
	v.POST("/auth", Auth())

	api := v.Group("/api")
	api.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return call(func(c *Context) error {
				req := c.Request()
				token := req.Header.Get("Authorization")
				l := len(authScheme)
				if len(token) > l+1 && token[:l] == authScheme {
					user, err := model.TokenAuth(c.DB, token[l+1:])
					if err != nil {
						return errors.Wrap(err, "Failed to auth by token")
					}
					c.Set("user", user)
					return next(c)
				}
				return errors.New("Incorrect authorization header")
			})
		},
	)
	api.GET("/data", GetDataList())
	api.GET("/tags", GetTagList())
	api.POST("/data", PostData())
	api.POST("/tag", PostTag())
	api.DELETE("/tag/:name", DeleteTag())

	auth := v.Group("/mypage")
	auth.Use(
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				err := next(c)
				if err != nil {
					if herr, ok := err.(*echo.HTTPError); ok && herr.Code == 404 {
						return err
					}
					v.Logger.Error("Error on restricted group", zap.Error(err))
					return c.Redirect(http.StatusFound, "/login")
				}
				return nil
			}
		},
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return call(func(c *Context) error {
				status := c.GetUserStatus()
				if !status.IsAuthed() {
					return c.Redirect(http.StatusFound, "/login")
				}
				return next(c)
			})
		},
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
			return call(func(c *Context) error {
				status := c.GetUserStatus()
				if !status.IsAdmin() {
					return c.Redirect(http.StatusFound, "/login")
				}
				return next(c)
			})
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
	return call(func(c *Context) error {
		s := session.Get(c)
		user := s.Get("user").(*model.User)
		users, err := model.GetUsers(c.DB)
		if err != nil {
			c.Zap.Error("Failed to get user list", zap.Error(err))
			return c.Redirect(http.StatusFound, "/mypage")
		}
		token, err := c.CreateAPIToken(user.Name)
		if err != nil {
			c.Zap.Error("Failed to create api token", zap.Error(err))
			return c.Redirect(http.StatusFound, "/mypage")
		}
		args := &adminArgs{
			Args:  c.GetUserStatus(),
			token: token,
			users: users,
		}
		html.Admin(args, c.Response())
		return nil
	})
}

type mypageArgs struct {
	html.Args
	user  *model.User
	token string
}

func (m *mypageArgs) Token() string     { return m.token }
func (m *mypageArgs) User() *model.User { return m.user }

func MyPage() echo.HandlerFunc {
	return call(func(c *Context) error {
		s := session.Get(c)
		cu := s.Get("user").(*model.User)
		user, err := model.FindUserByName(c.DB, cu.Name)
		if err != nil {
			return errors.Wrap(err, "Failed to find user")
		}
		t, err := c.CreateAPIToken(user.Name)
		if err != nil {
			return errors.Wrap(err, "Failed to create api token at mypage")
		}
		args := &mypageArgs{
			Args:  c.GetUserStatus(),
			user:  user,
			token: t,
		}
		html.MyPage(args, c.Response())
		return nil
	})
}

type settingsArgs struct {
	html.Args
	user  *model.User
	token string
}

func (s *settingsArgs) Token() string     { return s.token }
func (s *settingsArgs) User() *model.User { return s.user }

func Settings() echo.HandlerFunc {
	return call(func(c *Context) error {
		s := session.Get(c)
		cu := s.Get("user").(*model.User)
		user, err := model.FindUserByName(c.DB, cu.Name)
		if err != nil {
			return errors.Wrap(err, "Failed to find user")
		}
		t, err := c.CreateAPIToken(user.Name)
		if err != nil {
			return errors.Wrap(err, "Failed to create api token at mypage")
		}
		args := &settingsArgs{
			Args:  c.GetUserStatus(),
			user:  user,
			token: t,
		}
		html.Settings(args, c.Response())
		return nil
	})
}

func Logout() echo.HandlerFunc {
	return call(func(c *Context) error {
		s := session.Get(c)
		if err := s.Expire(); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/login")
	})
}

func Login() echo.HandlerFunc {
	return call(func(c *Context) error {
		arg := c.GetUserStatus()
		if arg.IsAuthed() {
			return c.Redirect(http.StatusFound, "/mypage")
		}
		html.Login(arg, c.Response())
		return nil
	})
}

func Auth() echo.HandlerFunc {
	return call(func(c *Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		user, err := model.BasicAuth(c.DB, username, password)
		if err != nil {
			c.Zap.Error("Failed to auth user", zap.String("username", username))
			return c.Redirect(http.StatusFound, "/login")
		}
		s := session.Get(c)
		s.Set("user", user)
		if err := s.Save(); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/mypage")
	})
}

func Index() echo.HandlerFunc {
	return call(func(c *Context) error {
		html.Index(c.GetUserStatus(), c.Response())
		return nil
	})
}
