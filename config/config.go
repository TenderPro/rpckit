package rpckit_config

import (
	"errors"
	"fmt"

	"github.com/jessevdk/go-flags"
)

var (
	// ErrGotHelp returned after showing requested help
	ErrGotHelp = errors.New("help printed")
	// ErrBadArgs returned after showing command args error message
	ErrBadArgs = errors.New("option error printed")
)

// setupConfig loads flags from args (if given) or command flags and ENV otherwise
func New(cfg interface{}, args ...string) error {
	p := flags.NewParser(cfg, flags.Default) //  HelpFlag | PrintErrors | PassDoubleDash
	var err error
	if len(args) == 0 {
		_, err = p.Parse()
	} else {
		_, err = p.ParseArgs(args)
	}
	if err != nil {
		//fmt.Printf("Args error: %#v", err)
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			return ErrGotHelp
		}
		return ErrBadArgs
	}
	return nil
}

// exit after deferred cleanups have run
func Close(exitFunc func(code int), e error) {
	if e != nil {
		var code int
		switch e {
		case ErrGotHelp:
			code = 3
		case ErrBadArgs:
			code = 2
		default:
			code = 1
			fmt.Printf("Run error: %+v", e)
		}
		exitFunc(code)
	}
}
