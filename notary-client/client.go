package main

import (
	"context"
	"github.com/ehousecy/notary-samples/proto"
	"google.golang.org/grpc"
	"log"
)

const (
	serviceAddr = ":37788"
)

func main() {
	conn, err := grpc.Dial(serviceAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := proto.NewNotaryServiceClient(conn)

	CreateCTX(client)

	SubmitTx(client)
	GetTicket(client)

}

func GetTicket(client proto.NotaryServiceClient) {
	log.Println("start transfer GetTicket method=====================")
	response, err := client.GetTicket(context.Background(), &proto.QueryTxRequest{TicketId: "1"})
	if err != nil {
		log.Fatal(err)
	}
	log.Print(response.BlockchainTxs)
	log.Println(response.Detail)
	log.Println("end transfer GetTicket method=====================")
}

func SubmitTx(client proto.NotaryServiceClient) {
	log.Println("start transfer SubmitTx method=====================")
	response, err := client.SubmitTx(context.Background(), &proto.TransferPropertyRequest{
		SignedData:  []byte("hello boy"),
		CTxId:       "123456",
		NetworkType: proto.TransferPropertyRequest_eth,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("txId=%s", response.ETxid)
	log.Println("end transfer SubmitTx method=====================")
}

func CreateCTX(client proto.NotaryServiceClient) {
	log.Println("start transfer CreateCTX method=====================")
	response, err := client.CreateCTX(context.Background(), &proto.CreateCrossTxReq{Detail: &proto.CrossTxDetail{
		EFrom:          "123",
		ETo:            "456",
		EAmount:        "789",
		FFrom:          "ffrom",
		FTo:            "fto",
		FAmount:        "fa",
		FChannel:       "channel",
		FChaincodeName: "chaincode",
	}})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("cid=%v", response.CTxId)
	log.Println("end transfer CreateCTX method=====================")
}
