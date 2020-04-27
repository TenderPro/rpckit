// Package config holds config data and operations
package config

import (
	"context"
	"errors"
	"fmt"

	//"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
)

// Config holds all config vars
type Config struct {
	Args struct {
		//nolint:staticcheck // Multiple struct tag "choice" is allowed
		Command string `choice:"mono" choice:"bus" choice:"handler" choice:"proxy" description:"mono|bus|handler|proxy"`
	} `positional-args:"yes" required:"yes"`
	Debug       bool   `long:"debug"  description:"Run in debug mode"`
	MQ          string `long:"mq_url" default:"localhost:4222" description:"Addr:port for NATS server"`
	OutsideHost string `long:"host" default:"localhost:8081" description:"Addr:port for request from outside"`
	BindHTTP    string `long:"bind_http" default:":8081" description:"Addr:port for HTTP server"`
	HTML        string `long:"html" default:"" description:"Path to static html files"`
	//	MemoryLimit int64          `long:"mem_max" default:"8"  description:"Memory limit for multipart forms, Mb"`
}

var (
	// ErrGotHelp returned after showing requested help
	ErrGotHelp = errors.New("help printed")
	// ErrBadArgs returned after showing command args error message
	ErrBadArgs = errors.New("option error printed")
	// ErrCancelled returned when Ctrl+C pressed (Interrupt signal)
	ErrCancelled = errors.New("cancelled")
)

// UseServer returns true if applied logic used
func (cfg Config) UseServer() bool { return cfg.Args.Command != "proxy" }

// UseProxy returns true if Proxy services used
func (cfg Config) UseProxy() bool { return cfg.Args.Command != "handler" }

// UseNRPC returns true if NRPC used
func (cfg Config) UseNRPC() bool { return cfg.Args.Command != "mono" }

// New loads flags from args (if given) or command flags and ENV otherwise
func New(cfg interface{}, args ...string) error {
	p := flags.NewParser(cfg, flags.Default) //  HelpFlag | PrintErrors | PassDoubleDash
	var err error
	if len(args) == 0 {
		_, err = p.Parse()
	} else {
		_, err = p.ParseArgs(args)
	}
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			return ErrGotHelp
		}
		fmt.Printf("Args error: %#v", err)
		return ErrBadArgs
	}
	return nil
}

// Close runs exit after deferred cleanups have run and able show config parsing errors
func Close(exitFunc func(code int), e error) {
	if e != nil {
		var code int
		switch e {
		case ErrGotHelp:
			code = 3
		case ErrBadArgs:
			code = 2
		case ErrCancelled:
			code = 0
		default:
			code = 1
			fmt.Printf("Run error: %+v\n", e)
		}
		//fmt.Printf("Program error [%d]: %+v\n", code, e)
		exitFunc(code)
	}
}

// WaitSignal waits group for gracefull shutdown
func WaitSignal(gctx context.Context) error {

	// Wait for interrupt signal to gracefully shutdown the server with
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalChannel:
		return ErrCancelled

	case <-gctx.Done():
		return gctx.Err()
	}
}
