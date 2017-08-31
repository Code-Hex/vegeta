package vegeta

import "net/http"

type Response struct {
	vegeta    *Vegeta
	Writer    http.ResponseWriter
	Size      int64
	Status    int
	Committed bool
}

// NewResponse creates a new instance of Response.
func NewResponse(w http.ResponseWriter, v *Vegeta) (r *Response) {
	return &Response{Writer: w, vegeta: v}
}

// Header returns the header map for the writer that will be sent by
// WriteHeader. Changing the header after a call to WriteHeader (or Write) has
// no effect unless the modified headers were declared as trailers by setting
// the "Trailer" header before the call to WriteHeader (see example)
// To suppress implicit response headers, set their value to nil.
// Example: https://golang.org/pkg/net/http/#example_ResponseWriter_trailers
func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

func (r *Response) WriteHeader(code int) {
	if r.Committed {
		r.vegeta.Logger.Warn("response already committed")
		return
	}
	r.Status = code
	r.Writer.WriteHeader(code)
	r.Committed = true
}

func (r *Response) Write(b []byte) (n int, err error) {
	if !r.Committed {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.Writer.Write(b)
	r.Size += int64(n)
	return
}
