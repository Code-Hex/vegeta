package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/Code-Hex/vegeta/protos"
	"google.golang.org/grpc"

	"github.com/Code-Hex/vegeta/internal/utils"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type CLI struct {
	Options
}

const (
	version    = "0.0.1"
	name       = "vegeta-cli"
	msg        = name + " project to collect large amounts of vegetable data using IoT"
	targetHost = "https://neo.ie.u-ryukyu.ac.jp/"
)

func main() {
	os.Exit(New().Run())
}

func New() *CLI {
	return &CLI{}
}

func (c *CLI) Run() int {
	if e := c.run(); e != nil {
		exitCode, err := UnwrapErrors(e)
		if c.StackTrace {
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

func (c *CLI) run() error {
	if err := c.prepare(); err != nil {
		return errors.Wrap(err, "Failed to prepare")
	}
	if err := c.exec(); err != nil {
		return errors.Wrap(err, "Failed to exec")
	}
	return nil
}

func (c *CLI) exec() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	r := strings.NewReplacer(
		" ", "",
		"\n", "",
		"\t", "",
	)
	jsonStr := r.Replace(string(data))
	if !utils.IsValidJSON(jsonStr) {
		return errors.New("Invalid json format")
	}
	addr, err := utils.GetIPAddress()
	if err != nil {
		return errors.Wrap(err, "Failed to get ip address")
	}
	host, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "Failed to get hostname")
	}
	conn, err := c.setupConnection()
	if err != nil {
		return errors.Wrap(err, "Failed to connect grpc")
	}
	defer conn.Close()

	cli := protos.NewCollectionClient(conn)
	_, err = cli.AddData(context.Background(), &protos.RequestFromDevice{
		TagName:    c.Tag,
		Payload:    jsonStr,
		RemoteAddr: addr,
		Hostname:   host,
		Token:      c.Token,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to send data")
	}
	fmt.Println("Send complete")
	return nil
}

func (c *CLI) setupConnection() (*grpc.ClientConn, error) {
	// Setup grpc stub
	if isDevelopment() {
		addr := fmt.Sprintf("localhost:%d", c.Port)
		return grpc.Dial(addr, grpc.WithInsecure())
	}
	addr := fmt.Sprintf("%s:%d", targetHost, c.Port)
	return grpc.Dial(addr, grpc.WithTimeout(5*time.Second))
}

func (c *CLI) prepare() error {
	_, err := parseOptions(&c.Options, os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "Failed to parse command line args")
	}
	c.Port = c.Options.Port
	return nil
}

func parseOptions(opts *Options, argv []string) ([]string, error) {
	o, err := opts.parse(argv)
	if err != nil {
		os.Stdout.Write(opts.usage())
		return nil, errors.Wrap(err, "invalid command line options")
	}
	if opts.Version {
		fmt.Printf("%s: %s\n", version, msg)
		return nil, makeIgnore()
	}
	if opts.Help {
		os.Stdout.Write(opts.usage())
		return nil, makeIgnore()
	}
	return o, nil
}

func isDevelopment() bool {
	return os.Getenv("STAGE") == "development"
}
