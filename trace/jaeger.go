package rpckit_trace

// code from github.com/yurishkuro/opentracing-tutorial/go/lib/tracing"

import (
	"fmt"
	"io"

	opentracing "github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	//	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	logzap "github.com/uber/jaeger-client-go/log/zap"
)

type Config struct {
	Host string `long:"host" default:"localhost" description:"Agent host"`
	Port string `long:"port" default:"6831" description:"Agent port"`
}

// Init returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func New(cfg Config, logger *zap.Logger, service string) (tracer opentracing.Tracer, closer io.Closer, err error) {
	c := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			LogSpans:           true,
		},
	}
	tracer, closer, err = c.New(service, config.Logger(logzap.NewLogger(logger)))
	if err == nil {
		opentracing.SetGlobalTracer(tracer)
	}
	return
}
