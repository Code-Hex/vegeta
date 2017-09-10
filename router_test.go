package vegeta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Code-Hex/vegeta/internal/header"
	"github.com/Code-Hex/vegeta/internal/mime"
	"github.com/Code-Hex/vegeta/internal/status"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	e := InitEngine(t)
	// Route
	e.GET("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	e.POST("/abc", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	e.PATCH("/abc/def", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	e.PUT("/abc/def/ghi", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	e.OPTIONS("/abc/def/ghi/j", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	e.HEAD("/abc/def/ghi/j", func(c Context) error {
		return c.String(status.OK, "OK")
	})

	c := e.NewContext(nil, nil).(*ctx)

	valid1 := e.Find(GET, "/", c)
	assert.True(t, valid1)

	valid2 := e.Find(POST, "/abc", c)
	assert.True(t, valid2)

	valid3 := e.Find(PATCH, "/abc/def", c)
	assert.True(t, valid3)

	valid4 := e.Find(PUT, "/abc/def/ghi", c)
	assert.True(t, valid4)

	valid5 := e.Find(OPTIONS, "/abc/def/ghi/j", c)
	assert.True(t, valid5)

	valid6 := e.Find(HEAD, "/abc/def/ghi/j", c)
	assert.True(t, valid6)

	valid7 := e.Find(GET, "/abc", c)
	assert.False(t, valid7)

	valid8 := e.Find(POST, "/", c)
	assert.False(t, valid8)

	valid9 := e.Find(PUT, "/abc/def/ghi/j", c)
	assert.False(t, valid9)
}

