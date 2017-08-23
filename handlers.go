package vegeta

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/Code-Hex/vegeta/protos"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/stephens2424/muxchain"
	"google.golang.org/grpc"
)

func (v *Vegeta) registerHandlers() *httprouter.Router {
	r := httprouter.New()

	logHandler := http.HandlerFunc(v.LoggingHandler)
	index := http.HandlerFunc(v.Index)
	chain := muxchain.ChainHandlers(logHandler, index)
	r.GET("/:arg", ContextHandler(chain))
	s := grpc.NewServer()
	protos.RegisterCollectionServer(s, NewAPIServer())
	r.HandlerFunc("POST", "/api", s.ServeHTTP)

	return r
}

func (v *Vegeta) Index(w http.ResponseWriter, r *http.Request) {
	p := context.Get(r, "params").(httprouter.Params)
	err := v.Xslate.RenderInto(w, "index.tt", xslate.Vars{"arg": p.ByName("arg")})
	if err != nil {
		v.Logger.Error("render error", zap.Error(err))
		fmt.Fprint(w, "Error!!")
		return
	}
}

func (v *Vegeta) LoggingHandler(w http.ResponseWriter, r *http.Request) {
	v.Logger.Info("Detected access", zap.String("path", r.URL.Path))
}

func ContextHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		context.Set(r, "params", p)
		h.ServeHTTP(w, r)
	}
}
