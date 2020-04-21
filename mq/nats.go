package rpckit_mq

import (
	"context"
	"io"

	jaeger "github.com/uber/jaeger-client-go"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	rpclog "github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

var (
	natsServerTag = opentracing.Tag{Key: string(ext.Component), Value: "gNATS-Server"}
	natsClientTag = opentracing.Tag{Key: string(ext.Component), Value: "gNATS-Client"}
)

// metadataTextMap extends a metadata.MD to be an opentracing textmap
type metadataTextMap metadata.MD

type clientSpanTagKey struct{}

func newServerSpanFromString(ctx context.Context, addon, fullMethodName string) (context.Context, opentracing.Span) {
	parentSpanCtx, _ := jaeger.ContextFromString(addon)
	opts := []opentracing.StartSpanOption{
		opentracing.ChildOf(parentSpanCtx),
		ext.SpanKindRPCServer,
		natsServerTag,
	}
	return newSpanFromParent(ctx, fullMethodName, opts)
}

func newClientSpanFromContext(ctx context.Context, fullMethodName string) (context.Context, opentracing.Span) {

	if !opentracing.IsGlobalTracerRegistered() {
		return ctx, nil
	}

	var parentSpanCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpanCtx = parent.Context()
	}
	opts := []opentracing.StartSpanOption{
		opentracing.ChildOf(parentSpanCtx),
		ext.SpanKindRPCClient,
		natsClientTag,
	}
	return newSpanFromParent(ctx, fullMethodName, opts)
}

func newSpanFromParent(ctx context.Context, fullMethodName string, opts []opentracing.StartSpanOption) (context.Context, opentracing.Span) {

	tracer := opentracing.GlobalTracer()
	if tagx := ctx.Value(clientSpanTagKey{}); tagx != nil {
		if opt, ok := tagx.(opentracing.StartSpanOption); ok {
			opts = append(opts, opt)
		}
	}
	clientSpan := tracer.StartSpan(fullMethodName, opts...)
	// Make sure we add this to the metadata of the call, so it gets propagated:
	md := metautils.ExtractOutgoing(ctx).Clone()
	if err := tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, metadataTextMap(md)); err != nil {
		grpclog.Infof("grpc_opentracing: failed serializing trace information: %v", err)
	}
	ctxWithMetadata := md.ToOutgoing(ctx)
	return opentracing.ContextWithSpan(ctxWithMetadata, clientSpan), clientSpan
}

func finishClientSpan(clientSpan opentracing.Span, err error) {
	if !opentracing.IsGlobalTracerRegistered() || clientSpan == nil {
		return
	}
	if err != nil && err != io.EOF {
		ext.Error.Set(clientSpan, true)
		clientSpan.LogFields(rpclog.String("event", "error"), rpclog.String("message", err.Error()))
	}
	clientSpan.Finish()
}
