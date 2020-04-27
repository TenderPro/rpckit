package soap

import (
	"fmt"
	"net/http"
	"strings"

	log "go.uber.org/zap"

	"github.com/UNO-SOFT/grpcer"
	soapproxy "github.com/UNO-SOFT/soap-proxy"
	"google.golang.org/grpc"
)

// Config holds all config vars
type Config struct {
	Prefix string `long:"prefix" default:"soap/" description:"Service prefix"`
}

// Service holds SOAP service
type Service struct {
	Config *Config
	Log    *log.Logger
	Server *grpc.ClientConn
	Host   string
}

// New creates an Service object
func New(cfg Config, logger *log.Logger, endpoint, host string) (*Service, error) {

	cc, err := grpcer.Connect(endpoint, "", "")
	if err != nil {
		return nil, err
	}
	srv := &Service{
		Config: &cfg,
		Log:    logger,
		Server: cc,
		Host:   host,
	}
	return srv, nil
}

/*

Get current location in runtime:

X-Forwarded-Host:[gars.dev.lan]
X-Forwarded-Port:[80]
X-Forwarded-Proto:[http]

*/

func (srv Service) SetupRouter(mux *http.ServeMux, client grpcer.Client, WSDL string) {
	addr := fmt.Sprintf("://%s/%s", srv.Host, srv.Config.Prefix)
	handler := &soapproxy.SOAPHandler{
		Client:    client,
		WSDL:      soapproxy.Ungzb64(WSDL),
		Locations: []string{"http" + addr, "https" + addr},
		Log: func(keyvals ...interface{}) error {
			srv.Log.Warn("--- SOAP ---")
			var buf strings.Builder
			for i := 0; i < len(keyvals); i += 2 {
				fmt.Fprintf(&buf, "%s=%+v ", keyvals[i], keyvals[i+1])
			}
			srv.Log.Warn(buf.String())
			return nil
		},
	}
	mux.Handle("/"+srv.Config.Prefix, handler)
}
