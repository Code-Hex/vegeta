package vegeta

import (
	"bufio"
	"net"
	"net/http"

	"github.com/Code-Hex/vegeta/internal/status"
)

type Response struct {
	Writer    http.ResponseWriter
	Size      int64
	Status    int
	Committed bool
}

// NewResponse creates a new instance of Response.
func NewResponse(w http.ResponseWriter) (r *Response) {
	return &Response{Writer: w}
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

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *Response) Flush() {
	r.Writer.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.Writer.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotifier interface to allow detecting
// when the underlying connection has gone away.
// This mechanism can be used to cancel long operations on the server if the
// client has disconnected before the response is ready.
// See [http.CloseNotifier](https://golang.org/pkg/net/http/#CloseNotifier)
func (r *Response) CloseNotify() <-chan bool {
	return r.Writer.(http.CloseNotifier).CloseNotify()
}

func (r *Response) reset(w http.ResponseWriter) {
	r.Writer = w
	r.Size = 0
	r.Status = status.OK
	r.Committed = false
}
