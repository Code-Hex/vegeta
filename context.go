package vegeta

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/lestrrat/go-xslate"
	"go.uber.org/zap"

	"github.com/Code-Hex/vegeta/internal/header"
	"github.com/Code-Hex/vegeta/internal/mime"
	"github.com/Code-Hex/vegeta/internal/status"
	"github.com/julienschmidt/httprouter"
)

type ctx struct {
	logger   *zap.Logger
	xslate   *xslate.Xslate
	request  *http.Request
	response *Response
	params   httprouter.Params
	path     string
	query    url.Values
	handler  HandlerFunc
	store    Map
}

type Context interface {
	Path() string
	SetPath(string)

	Request() *http.Request
	SetRequest(*http.Request)
	Response() *Response

	QueryParam(string) string
	QueryParams() url.Values
	QueryString() string

	Params() httprouter.Params
	FormValue(string) string
	FormParams() (url.Values, error)

	Cookie(string) (*http.Cookie, error)
	SetCookie(*http.Cookie)
	Cookies() []*http.Cookie

	Error(error)

	Get(key string) interface{}
	Set(key string, val interface{})

	Logger() *zap.Logger
	Render(tmpl string, vars xslate.Vars) error
}

const defaultMemory = 32 << 20 // 32 MB

func (c *ctx) Path() string {
	return c.path
}

func (c *ctx) SetPath(p string) {
	c.path = p
}

func (c *ctx) Request() *http.Request {
	return c.request
}

func (c *ctx) SetRequest(r *http.Request) {
	c.request = r
}

func (c *ctx) Response() *Response {
	return c.response
}

func (c *ctx) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *ctx) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *ctx) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *ctx) Params() httprouter.Params {
	return c.params
}

func (c *ctx) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *ctx) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.request.Header.Get(header.ContentType), mime.MultipartForm) {
		if err := c.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.request.Form, nil
}

func (c *ctx) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *ctx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
}

func (c *ctx) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *ctx) Get(key string) interface{} {
	v, ok := c.store.Load(key)
	if !ok {
		return nil
	}
	return v
}

func (c *ctx) Set(key string, val interface{}) {
	c.store.Store(key, val)
}

func (c *ctx) Error(err error) {
	var (
		code    = status.InternalServerError
		message = http.StatusText(code)
	)
	if he, ok := err.(*HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	} else {
		c.Logger().Error("Error", zap.Error(err))
	}
	c.Render("error.tt", xslate.Vars{"code": code, "message": message})
}

func (c *ctx) Logger() *zap.Logger {
	return c.logger
}

func (c *ctx) Render(tmpl string, vars xslate.Vars) error {
	return c.xslate.RenderInto(c.Response(), tmpl, vars)
}

// NewContext returns a Context instance.
func (v *Vegeta) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return &ctx{
		request:  r,
		response: NewResponse(w),
		logger:   v.Logger,
		xslate:   v.Xslate,
		store:    &sync.Map{},
		handler:  NotFoundHandler,
	}
}

func (v *Vegeta) CreateContext(w http.ResponseWriter, r *http.Request, params httprouter.Params) Context {
	ctx := v.Pool.Get().(*ctx)
	ctx.request = r
	ctx.response.reset(w)
	ctx.path = r.RequestURI
	ctx.params = params
	return ctx
}

func (v *Vegeta) ReUseContext(c Context) {
	v.Pool.Put(c)
}
