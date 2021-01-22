package client

import (
	"fmt"
	"github.com/ehousecy/notary-samples/notary-server/fabric/sdkutil"
	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"log"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	client, err := New("mychannel")
	if err != nil {
		t.Fatal(err)
	}
	//1.客户端获取org1admin Creator
	fmt.Println("客户端获取org1admin Creator=================")
	org1Admin, _ := sdkutil.GetClient(client.SDK, fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	creator, _ := org1Admin.Serialize()
	//2.服务端创建proposal
	fmt.Println("服务端创建proposal=================")
	request := channel.Request{ChaincodeID: "basic", Fcn: "CreateAsset", Args: GetArgs("asset19", "yellow", "5", "Tom", "1300")}
	proposal, err := client.CreateTransactionProposal(&request, creator)
	if err != nil {
		t.Fatal(err)
	}
	//3.客户端签名Proposal
	fmt.Println("客户端签名Proposal=================")
	protosalBytes, err := proto.Marshal(proposal.Proposal)
	signedProposal := SignProposal(org1Admin, protosalBytes)
	//4.发送签名提案
	fmt.Println("发送签名提案=================")
	payloadBytes, err := client.CreateTransactionPayload(request, signedProposal)
	if err != nil {
		t.Fatal(err)
	}
	//5.客户端签名payload
	fmt.Println("客户端签名payload=================")
	signedEnvelope, err := SignPayload(org1Admin, payloadBytes)

	go ListenTxID(client.ChannelProvider, proposal.TxnID)

	//6.发送交易
	fmt.Println("发送交易=================")
	_, err = client.SendSignedEnvelopTx(signedEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 5)

}

func SignProposal(ctx contextApi.Client, proposalBytes []byte) *pb.SignedProposal {

	signingMgr := ctx.SigningManager()
	if signingMgr == nil {
		log.Fatalln("签名Proposal失败, err=signing manager is nil")
	}

	signature, err := signingMgr.Sign(proposalBytes, ctx.PrivateKey())
	if err != nil {
		log.Fatalf("签名Proposal失败, err=%v", err)
	}

	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
}

func SignPayload(ctx contextApi.Client, payloadBytes []byte) (*fab.SignedEnvelope, error) {

	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(payloadBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of payload failed")
	}
	return &fab.SignedEnvelope{Payload: payloadBytes, Signature: signature}, nil
}

func ListenTxID(ccp contextApi.ChannelProvider, txID fab.TransactionID) {
	client, _ := event.New(ccp)
	reg, events, err := client.RegisterTxStatusEvent(string(txID))
	if err != nil {
		log.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer client.Unregister(reg)

	select {
	case e, ok := <-events:
		if !ok {
			log.Fatal("unexpected closed channel")
		} else {
			if string(txID) != e.TxID {
				log.Fatalf("expecting event for TxID [%v] but received event for TxID [%v]", txID, e.TxID)
			} else if e.TxValidationCode != pb.TxValidationCode_VALID {
				log.Fatalf("expecting TxValidationCode [%v] but received [%v]", pb.TxValidationCode_VALID, e.TxValidationCode)
			}
			log.Printf("交易成功，txID: %v", txID)
		}
	}
}

func GetArgs(args ...string) [][]byte {
	bytes := make([][]byte, len(args))
	for i, v := range args {
		bytes[i] = []byte(v)
	}
	return bytes
}

func filteredBlockListener(ccp contextApi.ChannelProvider) {
	client, _ := event.New(ccp)
	reg, events, err := client.RegisterFilteredBlockEvent()
	if err != nil {
		log.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer client.Unregister(reg)
	go func() {
		for e := range events {
			log.Printf("Receive filterd block event: number=%v", e.FilteredBlock.Number)
			for i, tx := range e.FilteredBlock.FilteredTransactions {
				log.Printf("tx index %d: type: %v, txid: %v, "+
					"validation code: %v", i,
					tx.Type, tx.Txid,
					tx.TxValidationCode)
			}
			log.Println() // Just go print empty log for easy to read
		}
	}()
}
