package client

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	log "github.com/inconshreveable/log15"

	"github.com/kahgeh/lemonade/lemon"
	"github.com/kahgeh/lemonade/param"
	"github.com/kahgeh/lemonade/server"
)

type client struct {
	host               string
	port               int
	lineEnding         string
	noFallbackMessages bool
	logger             log.Logger
	timeout            time.Duration
}

func New(c *lemon.CLI, logger log.Logger) *client {
	return &client{
		host:               c.Host,
		port:               c.Port,
		lineEnding:         c.LineEnding,
		noFallbackMessages: c.NoFallbackMessages,
		logger:             logger,
		timeout:            c.Timeout,
	}
}

var dummy = &struct{}{}

func fileExists(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
}

func serveFile(fname string) (string, <-chan struct{}, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", nil, err
	}
	finished := make(chan struct{})

	go func() {
		http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := ioutil.ReadFile(fname)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(b)

			w.(http.Flusher).Flush()
			finished <- struct{}{}
		}))
	}()

	return fmt.Sprintf("http://127.0.0.1:%d/%s", l.Addr().(*net.TCPAddr).Port, fname), finished, nil
}

func (c *client) Open(uri string, transLocalfile, transLoopback bool) error {
	var finished <-chan struct{}
	if transLocalfile && fileExists(uri) {
		var err error
		uri, finished, err = serveFile(uri)
		if err != nil {
			return err
		}
	}

	c.logger.Info("Opening " + uri)
	err := c.withRPCClient(func(rc *rpc.Client) error {
		p := &param.OpenParam{
			URI:           uri,
			TransLoopback: transLoopback || transLocalfile,
		}

		return rc.Call("URI.Open", p, dummy)
	})
	if err != nil {
		return err
	}

	if finished != nil {
		<-finished
	}
	return nil
}

func (c *client) Paste() (string, error) {
	var resp string

	err := c.withRPCClient(func(rc *rpc.Client) error {
		return rc.Call("Clipboard.Paste", dummy, &resp)
	})
	if err != nil {
		return "", err
	}

	return lemon.ConvertLineEnding(resp, c.lineEnding), nil
}

func (c *client) Copy(text string) error {
	c.logger.Debug("Sending: " + text)
	return c.withRPCClient(func(rc *rpc.Client) error {
		return rc.Call("Clipboard.Copy", text, dummy)
	})
}

func (c *client) withRPCClient(f func(*rpc.Client) error) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), c.timeout)
	if err != nil {
		if !c.noFallbackMessages {
			c.logger.Error(err.Error())
			c.logger.Error("Falling back to localhost")
		}
		conn, err = c.fallbackLocal()
	}
	rc := rpc.NewClient(conn)
	return f(rc)
}

func (c *client) TmuxSendKeys(target, keys string, selectTargetPane bool) error {
	c.logger.Debug("Sending tmux keys", "target", target, "keys", keys)
	return c.withRPCClient(func(rc *rpc.Client) error {
		p := &param.TmuxSendKeysParam{
			Target:           target,
			Keys:             keys,
			SelectTargetPane: selectTargetPane,
		}
		err := rc.Call("Tmux.SendKeys", p, dummy)
		if err != nil {
			c.logger.Error("Tmux.SendKeys failed", "error", err)
		}
		return err
	})
}

func (c *client) fallbackLocal() (net.Conn, error) {
	port, err := server.ServeLocal(c.logger)
	server.LineEndingOpt = c.lineEnding
	if err != nil {
		return nil, err
	}
	return net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), c.timeout)
}
