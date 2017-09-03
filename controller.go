package vegeta

import (
	"fmt"

	xslate "github.com/lestrrat/go-xslate"
	"go.uber.org/zap"
)

func Index(c Context) error {
	w := c.Response().Writer
	p := c.Params()
	err := c.Render(w, "index.tt", xslate.Vars{"arg": p.ByName("arg")})
	if err != nil {
		c.Logger().Error("render error", zap.Error(err))
		fmt.Fprint(w, "Error!!")
		return err
	}
	return nil
}

func Panic(c Context) error {
	panic("KOREHA PANIC DESUYO!!")
	return nil
}
