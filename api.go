package vegeta

import (
	"context"

	"github.com/Code-Hex/vegeta/protos"
	"github.com/jinzhu/gorm"
)

type API struct {
	DB *gorm.DB
}

func (v *Vegeta) NewAPI() *API {
	return &API{DB: v.DB}
}

func (a *API) RecvData(ctx context.Context, r *protos.RequestFromDevice) (*protos.ResultResponse, error) {
	token := r.GetToken()
	user, err := TokenAuth(a.DB, token)
	if err != nil {
		return nil, err
	}
	tag, err := user.FindByTagName(a.DB, r.GetTagName())
	if err != nil {
		return nil, err
	}
	data := Data{
		RemoteAddr: r.GetRemoteAddr(),
		Payload:    r.GetPayload(),
		DeviceID:   r.GetDeviceID(),
	}
	if err := tag.AddData(a.DB, data); err != nil {
		return nil, err
	}
	return &protos.ResultResponse{}, nil
}
