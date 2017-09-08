package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/Code-Hex/vegeta"
	"github.com/Code-Hex/vegeta/app/vegeta/controller"
	"github.com/Code-Hex/vegeta/middleware"
	"github.com/Code-Hex/vegeta/protos"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	version = "0.0.1"
	name    = "vegeta"
	msg     = name + " project to collect large amounts of vegetable data using IoT"
)

var stdout io.Writer = os.Stdout

type Vegeta struct {
	Options
	Engine     *vegeta.Engine
	waitSignal chan os.Signal
}

func New() *Vegeta {
	sigch := make(chan os.Signal)
	signal.Notify(
		sigch,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	return &Vegeta{
		waitSignal: sigch,
		Engine:     vegeta.New(),
	}
}

func (v *Vegeta) Run() int {
	if e := v.run(); e != nil {
		exitCode, err := UnwrapErrors(e)
		if v.StackTrace {
			fmt.Fprintf(os.Stderr, "Error:\n  %+v\n", e)
		} else {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error:\n  %v\n", err)
			}
		}
		return exitCode
	}
	return 0
}

func (v *Vegeta) run() error {
	if err := v.prepare(); err != nil {
		return errors.Wrap(err, "Failed to prepare")
	}
	return v.serve()
}

func (v *Vegeta) serve() error {
	ctx := context.Background()
	go func() {
		err := v.Engine.Start(ctx)
		if err != nil {
			v.Engine.Warn("Server is stopped", zap.Error(err))
		}
	}()
	return v.wait(ctx)
}

func (v *Vegeta) wait(ctx context.Context) error {
	<-v.waitSignal
	return v.Engine.Shutdown(ctx)
}

func (v *Vegeta) prepare() error {
	_, err := parseOptions(&v.Options, os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "Failed to parse command line args")
	}
	v.Engine.Port = v.Options.Port
	return v.setupHandlers()
}

func (v *Vegeta) setupHandlers() error {
	v.Engine.UseMiddleWare(
		middleware.AccessLog,
		middleware.Recover,
	)
	v.Engine.GET("/test/:arg", controller.Index)
	v.Engine.GET("/panic", controller.Panic)
	s := grpc.NewServer()
	protos.RegisterCollectionServer(s, NewAPIServer())
	//r.POST("/api", s.ServeHTTP)
	return nil
}

func parseOptions(opts *Options, argv []string) ([]string, error) {
	o, err := opts.parse(argv)
	if err != nil {
		stdout.Write(opts.usage())
		return nil, errors.Wrap(err, "invalid command line options")
	}
	if opts.Version {
		fmt.Fprintf(stdout, "%s: %s\n", version, msg)
		return nil, makeIgnore()
	}
	if opts.Help {
		stdout.Write(opts.usage())
		return nil, makeIgnore()
	}

	return o, nil
}
