package vegeta

import (
	"context"

	"github.com/Code-Hex/vegeta/protos"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	DB *gorm.DB
}

func (v *Vegeta) NewAPI() *API {
	return &API{DB: v.DB}
}

func (a *API) AddData(ctx context.Context, r *protos.RequestFromDevice) (*protos.ResultResponse, error) {
	token := r.GetToken()
	user, err := TokenAuth(a.DB, token)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	tag, err := user.FindByTagName(a.DB, r.GetTagName())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	data := Data{
		RemoteAddr: r.GetRemoteAddr(),
		Payload:    r.GetPayload(),
		Hostname:   r.GetHostname(),
	}
	if err := tag.AddData(a.DB, data); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &protos.ResultResponse{}, nil
}
