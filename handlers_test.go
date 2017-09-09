package vegeta

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Code-Hex/vegeta/internal/status"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	e := New()
	if err := e.setup(); err != nil {
		t.Errorf("setup is failed: %s", err.Error())
	}
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

	c, b := request(GET, "/", e)
	assert.Equal(t, "123", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func request(method, path string, e *Engine) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}
