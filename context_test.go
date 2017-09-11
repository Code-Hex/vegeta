package vegeta

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Code-Hex/vegeta/internal/header"
	"github.com/Code-Hex/vegeta/internal/mime"
	"github.com/Code-Hex/vegeta/internal/status"
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
