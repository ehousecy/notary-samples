package main

import (
	"context"
	"github.com/ehousecy/notary-samples/notary-client/fabutil"
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

	FabricSubmitTx(client)
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

func FabricSubmitTx(client proto.NotaryServiceClient) {
	log.Println("start transfer FabricSubmitTx method=====================")
	srv, err := client.SubmitTx(context.Background())
	checkErr(err)
	//获取creator
	creator, err := fabutil.GetCreator("Org1MSP", "Admin@org1.example.com-cert.pem")
	checkErr(err)
	//获取签名
	privateKey, err := fabutil.GetPrivateKey("priv_sk")
	checkErr(err)
	err = srv.Send(&proto.TransferPropertyRequest{
		Data:        creator,
		CTxId:       "1",
		NetworkType: proto.TransferPropertyRequest_fabric,
	})
	checkErr(err)
	recv, err := srv.Recv()
	checkErr(err)
	//签名proposal
	sign, err := fabutil.Sign(recv.TxData, privateKey)
	checkErr(err)
	err = srv.Send(&proto.TransferPropertyRequest{Data: sign})
	checkErr(err)
	recv, err = srv.Recv()
	checkErr(err)
	//签名交易
	sign, err = fabutil.Sign(recv.TxData, privateKey)
	checkErr(err)
	err = srv.Send(&proto.TransferPropertyRequest{Data: sign})
	checkErr(err)
	log.Println("end transfer FabricSubmitTx method=====================")
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

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
