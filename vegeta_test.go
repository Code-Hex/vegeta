package vegeta

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestRouter(t *testing.T) {
	e := New()
	want := httprouter.Params{
		httprouter.Param{
			Key:   "name",
			Value: "gopher",
		},
	}
	routed := false
	e.Handle("GET", "/user/:name", func(c Context) error {
		routed = true
		ps := c.Params()
		if !reflect.DeepEqual(ps, want) {
			t.Errorf("wrong wildcard values: want %v, got %v", want, ps)
		}
		return nil
	})

	req := httptest.NewRequest(GET, "/user/gopher", nil)
	rec := httptest.NewRecorder()
	if err := e.setup(); err != nil {
		t.Errorf("setup is failed: %s", err.Error())
	}
	e.ServeHTTP(rec, req)

	if !routed {
		t.Errorf("routing is failed: %s", rec.Body.String())
	}
}
