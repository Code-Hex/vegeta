package vegeta

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Code-Hex/vegeta/protos"

	"github.com/jinzhu/gorm"
	"google.golang.org/grpc"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/lestrrat/go-server-starter/listener"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	version = "0.0.1"
	name    = "vegeta"
	msg     = name + " project to collect large amounts of vegetable data using IoT"
)

var stdout io.Writer = os.Stdout

type Vegeta struct {
	Options
	*echo.Echo
	*zap.Logger
	DB         *gorm.DB
	GRPC       *grpc.Server
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
		Echo:       echo.New(),
		GRPC:       grpc.NewServer(),
	}
}

func (v *Vegeta) Run() int {
	defer v.close()
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

func (v *Vegeta) close() {
	if v.DB != nil {
		v.DB.Close()
	}
}

func (v *Vegeta) listen() (net.Listener, error) {
	var li net.Listener
	if os.Getenv("SERVER_STARTER_PORT") != "" {
		listeners, err := listener.ListenAll()
		if err != nil {
			return nil, errors.Wrap(err, "server-starter error")
		}
		if 0 < len(listeners) {
			li = listeners[0]
		}
	}

	if li == nil {
		var err error
		li, err = net.Listen("tcp", fmt.Sprintf(":%d", v.Port))
		if err != nil {
			return nil, errors.Wrap(err, "listen error")
		}
	}

	return li, nil
}

func (v *Vegeta) startServer() {
	li, err := v.listen()
	if err != nil {
		v.Error("Failed to get port for server", zap.Error(err))
		return
	}
	if os.Getenv("SERVER_STARTER_PORT") == "" {
		fmt.Println("Start Server at", li.Addr().String())
	}
	if err := v.Server.Serve(li); err != nil {
		v.Error("Server is stopped", zap.Error(err))
	}
}

func (v *Vegeta) serveGRPC() {
	protos.RegisterCollectionServer(v.GRPC, v.NewAPI())
	port := v.Port + 1
	li, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		v.Error("Failed to get port for grpc", zap.Error(err))
		return
	}
	fmt.Println("Start GRPC Server at", li.Addr().String())
	if err := v.GRPC.Serve(li); err != nil {
		v.Error("Failed to serve grpc", zap.Error(err))
	}
}

func (v *Vegeta) serve() error {
	ctx := context.Background()
	go v.startServer()
	go v.serveGRPC()
	return v.wait(ctx)
}

func (v *Vegeta) wait(ctx context.Context) error {
	<-v.waitSignal
	v.GRPC.GracefulStop()
	return v.Shutdown(ctx)
}

func (v *Vegeta) prepare() error {
	_, err := parseOptions(&v.Options, os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "Failed to parse command line args")
	}
	if s := os.Getenv("SERVER_STARTER_PORT"); s != "" {
		port := strings.Split(s, "=")
		i, err := strconv.ParseInt(port[0], 10, 32)
		if err != nil {
			return err
		}
		v.Port = int(i)
	} else {
		v.Port = v.Options.Port
	}

	if err := v.setup(); err != nil {
		return err
	}
	if v.Migrate {
		r := v.DB.AutoMigrate(&User{}, &Tag{}, &Data{})
		if err := r.Error; err != nil {
			return err
		}
		return makeIgnore()
	}
	return nil
}

func (v *Vegeta) setupHandlers() error {
	v.HTTPErrorHandler = v.ErrorHandler
	v.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc, err := v.NewContext(c)
			if err != nil {
				return err
			}
			return h(cc)
		}
	})
	v.Use(
		v.LogHandler(),
		middleware.Recover(),
	)

	// Add route for echo
	v.GET("/test/:arg", Index())

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
