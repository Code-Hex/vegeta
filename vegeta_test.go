package vegeta

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/Code-Hex/exit"
)

func TestVegeta_Run(t *testing.T) {
	program := "dummy"
	var opts Options

	tests := []struct {
		name   string
		stdout string
		args   []string
		want   int
	}{
		{
			name:   "run test when -h",
			args:   []string{program, "-h"},
			stdout: string(opts.usage()),
			want:   exit.USAGE,
		},
		{
			name:   "run test when --help",
			args:   []string{program, "--help"},
			stdout: string(opts.usage()),
			want:   exit.USAGE,
		},
		{
			name:   "run test when -v",
			args:   []string{program, "-v"},
			stdout: fmt.Sprintf("%s: %s\n", version, msg),
			want:   exit.USAGE,
		},
		{
			name:   "run test when --version",
			args:   []string{program, "--version"},
			stdout: fmt.Sprintf("%s: %s\n", version, msg),
			want:   exit.USAGE,
		},
	}
	v := New()
	for _, tt := range tests {
		stdout = new(bytes.Buffer)
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			exitCode := v.Run()
			if exitCode != tt.want {
				t.Errorf("Vegeta.Run() = %v, want %v", exitCode, tt.want)
			}
			out := stdout.(*bytes.Buffer).String()
			if out != tt.stdout {
				t.Errorf("stdout in Vegeta.Run() = %s, want %s\nargs %v", out, tt.stdout, tt.args)
			}
		})
	}
}

func TestVegeta_run(t *testing.T) {
	program := "dummy"
	var opts Options
	tests := []struct {
		name    string
		stdout  string
		args    []string
		wantErr bool
	}{
		{
			name:    "run test when -h",
			args:    []string{program, "-h"},
			stdout:  string(opts.usage()),
			wantErr: true,
		},
		{
			name:    "run test when --help",
			args:    []string{program, "--help"},
			stdout:  string(opts.usage()),
			wantErr: true,
		},
		{
			name:    "run test when -v",
			args:    []string{program, "-v"},
			stdout:  fmt.Sprintf("%s: %s\n", version, msg),
			wantErr: true,
		},
		{
			name:    "run test when --version",
			args:    []string{program, "--version"},
			stdout:  fmt.Sprintf("%s: %s\n", version, msg),
			wantErr: true,
		},
	}
	v := New()
	for _, tt := range tests {
		stdout = new(bytes.Buffer)
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			if err := v.run(); (err != nil) != tt.wantErr {
				t.Errorf("Vegeta.run() error = %v, wantErr %v", err, tt.wantErr)
			}
			out := stdout.(*bytes.Buffer).String()
			if out != tt.stdout {
				t.Errorf("stdout in Vegeta.Run() = %s, want %s\nargs %v", out, tt.stdout, tt.args)
			}
		})
	}
}
