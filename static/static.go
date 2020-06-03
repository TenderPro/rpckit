package static

import (
	"io"
	"mime"
	"net/http"

	"github.com/apisite/apitpl/lookupfs"
	muxer "github.com/gorilla/mux"

	"go.uber.org/zap"
	log "go.uber.org/zap"
)

// Config holds all config vars
type Config struct {
	Prefix string `long:"prefix" default:"html" description:"EmbeddedFS prefix"`
}

// Service holds SOAP service
type Service struct {
	Config *Config
	Log    *log.Logger
	fs     lookupfs.FileSystem
}

// New creates an Service object
func New(cfg Config, logger *log.Logger,
	fs lookupfs.FileSystem,
) *Service {
	return &Service{
		Config: &cfg,
		Log:    logger,
		fs:     fs,
	}
}

// SetupRouter add routes to mux
func (srv Service) SetupRouter(mux *muxer.Router) {
	/*

		   TODO: Walk
		for _, v := range srv.ListHandler() {
			if !strings.HasPrefix(v, srv.Config.Prefix+"/") {
				continue
			}
			uri := strings.TrimPrefix(strings.TrimSuffix(v, "index.html"), srv.Config.Prefix)
			srv.Log.Info("Serve static", zap.String("uri", uri))
			mux.HandleFunc(uri, srv.serveStatic(v))
		}
	*/
}

func (srv Service) serveStatic(name string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		srv.Log.Info("Handle static", zap.String("file", name))
		data, err := srv.fs.Open(name)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
		} else {
			// TODO: add file info
			w.Header().Set("Content-Type", mime.TypeByExtension(name))
			io.Copy(w, data) // TODO: optimal? errors?
		}
	}
}

// Proxy holds proxied handler for grpc-gateway calls
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
