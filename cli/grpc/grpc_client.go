package grpc

import (
	"github.com/ehousecy/notary-samples/proto"
	"google.golang.org/grpc"
	"log"
)

const (
	address     = "localhost:50051"
	defaultName = "notary-service"
)

func NewClient() proto.NotaryServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	c := proto.NewNotaryServiceClient(conn)
	return c
}
