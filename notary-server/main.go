package main

import (
	pb "github.com/ehousecy/notary-samples/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	servicePort = "377880"
)

func main()  {
	lis, err := net.Listen("tcp", servicePort)
	if err != nil {
		panic(err)
	}
	defer  lis.Close()
	s := grpc.NewServer()
	ns := NewNotaryService()
	pb.RegisterNotaryServiceServer(s, ns)
	if err := s.Serve(lis); err != nil{
		log.Fatalf("Failed to serve %v", err)
	}
	defer  s.Stop()

}