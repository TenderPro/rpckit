package ticker

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/opentracing/opentracing-go"

	//	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	gproto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

// Run publishes time ticks to NATS channel with given subject
func Run(ctx context.Context, log *zap.Logger, nc *nats.Conn, subject string) error {
	t := time.NewTicker(time.Second * 1)
	for {
		select {
		case ts := <-t.C:
			pbts, err := types.TimestampProto(ts)
			if err != nil {
				log.Error("Ticker timestamp parse error", zap.Error(err))
				return err
			}
			msg := &TimeResponse{Ts: pbts}
			rawResponse, err := gproto.Marshal(msg)
			if err != nil {
				log.Error("Ticker Marshal error", zap.Error(err))
				return err
			}
			if err := nc.Publish(subject, rawResponse); err != nil {
				log.Error("Ticker publish error", zap.Error(err))
				return err
			}
		case <-ctx.Done():
			log.Info("Time service exited")
			return ctx.Err()
		}
	}
}

type Service struct {
	subject string
	log     *zap.Logger
	mq      *nats.Conn
}
type TimeServiceServer interface {
	Send(data *TimeResponse) error
	Context() context.Context
}

func New(log *zap.Logger, mq *nats.Conn, subject string) *Service {
	return &Service{subject: subject, log: log, mq: mq}
}

// TimeService is a gRPC service for ticker
func (p Service) TimeService(in *TimeRequest, stream TimeServiceServer) error {
	p.log.Debug("--- TimeService ---", zap.Reflect("in", in))

	ch := make(chan *nats.Msg, 64)
	sub, err := p.mq.ChanSubscribe(p.subject, ch)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	ctx := stream.Context()
	span, _ := opentracing.StartSpanFromContext(ctx, "Timer")
	if span != nil {
		defer span.Finish()
	}
	first := true
	var i int32
	for {
		select {
		case <-ctx.Done():
			p.log.Debug("client exited")
			return nil
		case msg := <-ch:
			data := TimeResponse{}
			err = gproto.Unmarshal(msg.Data, &data)
			if err != nil {
				p.log.Error("TimeService unmarshal error", zap.Error(err))
				return err
			}
			span.LogKV("event", data.Ts)

			// Ticker fired every 1 sec, but code will support any other timings
			if first || data.Ts.Seconds%int64(in.Every) == 0 {
				first = false
				i += 1
				p.log.Debug("Receive", zap.Int32("#", i), zap.Reflect("stamp", data.Ts))
				err = stream.Send(&data)
				if err != nil {
					p.log.Error("TimeService send error", zap.Error(err))
					return err
				}
			}
		}
		if in.Max > 0 && in.Max <= i {
			break
		}
	}
	return nil
}
