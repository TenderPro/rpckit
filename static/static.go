package static

import (
	"mime"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"
	log "go.uber.org/zap"
)

// Config holds all config vars
type Config struct {
	Prefix string `long:"prefix" default:"html" description:"EmbeddedFS prefix"`
}

// Service holds SOAP service
type Service struct {
	Config      *Config
	Log         *log.Logger
	ListHandler func() []string
	DataHandler func(name string) ([]byte, error)
	InfoHandler func(name string) (os.FileInfo, error)
}

// New creates an Service object
func New(cfg Config, logger *log.Logger,
	lh func() []string,
	dh func(name string) ([]byte, error),
	ih func(name string) (os.FileInfo, error),
) *Service {
	return &Service{
		Config:      &cfg,
		Log:         logger,
		ListHandler: lh,
		DataHandler: dh,
		InfoHandler: ih,
	}
}

// SetupRouter add routes to mux
func (srv Service) SetupRouter(mux *http.ServeMux) {
	for _, v := range srv.ListHandler() {
		if !strings.HasPrefix(v, srv.Config.Prefix+"/") {
			continue
		}
		uri := strings.TrimPrefix(strings.TrimSuffix(v, "index.html"), srv.Config.Prefix)
		srv.Log.Info("Serve static", zap.String("uri", uri))
		mux.HandleFunc(uri, srv.serveStatic(v))
	}

}

func (srv Service) serveStatic(name string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		srv.Log.Info("Handle static", zap.String("file", name))
		data, err := srv.DataHandler(name)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
		} else {
			// TODO: add file info
			w.Header().Set("Content-Type", mime.TypeByExtension(name))
			w.Write(data)
		}
	}
}

// Proxy holds proxied handler
type Proxy struct {
	h http.Handler
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	p.h.ServeHTTP(w, r)

}

// OptionsProxy returns proxy for Options request handler
func OptionsProxy(h http.Handler) http.Handler { // , opts ...Option
	p := &Proxy{
		h: h,
	}
	return p
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, x-grpc-web")
}
