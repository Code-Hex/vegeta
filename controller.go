package vegeta

import (
	xslate "github.com/lestrrat/go-xslate"
	"go.uber.org/zap"
)

func Index(c Context) error {
	p := c.Params()
	err := c.Render("index.tt", xslate.Vars{"arg": p.ByName("arg")})
	if err != nil {
		c.Logger().Error("render error", zap.Error(err))
		return err
	}
	return nil
}

func Panic(c Context) error {
	panic("KOREHA PANIC DESUYO!!")
	return nil
}
