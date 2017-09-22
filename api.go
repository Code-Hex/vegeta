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
	return &protos.ResultResponse{}, nil
}
