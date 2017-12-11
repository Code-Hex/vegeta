package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	version      = "0.0.2"
	name         = "vegeta-cli"
	msg          = name + " project to collect large amounts of vegetable data using IoT"
	targetHost   = "https://vegeta.neo.ie.u-ryukyu.ac.jp"
	completedMsg = "Send Complete"
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
	fmt.Println(completedMsg)
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
		return nil
	}

	// Remove tag mode
	if c.Remove {
		err := c.deleteRequest("/api/tag/" + c.Tag)
		if err != nil {
			return errors.Wrap(err, "Failed to remove tag")
		}
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

func (c *CLI) postRequest(path string, v interface{}) error {
	url, err := c.makeURL(path)
	if err != nil {
		return errors.Wrap(err, "Failed to make URL")
	}
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(v); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	return c.sendRequest(req)
}

func (c *CLI) deleteRequest(path string) error {
	url, err := c.makeURL(path)
	if err != nil {
		return errors.Wrap(err, "Failed to make URL")
	}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	return c.sendRequest(req)
}

func (c *CLI) makeURL(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	base, err := url.Parse(c.URL)
	if err != nil {
		return "", err
	}
	url := base.ResolveReference(u).String()
	fmt.Println("Request to " + url)
	return url, nil
}
func (c *CLI) sendRequest(req *http.Request) error {
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
