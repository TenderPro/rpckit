package template

import (
	"html/template"
	tmpl "html/template"
	"net/http"

	"go.uber.org/zap"
	log "go.uber.org/zap"

	"github.com/apisite/apitpl"
	"github.com/apisite/apitpl/lookupfs"
	"github.com/gorilla/mux"
)

// Config holds all config vars
type Config struct {
	lookupfs.Config
	Prefix     string `long:"prefix" default:"static/tmpl" description:"EmbeddedFS prefix"`
	BufferSize int    `long:"buffer_size" default:"64" description:"Template buffer size"`
}

// Service holds Template service
type Service struct {
	Config *Config
	Log    *log.Logger
	Funcs  template.FuncMap
	lfs    *lookupfs.LookupFileSystem
	ts     *apitpl.TemplateService
}

// New creates an Service object
func New(cfg Config, logger *log.Logger,
	version string,
	debug bool,
	fs lookupfs.FileSystem,
	handler http.Handler,
) (*Service, error) {

	// apitpl
	allFuncs := tmpl.FuncMap{
		"version": func() string { return version },
	}
	SetSimpleFuncs(allFuncs)
	SetProtoFuncs(allFuncs)

	lfs := lookupfs.New(cfg.Config).FileSystem(fs)
	ts, err := apitpl.New(cfg.BufferSize).Funcs(allFuncs).LookupFS(lfs).Parse()
	if err != nil {
		return nil, err
	}
	ts.ParseAlways(debug)

	return &Service{
		Config: &cfg,
		Log:    logger,
		Funcs:  allFuncs,
		lfs:    lfs,
		ts:     ts,
	}, nil
}

// SetupRouter add routes to mux
func (srv Service) SetupRouter(mux *mux.Router, apiPrefix string) {
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

func (srv Service) serve(name string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		srv.Log.Info("Handle static", zap.String("file", name))
		funcs := srv.Funcs
		// TODO: setup request funcs
		var meta apitpl.MetaData
		err := srv.ts.Execute(w, name, funcs, meta)
		/*
			TODO
			if err != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusNotFound)
			} else {
				// TODO: add file info
				w.Header().Set("Content-Type", mime.TypeByExtension(name))
				w.Write(data)
			}
		*/
		if err != nil {
			srv.Log.Error("Serve template", zap.Error(err))

		}
	}
}
