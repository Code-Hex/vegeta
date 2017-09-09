package vegeta

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/Code-Hex/vegeta/internal/status"
)

func TestNewHTTPError(t *testing.T) {
	type args struct {
		code    int
		message []string
	}
	tests := []struct {
		name string
		args args
		want *HTTPError
	}{
		{
			name: "Status is 404",
			args: args{
				code: status.NotFound,
			},
			want: &HTTPError{
				Code:    status.NotFound,
				Message: http.StatusText(status.NotFound),
			},
		},
		{
			name: "Status is 500",
			args: args{
				code: status.InternalServerError,
			},
			want: &HTTPError{
				Code:    status.InternalServerError,
				Message: http.StatusText(status.InternalServerError),
			},
		},
		{
			name: "Status is 500 but message is changed",
			args: args{
				code:    status.InternalServerError,
				message: []string{"Too Bad..."},
			},
			want: &HTTPError{
				Code:    status.InternalServerError,
				Message: "Too Bad...",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHTTPError(tt.args.code, tt.args.message...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHTTPError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPError_Error(t *testing.T) {
	type fields struct {
		Code    int
		Message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Status is 404 message",
			fields: fields{
				Code:    status.NotFound,
				Message: http.StatusText(status.NotFound),
			},
			want: fmt.Sprintf(
				"code=%d, message=%v",
				status.NotFound,
				http.StatusText(status.NotFound),
			),
		},
		{
			name: "Status is 500 message but message is changed",
			fields: fields{
				Code:    status.NotFound,
				Message: []string{"Too Bad..."}[0],
			},
			want: fmt.Sprintf(
				"code=%d, message=%v",
				status.NotFound,
				"Too Bad...",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			he := &HTTPError{
				Code:    tt.fields.Code,
				Message: tt.fields.Message,
			}
			if got := he.Error(); got != tt.want {
				t.Errorf("HTTPError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
