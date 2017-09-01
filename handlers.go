package vegeta

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Code-Hex/vegeta/internal/middleware"
	"github.com/Code-Hex/vegeta/protos"
	"github.com/julienschmidt/httprouter"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/urfave/negroni"
	"google.golang.org/grpc"
)

type VegetaCtrlr struct {
}

func (v *Vegeta) setupHandler() {
	n := negroni.New(
		middleware.AccessLog(v.Logger),
		middleware.Recover(v.Logger),
	)
	n.UseHandler(v.router())
	v.Handler = n
}

func (v *Vegeta) router() http.Handler {
	r := httprouter.New()

	r.GET("/test/:arg", v.Index)
	r.GET("/panic", v.Panic)
	s := grpc.NewServer()
	protos.RegisterCollectionServer(s, NewAPIServer())
	r.HandlerFunc(POST, "/api", s.ServeHTTP)

	return r
}

func (v *Vegeta) Index(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := v.Xslate.RenderInto(w, "index.tt", xslate.Vars{"arg": p.ByName("arg")})
	if err != nil {
		v.Logger.Error("render error", zap.Error(err))
		fmt.Fprint(w, "Error!!")
		return
	}
}

func (v *Vegeta) Panic(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	panic("KOREHA PANIC DESUYO!!")
}
