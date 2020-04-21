package rpckit_system

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	//	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/tmc/grpc-websocket-proxy/wsproxy"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	// opentracing "github.com/opentracing/opentracing-go"
	openlog "github.com/opentracing/opentracing-go/log"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rakyll/statik/fs"

	"github.com/nats-io/nats.go"

	_ "google.golang.org/grpc/encoding/gzip"

	pb "stage1/gars/proto"

	"stage1/gars/pkg/pkgconfig"
	"stage1/gars/pkg/pkglog"
	"stage1/gars/pkg/pkgmq"
	"stage1/gars/pkg/pkgsoap"
	_ "stage1/gars/pkg/pkgstatik"
	"stage1/gars/pkg/pkgtrace"
)

// Config holds all config vars
type Config struct {
	MQ               string `long:"mq_url" default:"localhost:4222" description:"Addr:port for NATS server"`
	OutsideHost      string `long:"host" default:"localhost:8081" description:"Addr:port for request from outside"`
	BindRPC          string `long:"bind_rpc" default:"localhost:9090" description:"Addr:port for gRPC server"`
	BindHTTP         string `long:"bind_http" default:":8081" description:"Addr:port for HTTP server"`
	TraceServiceName string `long:"trace_name" default:"proxy" description:"Tracing service name"`
	HTML             string `long:"html" default:"" description:"Path to static html files"`
	//	MemoryLimit int64          `long:"mem_max" default:"8"  description:"Memory limit for multipart forms, Mb"`
	SOAP  pkgsoap.Config  `group:"SOAP Options" namespace:"soap"`
	Trace pkgtrace.Config `group:"Trace Options" namespace:"trace"`
}

func run(exitFunc func(code int)) {
	var err error
	defer func() { pkgconfig.Close(exitFunc, err) }()

	cfg := &Config{}
	err = pkgconfig.New(cfg)
	if err != nil {
		return
	}

	log := pkglog.New(true)
	defer log.Sync()

	tracer, closer, er := pkgtrace.New(cfg.Trace, log, cfg.TraceServiceName)
	if er != nil {
		err = er
		return
	}
	defer closer.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var client *pkgmq.TestServiceClient
	log.Info("Connecting to MQ")

	for {
		srv, err := nats.Connect(cfg.MQ, nats.Timeout(5*time.Second))
		if err == nil {
			client = pkgmq.NewTestServiceClient(srv, "insta")
			break
		}
		fmt.Print(".")
		time.Sleep(time.Second)
	}
	log.Info("Connected")

	//	failOnError(err, "Connect")

	//	defer client.Close()

	lis, err := net.Listen("tcp", cfg.BindRPC)
	if err != nil {
		return
	}

	// used for opentracing.GlobalTracer()

	span := tracer.StartSpan("init")
	//        span.SetTag("hello-to", helloTo)

	//opts := []grpc.ServerOption{} //grpc.WithInsecure()}
	opts := []grpc_zap.Option{
		//	grpc_zap.WithLevels(customFunc),
	}
	grpcServer := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(log, opts...),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(log, opts...),
			grpc_recovery.StreamServerInterceptor(),
		),
	)
	if err != nil {
		return
	}
	pb.RegisterTestServiceServer(grpcServer, client)
	grpc_prometheus.Register(grpcServer)

	var group errgroup.Group

	group.Go(func() error {
		return grpcServer.Serve(lis)
	})

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	gwm := runtime.NewServeMux()

	opts1 := []grpc.DialOption{grpc.WithInsecure()}
	err = pb.RegisterTestServiceHandlerFromEndpoint(ctx, gwm, cfg.BindRPC, opts1)
	if err != nil {
		return
	}

	mux := http.NewServeMux()

	if cfg.HTML != "" {
		log.Debug("use fs", zap.String("path", cfg.HTML))
		fs := http.FileServer(http.Dir(cfg.HTML))
		mux.Handle("/", fs)
	} else {
		log.Debug("use embedded")
		statikFS, er := fs.New()
		if er != nil {
			err = errors.Wrap(er, "Attach statik")
			return
		}
		//		mux.Handle("/", http.StripPrefix("/public/", http.FileServer(statikFS)))
		mux.Handle("/", http.FileServer(statikFS))
	}

	mux.Handle("/v1/", wsproxy.WebsocketProxy(gwm))
	mux.Handle("/metrics", promhttp.Handler())

	soapService, err := pkgsoap.New(cfg.SOAP, log, cfg.BindRPC, cfg.OutsideHost)
	if err != nil {
		return
	}
	soapService.SetupRouter(mux)

	handler := func(resp http.ResponseWriter, req *http.Request) {
		setupResponse(&resp, req)
		if (*req).Method == "OPTIONS" {
			return
		}
		// fmt.Printf(">> REQ Headers: %+v", req.Header)
		mux.ServeHTTP(resp, req)
	}
	httpServer := http.Server{
		Addr:    cfg.BindHTTP,
		Handler: http.HandlerFunc(handler),
	}

	span.LogFields(
		openlog.String("event", "string-format"),
		openlog.String("value", "http"),
	)
	span.LogKV("event", "println")
	span.Finish()

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	group.Go(func() error {
		return httpServer.ListenAndServe()
	})

	err = group.Wait()
	return
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, x-grpc-web")
}
