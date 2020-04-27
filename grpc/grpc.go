package grpc

import (
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	//	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	go_grpc "google.golang.org/grpc"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	// opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

type Config struct {
	Bind string `long:"bind" default:"localhost:9090" description:"Addr:port for gRPC server"`
}

// Server holds all gRPC server vars
type Service struct {
	config Config
	log    *zap.Logger
	Server *go_grpc.Server
	Mux    *runtime.ServeMux
}

func New(cfg Config, logger *zap.Logger, tracer opentracing.Tracer) *Service {
	//opts := []go_grpc.ServerOption{} //go_grpc.WithInsecure()}
	opts := []grpc_zap.Option{
		//	grpc_zap.WithLevels(customFunc),
	}

	grpcServer := go_grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger, opts...),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger, opts...),
			grpc_recovery.StreamServerInterceptor(),
		),
	)
	srv := &Service{config: cfg, log: logger, Server: grpcServer, Mux: runtime.NewServeMux()}
	return srv
}

func (srv *Service) ListenAndServe() error {
	grpc_prometheus.Register(srv.Server)

	lis, err := net.Listen("tcp", srv.config.Bind)
	if err != nil {
		return err
	}
	return srv.Server.Serve(lis)
}

func (srv *Service) Shutdown() {
	srv.Server.GracefulStop()
}
