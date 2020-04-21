package pkgticker

import (
	"fmt"
	"time"

	"github.com/nats-rpc/nrpc"

	gproto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	"stage1/gars/proto"
)

func Run(subject string, nc nrpc.NatsConn, exitChan chan struct{}) error {
	t := time.NewTicker(time.Second * 1)
	go func() {
		for {
			select {
			case ts := <-t.C:
				pbts, err := types.TimestampProto(ts)
				if err != nil {
					fmt.Printf("err10: %v\n", err)
					//				return err
				}
				msg := &proto.TimeResponse{Ts: pbts}
				rawResponse, err := gproto.Marshal(msg)
				if err != nil {
					fmt.Printf("err11: %v\n", err)
					//					return err
				}
				if err := nc.Publish(subject, rawResponse); err != nil {
					fmt.Printf("err112 %v\n", err)
					return
				}
			case <-exitChan:
				fmt.Printf("Time service exited\n")
				return
			}
		}
	}()
	return nil
}
