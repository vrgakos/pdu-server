package app

import (
	pb "pdu-server/protos"
	"context"
	"time"
	"math/rand"
	"log"
	"fmt"
)


func (a *App) ClientHello(context context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	println(in.Id, in.Name, in.Mode)

	pod := a.getPod(in.Id)
	client := pod.getClient(in.Mode)
	client.name = in.Name
	client.node = a.getNode(in.Nid)
	a.clientConnected(client)

	return &pb.HelloResponse{
		Ok: true,
		Cid: client.cid,
	}, nil
}

func (a *App) ClientCollect(context context.Context, in *pb.CollectRequest) (*pb.CollectResponse, error) {
	client := a.getClient(in.Cid)
	if client == nil {
		return &pb.CollectResponse{Ok: false}, nil
	}

	preMeasure := client.node.NewMeasureData(in.Data)

	if ! preMeasure {
		a.GetBrowserGw().BroadcastClientMeasureData(client, in)
	}

	return &pb.CollectResponse{Ok: true}, nil
}

func (a *App) NodeCollect(context context.Context, in *pb.NodeCollectRequest) (*pb.CollectResponse, error) {
	node := a.getNode(in.Nid)
	if node == nil {
		return &pb.CollectResponse{Ok: false}, nil
	}

	a.GetBrowserGw().BroadcastNodeMeasureData(node, in)

	return &pb.CollectResponse{Ok: true}, nil
}

func (a *App) WatchClientControl(in *pb.ClientControlRequest, stream pb.PduServer_WatchClientControlServer) error {
	ctx := stream.Context()

	client := a.getClient(in.Cid)
	if client == nil {
		return fmt.Errorf("Cannot find client!")
	}
	client.control = make(chan *pb.ClientControlResponse, 5)

	if client.node.imState >= 5 && client.mode == pb.ClientMode_MEASURE {
		client.setPreMeasure(false)
	}

	for i := 0; ; i++ {
		select {
		case msg := <-client.control:
			err := stream.Send(msg)
			if err != nil {
				log.Printf("WatchClientControl: send error: %s\n", err)
				return err
			}
		case <-ctx.Done():
			log.Printf("WatchClientControl: ctx done: %s\n", ctx.Err())
			a.clientDisconnected(client)
			return ctx.Err()
		}
	}
}

func (a *App) WatchNodeControl(in *pb.NodeControlRequest, stream pb.PduServer_WatchNodeControlServer) error {
	ctx := stream.Context()

	node := a.getNode(in.Id)
	node.name = in.Name
	a.nodeConnected(node)
	if node == nil {
		return fmt.Errorf("Cannot find client!")
	}

	for i := 0; ; i++ {
		d := time.Duration(2 + rand.Intn(3)) * time.Second
		select {
		case <-time.After(d):
			//err := stream.Send(&pb.NodeControlResponse{
			//	Asd: "asd asd asd",
			//})
			//if err != nil {
			//	log.Printf("WatchNodeControl: send error: %s\n", err)
			//	return err
			//}
		case <-ctx.Done():
			log.Printf("WatchNodeControl: ctx done: %s\n", ctx.Err())
			a.nodeDisconnected(node)
			return ctx.Err()
		}
	}
}

