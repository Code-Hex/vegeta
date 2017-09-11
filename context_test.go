package vegeta

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Code-Hex/vegeta/internal/header"
	"github.com/Code-Hex/vegeta/internal/mime"
	"github.com/Code-Hex/vegeta/internal/status"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	egJSON     = `{"id":1,"name":"Taro"}` + "\n"
	prettyJSON = `{
    "id": 1,
    "name": "Taro"
}
`
	egXML     = `<user><id>1</id><name>Taro</name></user>`
	prettyXML = `<user>
    <id>1</id>
    <name>Taro</name>
</user>`
	egForm = `id=1&name=Taro` + "\n"
)

type user struct {
	ID   int    `json:"id" xml:"id"`
	Name string `json:"name" xml:"name"`
}

func TestNewContext(t *testing.T) {
	e := InitEngine(t)

	req := httptest.NewRequest(POST, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(rec, req).(*ctx)

	// Logger
	assert.NotNil(t, c.Logger)
	// Xslate
	assert.NotNil(t, c.xslate)
	// errhandler
	assert.NotNil(t, c.errhandler)

	// Request
	assert.NotNil(t, c.Request())
	// Response
	assert.NotNil(t, c.Response())
	// Map
	assert.NotNil(t, c.store)
	// Handler
	assert.NotNil(t, c.handler)

	// String
	err := c.String(status.OK, "OK")
	if assert.NoError(t, err) {
		assert.Equal(t, "OK", rec.Body.String())
	}
}

func TestJSON(t *testing.T) {
	e := InitEngine(t)

	req := httptest.NewRequest(POST, "/", strings.NewReader(egJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(rec, req).(*ctx)

	err := c.JSON(http.StatusOK, user{1, "Taro"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, mime.ApplicationJSONCharsetUTF8, rec.Header().Get(header.ContentType))
		assert.Equal(t, egJSON, rec.Body.String())
	}

	// Pretty
	req = httptest.NewRequest(POST, "/?pretty=1", strings.NewReader(egJSON))
	rec = httptest.NewRecorder()
	c = e.NewContext(rec, req).(*ctx)

	err = c.JSON(http.StatusOK, user{1, "Taro"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, mime.ApplicationJSONCharsetUTF8, rec.Header().Get(header.ContentType))
		assert.Equal(t, prettyJSON, rec.Body.String())
	}
}

func TestXML(t *testing.T) {
	e := InitEngine(t)

	req := httptest.NewRequest(POST, "/", strings.NewReader(egXML))
	rec := httptest.NewRecorder()
	c := e.NewContext(rec, req).(*ctx)

	err := c.XML(http.StatusOK, user{1, "Taro"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, mime.ApplicationXMLCharsetUTF8, rec.Header().Get(header.ContentType))
		assert.Equal(t, egXML, rec.Body.String())
	}

	// Pretty
	req = httptest.NewRequest(POST, "/?pretty=1", strings.NewReader(egXML))
	rec = httptest.NewRecorder()
	c = e.NewContext(rec, req).(*ctx)

	err = c.XML(http.StatusOK, user{1, "Taro"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, mime.ApplicationXMLCharsetUTF8, rec.Header().Get(header.ContentType))
		assert.Equal(t, prettyXML, rec.Body.String())
	}
}

func TestError(t *testing.T) {
	e := InitEngine(t)
	req := httptest.NewRequest(POST, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(rec, req).(*ctx)
	c.Error(errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestNoContent(t *testing.T) {
	e := InitEngine(t)
	// NoContent
	req := httptest.NewRequest(POST, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(rec, req).(*ctx)
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestContextCookie(t *testing.T) {
	e := InitEngine(t)
	req := httptest.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	data := "foo=bar"
	user := "user=John Manjirou"
	req.Header.Add(header.Cookie, data)
	req.Header.Add(header.Cookie, user)
	c := e.NewContext(rec, req).(*ctx)

	// Try to read single
	cookie, err := c.Cookie("foo")
	if assert.NoError(t, err) {
		assert.Equal(t, "foo", cookie.Name)
		assert.Equal(t, "bar", cookie.Value)
	}

	// Try to read multiple
	for _, cookie := range c.Cookies() {
		switch cookie.Name {
		case "foo":
			assert.Equal(t, "bar", cookie.Value)
		case "user":
			assert.Equal(t, "John Manjirou", cookie.Value)
		}
	}

	// Write
	ssid := url.QueryEscape("大塩平八郎のLAN")
	cookie = &http.Cookie{
		Name:     "SSID",
		Value:    ssid,
		Domain:   "history.love",
		Path:     "/",
		Expires:  time.Now(),
		Secure:   true,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	cookieStr := rec.Header().Get(header.SetCookie)
	assert.Contains(t, cookieStr, "SSID")
	assert.Contains(t, cookieStr, ssid)
	assert.Contains(t, cookieStr, "history.love")
	assert.Contains(t, cookieStr, "Secure")
	assert.Contains(t, cookieStr, "HttpOnly")
}

func TestContextPath(t *testing.T) {
	e := InitEngine(t)

	e.GET("/users/:id", nil)
	c := e.NewContext(nil, nil)
	e.Find(GET, "/users/1", c)
	assert.Equal(t, "/users/1", c.Path())

	e.Find(GET, "/users/12345678", c)
	assert.Equal(t, "/users/12345678", c.Path())
}
