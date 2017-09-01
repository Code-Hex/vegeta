package vegeta

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Code-Hex/vegeta/internal/header"
	"github.com/Code-Hex/vegeta/internal/mime"
)

type (
	ctx struct {
		request  *http.Request
		response *Response
		path     string
		pnames   []string
		pvalues  []string
		query    url.Values
		handler  HandlerFunc
		store    Map
		vegeta   *Vegeta
	}
	// Ctx represents the context of the current HTTP request. It holds request and
	// response objects, path, path parameters, data and registered handler.
	Ctx interface {
		Path() string
	}
)

const (
	defaultMemory = 32 << 20 // 32 MB
)

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
	http.SetCookie(c.Response(), cookie)
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

func (c *ctx) Vegeta() *Vegeta {
	return c.vegeta
}

func (c *ctx) Handler() HandlerFunc {
	return c.handler
}

func (c *ctx) SetHandler(h HandlerFunc) {
	c.handler = h
}
