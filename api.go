package vegeta

import (
	"context"
	"crypto/subtle"
	"net/http"

	"github.com/Code-Hex/vegeta/model"
	"github.com/Code-Hex/vegeta/protos"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	DB *gorm.DB
}

/* grpc */
func (v *Vegeta) NewAPI() *API {
	return &API{DB: v.DB}
}

func (a *API) AddData(ctx context.Context, r *protos.RequestFromDevice) (*protos.ResultResponse, error) {
	token := r.GetToken()
	user, err := model.TokenAuth(a.DB, token)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	tag, err := user.FindByTagName(a.DB, r.GetTagName())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	data := model.Data{
		RemoteAddr: r.GetRemoteAddr(),
		Payload:    r.GetPayload(),
		Hostname:   r.GetHostname(),
	}
	if err := tag.AddData(a.DB, data); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &protos.ResultResponse{}, nil
}

/* JSON API  */
type apiVegetaClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

type resultJSON struct {
	IsSuccess bool   `json:"is_success"`
	Reason    string `json:"reason"`
}

func JSONTagsData() echo.HandlerFunc {
	return func(c echo.Context) error {

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
		if err := ctx.BindValidate(createUser); err != nil {
			return err
		}

		password := createUser.Password
		verifyPassword := createUser.VerifyPassword
		if subtle.ConstantTimeCompare([]byte(password), []byte(verifyPassword)) != 1 {
			return ctx.JSON(http.StatusOK, &resultJSON{
				Reason: "入力したパスワードと確認用のパスワードが一致しませんでした。",
			})
		}
		username := createUser.Name
		isAdmin := createUser.IsAdmin
		if _, err := model.CreateUser(ctx.DB, username, password, isAdmin); err != nil {
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

type editUser struct {
	ID      string `json:"id" validate:"required"`
	IsAdmin bool   `json:"is_admin"`
}

func JSONEditUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		editUser := new(editUser)
		if err := ctx.BindValidate(editUser); err != nil {
			return err
		}

		userID := editUser.ID
		isAdmin := editUser.IsAdmin
		if _, err := model.EditUser(ctx.DB, userID, isAdmin); err != nil {
			ctx.Zap.Error("Failed to edit user", zap.Error(err))
			return ctx.JSON(http.StatusOK, &resultJSON{
				Reason: "ユーザー編集時にエラーが発生しました。",
			})
		}
		return ctx.JSON(http.StatusOK, &resultJSON{
			IsSuccess: true,
		})
	}
}

type deleteUser struct {
	ID string `json:"id" validate:"required"`
}

func JSONDeleteUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*Context)
		deleteUser := new(deleteUser)
		if err := ctx.BindValidate(deleteUser); err != nil {
			return err
		}

		userID := deleteUser.ID
		if _, err := model.DeleteUser(ctx.DB, userID); err != nil {
			ctx.Zap.Error("Failed to delete user", zap.Error(err))
			return ctx.JSON(http.StatusOK, &resultJSON{
				Reason: "ユーザー削除時にエラーが発生しました。",
			})
		}
		return ctx.JSON(http.StatusOK, &resultJSON{
			IsSuccess: true,
		})
	}
}
