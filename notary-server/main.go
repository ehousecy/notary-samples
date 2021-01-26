package main

import (
	"github.com/ehousecy/notary-samples/notary-server/fabric/monitor"
	"github.com/ehousecy/notary-samples/notary-server/fabric/tx"
	pb "github.com/ehousecy/notary-samples/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	servicePort = ":37788"
)

func main() {
	lis, err := net.Listen("tcp", servicePort)
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	//开启fabric区块监听
	monitor.New(tx.New()).Start()
	s := grpc.NewServer()
	ns := NewNotaryService()
	pb.RegisterNotaryServiceServer(s, ns)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
	defer s.Stop()

}