func TestMiddleware(t *testing.T) {
	e := InitEngine(t)
	buf := new(bytes.Buffer)
	e.UseMiddleWare(
		func(next HandlerFunc) HandlerFunc {
			return func(c Context) error {
				buf.WriteString("1")
				return next(c)
			}
		},
		func(next HandlerFunc) HandlerFunc {
			return func(c Context) error {
				buf.WriteString("2")
				return next(c)
			}
		},
		func(next HandlerFunc) HandlerFunc {
			return func(c Context) error {
				buf.WriteString("3")
				return next(c)
			}
		},
	)

	// Route
	e.GET("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})

	c, b := GETrequest("/", e)
	assert.Equal(t, "123", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoMiddlewareError(t *testing.T) {
	e := InitEngine(t)
	e.UseMiddleWare(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return errors.New("error")
		}
	})
	e.GET("/", NotFoundHandler)
	c, _ := GETrequest("/", e)
	assert.Equal(t, status.InternalServerError, c)
}

func TestRouterParam(t *testing.T) {
	e := InitEngine(t)
	e.Handle(GET, "/users/:id", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*ctx)
	e.Find(GET, "/users/1", c)
	assert.Equal(t, "1", c.Params().ByName("id"))
}

func TestRouterTwoParam(t *testing.T) {
	e := InitEngine(t)
	e.Handle(GET, "/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*ctx)
	e.Find(GET, "/users/1/files/1", c)
	params := c.Params()
	assert.Equal(t, "1", params.ByName("uid"))
	assert.Equal(t, "1", params.ByName("fid"))
}

func TestRouterMatchAny(t *testing.T) {
	e := InitEngine(t)

	// Routes
	e.Handle(GET, "/users/*name", func(Context) error {
		return nil
	})
	c := e.NewContext(nil, nil).(*ctx)
	e.Find(GET, "/users/joe", c)
	assert.Equal(t, "/joe", c.Params().ByName("name"))
}

func TestGET(t *testing.T) {
	e := InitEngine(t)
	e.GET("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	code, body := GETrequest("/", e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
}

func TestGETWithParams(t *testing.T) {
	e := InitEngine(t)
	var param string
	e.GET("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := GETrequest("/"+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestGETWithQueryParam(t *testing.T) {
	e := InitEngine(t)

	var param string
	e.GET("/", func(c Context) error {
		param = c.QueryParam("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := GETrequest("/?foo="+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestGETWithParameters(t *testing.T) {
	e := InitEngine(t)

	var param, qparam string
	e.GET("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		qparam = c.QueryParam("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	qexpected := "Bob"
	code, body := GETrequest("/"+expected+"?foo="+qexpected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
	assert.Equal(t, qexpected, qparam)
}

func TestDELETE(t *testing.T) {
	e := InitEngine(t)
	e.DELETE("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	code, body := DELETErequest("/", e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
}

func TestDELETEWithParams(t *testing.T) {
	e := InitEngine(t)
	var param string
	e.DELETE("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := DELETErequest("/"+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestDELETEWithQueryParam(t *testing.T) {
	e := InitEngine(t)

	var param string
	e.DELETE("/", func(c Context) error {
		param = c.QueryParam("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := DELETErequest("/?foo="+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestDELETEWithParameters(t *testing.T) {
	e := InitEngine(t)

	var param, qparam string
	e.DELETE("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		qparam = c.QueryParam("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	qexpected := "Bob"
	code, body := DELETErequest("/"+expected+"?foo="+qexpected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
	assert.Equal(t, qexpected, qparam)
}

func TestHEAD(t *testing.T) {
	e := InitEngine(t)
	e.HEAD("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	code, body := HEADrequest("/", e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
}

func TestHEADWithParams(t *testing.T) {
	e := InitEngine(t)
	var param string
	e.HEAD("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := HEADrequest("/"+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestHEADWithQueryParam(t *testing.T) {
	e := InitEngine(t)

	var param string
	e.HEAD("/", func(c Context) error {
		param = c.QueryParam("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := HEADrequest("/?foo="+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestHEADWithParameters(t *testing.T) {
	e := InitEngine(t)

	var param, qparam string
	e.HEAD("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		qparam = c.QueryParam("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	qexpected := "Bob"
	code, body := HEADrequest("/"+expected+"?foo="+qexpected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
	assert.Equal(t, qexpected, qparam)
}

func TestOPTIONS(t *testing.T) {
	e := InitEngine(t)
	e.OPTIONS("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	code, body := OPTIONSrequest("/", e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
}

func TestOPTIONSWithParams(t *testing.T) {
	e := InitEngine(t)
	var param string
	e.OPTIONS("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := OPTIONSrequest("/"+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestOPTIONSWithQueryParam(t *testing.T) {
	e := InitEngine(t)

	var param string
	e.OPTIONS("/", func(c Context) error {
		param = c.QueryParam("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := OPTIONSrequest("/?foo="+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestOPTIONSWithParameters(t *testing.T) {
	e := InitEngine(t)

	var param, qparam string
	e.OPTIONS("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		qparam = c.QueryParam("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	qexpected := "Bob"
	code, body := OPTIONSrequest("/"+expected+"?foo="+qexpected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
	assert.Equal(t, qexpected, qparam)
}

func TestPOST(t *testing.T) {
	e := InitEngine(t)
	e.POST("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	code, body := POSTrequest("/", e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
}

func TestPOSTWithParams(t *testing.T) {
	e := InitEngine(t)
	var param string
	e.POST("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := POSTrequest("/"+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestPOSTWithQueryParam(t *testing.T) {
	e := InitEngine(t)

	var param string
	e.POST("/", func(c Context) error {
		param = c.Request().FormValue("foo")
		return c.String(status.OK, "OK")
	})
	values := url.Values{}
	expected := "Alice"
	values.Set("foo", expected)
	code, body := POSTrequestWithForm("/", e,
		strings.NewReader(values.Encode()),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestPOSTWithJSONParam(t *testing.T) {
	e := InitEngine(t)
	var param struct {
		Name string `json:"name"`
	}
	e.POST("/", func(c Context) error {
		body := c.Request().Body
		err := json.NewDecoder(body).Decode(&param)
		if err != nil {
			t.Fatalf("json decode is failed: %s", err.Error())
		}
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := POSTrequestWithJSON("/", e,
		strings.NewReader(
			fmt.Sprintf(`{"name":"%s"}`, expected),
		),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param.Name)
}

func TestPOSTWithParameters(t *testing.T) {
	e := InitEngine(t)

	var param, qparam string
	e.POST("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		qparam = c.Request().FormValue("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	qexpected := "Bob"
	values := url.Values{}
	values.Set("foo", qexpected)
	code, body := POSTrequestWithForm("/"+expected, e,
		strings.NewReader(values.Encode()),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
	assert.Equal(t, qexpected, qparam)
}

func TestPATCH(t *testing.T) {
	e := InitEngine(t)
	e.PATCH("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	code, body := PATCHrequest("/", e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
}

func TestPATCHWithParams(t *testing.T) {
	e := InitEngine(t)
	var param string
	e.PATCH("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := PATCHrequest("/"+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestPATCHWithQueryParam(t *testing.T) {
	e := InitEngine(t)

	var param string
	e.PATCH("/", func(c Context) error {
		param = c.Request().FormValue("foo")
		return c.String(status.OK, "OK")
	})
	values := url.Values{}
	expected := "Alice"
	values.Set("foo", expected)
	code, body := PATCHrequestWithForm("/", e,
		strings.NewReader(values.Encode()),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestPATCHWithJSONParam(t *testing.T) {
	e := InitEngine(t)
	var param struct {
		Name string `json:"name"`
	}
	e.PATCH("/", func(c Context) error {
		body := c.Request().Body
		err := json.NewDecoder(body).Decode(&param)
		if err != nil {
			t.Fatalf("json decode is failed: %s", err.Error())
		}
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := PATCHrequestWithJSON("/", e,
		strings.NewReader(
			fmt.Sprintf(`{"name":"%s"}`, expected),
		),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param.Name)
}

func TestPATCHWithParameters(t *testing.T) {
	e := InitEngine(t)

	var param, qparam string
	e.PATCH("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		qparam = c.Request().FormValue("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	qexpected := "Bob"
	values := url.Values{}
	values.Set("foo", qexpected)
	code, body := PATCHrequestWithForm("/"+expected, e,
		strings.NewReader(values.Encode()),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
	assert.Equal(t, qexpected, qparam)
}

func TestPUT(t *testing.T) {
	e := InitEngine(t)
	e.PUT("/", func(c Context) error {
		return c.String(status.OK, "OK")
	})
	code, body := PUTrequest("/", e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
}

func TestPUTWithParams(t *testing.T) {
	e := InitEngine(t)
	var param string
	e.PUT("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := PUTrequest("/"+expected, e)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestPUTWithQueryParam(t *testing.T) {
	e := InitEngine(t)

	var param string
	e.PUT("/", func(c Context) error {
		param = c.Request().FormValue("foo")
		return c.String(status.OK, "OK")
	})
	values := url.Values{}
	expected := "Alice"
	values.Set("foo", expected)
	code, body := PUTrequestWithForm("/", e,
		strings.NewReader(values.Encode()),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
}

func TestPUTWithJSONParam(t *testing.T) {
	e := InitEngine(t)
	var param struct {
		Name string `json:"name"`
	}
	e.PUT("/", func(c Context) error {
		body := c.Request().Body
		err := json.NewDecoder(body).Decode(&param)
		if err != nil {
			t.Fatalf("json decode is failed: %s", err.Error())
		}
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	code, body := PUTrequestWithJSON("/", e,
		strings.NewReader(
			fmt.Sprintf(`{"name":"%s"}`, expected),
		),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param.Name)
}

func TestPUTWithParameters(t *testing.T) {
	e := InitEngine(t)

	var param, qparam string
	e.PUT("/:name", func(c Context) error {
		param = c.Params().ByName("name")
		qparam = c.Request().FormValue("foo")
		return c.String(status.OK, "OK")
	})
	expected := "Alice"
	qexpected := "Bob"
	values := url.Values{}
	values.Set("foo", qexpected)
	code, body := PUTrequestWithForm("/"+expected, e,
		strings.NewReader(values.Encode()),
	)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "OK", body)
	assert.Equal(t, expected, param)
	assert.Equal(t, qexpected, qparam)
}

// Utils
func InitEngine(t *testing.T) *Engine {
	e := New()
	if err := e.setup(); err != nil {
		t.Errorf("setup is failed: %s", err.Error())
	}
	return e
}

// Like GET requests
func GETrequest(path string, e *Engine) (int, string) {
	return LikeGETrequest(GET, path, e)
}

func DELETErequest(path string, e *Engine) (int, string) {
	return LikeGETrequest(DELETE, path, e)
}

func HEADrequest(path string, e *Engine) (int, string) {
	return LikeGETrequest(HEAD, path, e)
}

func OPTIONSrequest(path string, e *Engine) (int, string) {
	return LikeGETrequest(OPTIONS, path, e)
}

func LikeGETrequest(method, path string, e *Engine) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	return request(req, e)
}

// Like POST requests
func POSTrequest(path string, e *Engine) (int, string) {
	return LikePOSTrequest(POST, path, e)
}

func POSTrequestWithJSON(path string, e *Engine, body io.Reader) (int, string) {
	return LikePOSTrequestWithJSON(POST, path, e, body)
}

func POSTrequestWithForm(path string, e *Engine, body io.Reader) (int, string) {
	return LikePOSTrequestWithForm(POST, path, e, body)
}

func PATCHrequest(path string, e *Engine) (int, string) {
	return LikePOSTrequest(PATCH, path, e)
}

func PATCHrequestWithJSON(path string, e *Engine, body io.Reader) (int, string) {
	return LikePOSTrequestWithJSON(PATCH, path, e, body)
}

func PATCHrequestWithForm(path string, e *Engine, body io.Reader) (int, string) {
	return LikePOSTrequestWithForm(PATCH, path, e, body)
}

func PUTrequest(path string, e *Engine) (int, string) {
	return LikePOSTrequest(PUT, path, e)
}

func PUTrequestWithJSON(path string, e *Engine, body io.Reader) (int, string) {
	return LikePOSTrequestWithJSON(PUT, path, e, body)
}

func PUTrequestWithForm(path string, e *Engine, body io.Reader) (int, string) {
	return LikePOSTrequestWithForm(PUT, path, e, body)
}

func LikePOSTrequest(method, path string, e *Engine) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	return request(req, e)
}

func LikePOSTrequestWithJSON(method, path string, e *Engine, body io.Reader) (int, string) {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set(header.ContentType, mime.ApplicationJSON)
	return request(req, e)
}

func LikePOSTrequestWithForm(method, path string, e *Engine, body io.Reader) (int, string) {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set(header.ContentType, mime.ApplicationForm)
	return request(req, e)
}

func request(req *http.Request, e *Engine) (int, string) {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}
