package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Code-Hex/vegeta/internal/common"
	"github.com/Code-Hex/vegeta/internal/utils"

	"github.com/pkg/errors"
)

type CLI struct {
	Options
}

const (
	version    = "0.0.2"
	name       = "vegeta-cli"
	msg        = name + " project to collect large amounts of vegetable data using IoT"
	targetHost = "https://vegeta.neo.ie.u-ryukyu.ac.jp"
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
	// Add tag mode
	if c.Add {
		err := c.postRequest("/api/tag", &common.TagJSON{
			TagName: c.Tag,
		})
		if err != nil {
			return errors.Wrap(err, "Failed to add tag")
		}
		fmt.Println("Send Complete")
		return nil
	}

	if c.Remove {
		err := c.deleteRequest("/api/tag/" + c.Tag)
		if err != nil {
			return errors.Wrap(err, "Failed to remove tag")
		}
		fmt.Println("Send Complete")
		return nil
	}

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
	err = c.postRequest("/api/data", &common.PostDataJSON{
		TagName:    c.Tag,
		Payload:    jsonStr,
		RemoteAddr: addr,
		Hostname:   host,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to send data")
	}
	fmt.Println("Send complete")
	return nil
}

func (c *CLI) prepare() error {
	_, err := parseOptions(&c.Options, os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "Failed to parse command line args")
	}
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

func (c *CLI) postRequest(url string, v interface{}) error {
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(v); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", targetHost+url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	d := &common.ResultJSON{}
	if err := json.NewDecoder(resp.Body).Decode(d); err != nil {
		return err
	}

	if !d.IsSuccess {
		return errors.New(d.Reason)
	}

	return nil
}

func (c *CLI) deleteRequest(url string) error {
	req, err := http.NewRequest("DELETE", targetHost+url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	d := &common.ResultJSON{}
	if err := json.NewDecoder(resp.Body).Decode(d); err != nil {
		return err
	}

	if !d.IsSuccess {
		return errors.New(d.Reason)
	}

	return nil
}
