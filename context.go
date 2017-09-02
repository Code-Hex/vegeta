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
	"github.com/julienschmidt/httprouter"
)

type Context struct {
	*zap.Logger
	*xslate.Xslate
	request  *http.Request
	response *Response
	params   httprouter.Params
	path     string
	query    url.Values
	handler  HandlerFunc
	store    Map
}

const defaultMemory = 32 << 20 // 32 MB

func (c *Context) Path() string {
	return c.path
}

func (c *Context) SetPath(p string) {
	c.path = p
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *Context) Response() *Response {
	return c.response
}

func (c *Context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *Context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *Context) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *Context) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *Context) FormParams() (url.Values, error) {
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

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
}

func (c *Context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *Context) Get(key string) interface{} {
	v, ok := c.store.Load(key)
	if !ok {
		return nil
	}
	return v
}

func (c *Context) Set(key string, val interface{}) {
	c.store.Store(key, val)
}

// NewContext returns a Context instance.
func (v *Vegeta) NewContext(r *http.Request, w http.ResponseWriter) *Context {
	return &Context{
		request:  r,
		response: NewResponse(w),
		Logger:   v.Logger,
		Xslate:   v.Xslate,
		store:    &sync.Map{},
		handler:  NotFoundHandler,
	}
}

func (v *Vegeta) CreateContext(w http.ResponseWriter, r *http.Request, params httprouter.Params) *Context {
	ctx := v.Pool.Get().(*Context)
	ctx.request = r
	ctx.response.reset(w)
	ctx.path = r.RequestURI
	ctx.params = params
	return ctx
}

func (v *Vegeta) ReUseContext(ctx *Context) {
	v.Pool.Put(ctx)
}
