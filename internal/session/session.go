package session

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
)

type session struct {
	name    string
	session *sessions.Session
	store   sessions.Store
	req     *http.Request
	res     http.ResponseWriter
	written bool
}

var Name = "session"

func Middleware(name string, store sessions.Store) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			req, res := ctx.Request(), ctx.Response()
			ctx.Set(Name, &session{name, nil, store, req, res, false})
			defer context.Clear(req)
			return h(ctx)
		}
	}
}

func Get(ctx echo.Context) *session {
	value := ctx.Get(Name)
	if value == nil {
		return nil
	}
	return value.(*session)
}

func (s *session) Save() error {
	if s.written {
		e := s.Session().Save(s.req, s.res)
		if e == nil {
			s.written = false
		}
		return e
	}
	return nil
}

func (s *session) Session() *sessions.Session {
	if s.session == nil {
		var err error
		s.session, err = s.store.Get(s.req, s.name)
		if err != nil {
			panic(err)
		}
	}
	return s.session
}

func (s *session) Get(key interface{}) interface{} {
	return s.Session().Values[key]
}

func (s *session) Set(key interface{}, val interface{}) {
	s.Session().Values[key] = val
	s.written = true
}

func (s *session) Delete(key interface{}) {
	delete(s.Session().Values, key)
	s.written = true
}

func (s *session) Clear() {
	for key := range s.Session().Values {
		s.Delete(key)
	}
}

func (s *session) Options(options *sessions.Options) {
	s.Session().Options = options
}

func (s *session) Expire() error {
	s.Session().Options.MaxAge = -1
	s.written = true
	return s.Save()
}
