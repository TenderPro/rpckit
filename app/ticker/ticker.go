package ticker

import (
	"context"
	"fmt"
	"time"

	//	"github.com/nats-io/go-nats"
	"github.com/nats-rpc/nrpc"
	//	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	gproto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

// Run publishes time ticks to NATS channel with given subject
func Run(ctx context.Context, log *zap.Logger, nc nrpc.NatsConn, subject string) error {
	t := time.NewTicker(time.Second * 1)
	for {
		select {
		case ts := <-t.C:
			pbts, err := types.TimestampProto(ts)
			if err != nil {
				fmt.Printf("err10: %v\n", err)
				//				return err
			}
			msg := &TimeResponse{Ts: pbts}
			rawResponse, err := gproto.Marshal(msg)
			if err != nil {
				fmt.Printf("err11: %v\n", err)
				//					return err
			}
			if err := nc.Publish(subject, rawResponse); err != nil {
				fmt.Printf("err112 %v\n", err)
				return err
			}
		case <-ctx.Done():
			log.Info("Time service exited")
			return ctx.Err()
		}
	}
}

/*
// TimeService is a gRPC service for ticker
func (p Service) TimeService(in *proto.TimeRequest, stream proto.TestService_TimeServiceServer) error {
	p.log.Debug("--- TimeService ---")
	i := 0

	ch := make(chan *nats.Msg, 64)
	sub, err := p.mq.ChanSubscribe(p.subject, ch)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	ctx := stream.Context()

	span, _ := opentracing.StartSpanFromContext(ctx, "Timer")
	if span != nil {
		fmt.Printf("TimeEvent: %+v\n", span)
		defer span.Finish()
	}
	first := true
	for {
		select {
		case <-ctx.Done():
			p.log.Debug("client exited")
			return nil
		case msg := <-ch:
			//	p.mq.Subscribe(subj, func(msg *nats.Msg) {
			i += 1
			p.log.Debug("Receive", zap.Int("#", i), zap.String("subject", msg.Subject))
			span.LogKV("event", i)

			data := &proto.TimeResponse{}
			err = gproto.Unmarshal(msg.Data, data)
			if err != nil {
				fmt.Printf("err01: %v\n", err)
				return err
			}

			// Ticker fired every 1 sec, but code will support any other timings
			if first || data.Ts.Seconds%int64(in.Every) == 0 {
				first = false
				err = stream.Send(data)
				if err != nil {
					fmt.Printf("err02: %v\n", err)
					return err
				}
			}
		}
	}
	//	})
	return nil
}
*/
