package vegeta

import (
	"net/http"
	"net/url"
	"strings"
)

type (
	// Context represents the context of the current HTTP request. It holds request and
	// response objects, path, path parameters, data and registered handler.
	Context interface{}
	context struct {
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
)

const (
	defaultMemory = 32 << 20 // 32 MB
)

func (c *context) Path() string {
	return c.path
}

func (c *context) SetPath(p string) {
	c.path = p
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *context) Response() *Response {
	return c.response
}

func (c *context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *context) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *context) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *context) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.request.Header.Get(HeaderContentType), MIMEMultipartForm) {
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

func (c *context) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

func (c *context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *context) Get(key string) interface{} {
	v, ok := c.store.Load(key)
	if !ok {
		return nil
	}
	return v
}

func (c *context) Set(key string, val interface{}) {
	c.store.Store(key, val)
}

func (c *context) Vegeta() *Vegeta {
	return c.vegeta
}

func (c *context) Handler() HandlerFunc {
	return c.handler
}

func (c *context) SetHandler(h HandlerFunc) {
	c.handler = h
}
