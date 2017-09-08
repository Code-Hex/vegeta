package main

import (
	"io"

	"github.com/Code-Hex/vegeta/protos"
)

type API struct{}

func NewAPIServer() *API {
	return &API{}
}

// RecvData for collection
func (a *API) RecvData(c protos.Collection_RecvDataServer) error {
	ctx := c.Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, err := c.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
}
