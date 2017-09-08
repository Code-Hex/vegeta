package controller

import (
	"github.com/Code-Hex/vegeta"
	"go.uber.org/zap"
)

func Index(c vegeta.Context) error {
	p := c.Params()
	err := c.Render("index.tt", vegeta.Vars{"arg": p.ByName("arg")})
	if err != nil {
		c.Logger().Error("render error", zap.Error(err))
		return err
	}
	return nil
}

func Panic(c vegeta.Context) error {
	panic("KOREHA PANIC DESUYO!!")
	return nil
}
