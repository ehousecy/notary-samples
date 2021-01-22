package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"github.com/golang/protobuf/proto"
	pb_msp "github.com/hyperledger/fabric-protos-go/msp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"math/big"
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
	//creator, _ := org1Admin.Serialize()
	creator, _ := createCreator()
	//2.服务端创建proposal
	fmt.Println("服务端创建proposal=================")
	request := channel.Request{ChaincodeID: "basic", Fcn: "CreateAsset", Args: GetArgs("asset4", "yellow", "5", "Tom", "1300")}
	proposal, err := client.CreateTransactionProposal(&request, creator)
	if err != nil {
		t.Fatal(err)
	}
	//3.客户端签名Proposal
	fmt.Println("客户端签名Proposal=================")
	protosalBytes, err := proto.Marshal(proposal.Proposal)
	//signedProposal := SignProposal(org1Admin, protosalBytes)
	signedProposal := SignProposalByKey(protosalBytes)
	//4.发送签名提案
	fmt.Println("发送签名提案=================")
	payloadBytes, err := client.CreateTransactionPayload(request, signedProposal)
	if err != nil {
		t.Fatal(err)
	}
	//5.客户端签名payload
	fmt.Println("客户端签名payload=================")
	//signedEnvelope, err := SignPayload(org1Admin, payloadBytes)
	signedEnvelope, err := SignPayloadByKey(payloadBytes)

	go ListenTxID(client.ChannelProvider, proposal.TxnID)

	//6.发送交易
	fmt.Println("发送交易=================")
	_, err = client.SendSignedEnvelopTx(signedEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 5)

}

func createCreator() ([]byte, error) {
	idBytes, err := ioutil.ReadFile("./Admin@org1.example.com-cert.pem")
	if err != nil {
		panic(err)
	}
	serializedIdentity := &pb_msp.SerializedIdentity{
		Mspid:   "Org1MSP",
		IdBytes: idBytes,
	}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		panic(err)
	}
	return identity, nil
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

func Sign(object []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	if len(object) == 0 {
		return nil, errors.New("object (to sign) required")
	}

	if key == nil {
		return nil, errors.New("key (for signing) required")
	}
	hash := sha256.New()
	hash.Write(object)
	digest := hash.Sum(nil)
	r, s, err := ecdsa.Sign(rand.Reader, key, digest)
	if err != nil {
		panic(err)
	}
	s, err = ToLowS(&key.PublicKey, s)
	if err != nil {
		panic(err)
	}

	signature, err := asn1.Marshal(ecdsaSignature{r, s})
	if err != nil {
		return nil, err
	}
	return signature, nil

}

func SignProposalByKey(proposalBytes []byte) *pb.SignedProposal {
	path := "./priv_sk"
	privateKey, err := loadPrivateKey(path)
	if err != nil {
		log.Fatalf("签名Proposal失败, err=%v", err)
	}

	signature, err := Sign(proposalBytes, privateKey)
	if err != nil {
		log.Fatalf("签名Proposal失败, err=%v", err)
	}
	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
}

type ecdsaSignature struct {
	R, S *big.Int
}

var (
	// curveHalfOrders contains the precomputed curve group orders halved.
	// It is used to ensure that signature' S value is lower or equal to the
	// curve group order halved. We accept only low-S signatures.
	// They are precomputed for efficiency reasons.
	curveHalfOrders = map[elliptic.Curve]*big.Int{
		elliptic.P224(): new(big.Int).Rsh(elliptic.P224().Params().N, 1),
		elliptic.P256(): new(big.Int).Rsh(elliptic.P256().Params().N, 1),
		elliptic.P384(): new(big.Int).Rsh(elliptic.P384().Params().N, 1),
		elliptic.P521(): new(big.Int).Rsh(elliptic.P521().Params().N, 1),
	}
)

// IsLow checks that s is a low-S
func IsLowS(k *ecdsa.PublicKey, s *big.Int) (bool, error) {
	halfOrder, ok := curveHalfOrders[k.Curve]
	if !ok {
		return false, fmt.Errorf("curve not recognized [%s]", k.Curve)
	}

	return s.Cmp(halfOrder) != 1, nil

}

func ToLowS(k *ecdsa.PublicKey, s *big.Int) (*big.Int, error) {
	lowS, err := IsLowS(k, s)
	if err != nil {
		return nil, err
	}

	if !lowS {
		// Set s to N - s that will be then in the lower part of signature space
		// less or equal to half order
		s.Sub(k.Params().N, s)

		return s, nil
	}

	return s, nil
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, error) {

	raw, err := ioutil.ReadFile(path)

	if err != nil {
		log.Panicf("Failed loading private key [%s]: [%s].", path, err.Error())
		return nil, err
	}
	privateKey, err := pemToPrivateKey(raw, nil)
	if err != nil {
		log.Panicf("Failed parsing private key [%s]: [%s].", path, err.Error())

		return nil, err
	}

	pKey, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		log.Printf("Failed parsing private key")
	}

	return pKey, nil
}

func pemToPrivateKey(raw []byte, pwd []byte) (interface{}, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("failed decoding PEM. Block must be different from nil [% x]", raw)
	}

	// TODO: derive from header the type of the key

	if x509.IsEncryptedPEMBlock(block) {
		if len(pwd) == 0 {
			return nil, errors.New("encrypted Key. Need a password")
		}

		decrypted, err := x509.DecryptPEMBlock(block, pwd)
		if err != nil {
			return nil, fmt.Errorf("failed PEM decryption: [%s]", err)
		}

		key, err := derToPrivateKey(decrypted)
		if err != nil {
			return nil, err
		}
		return key, err
	}

	cert, err := derToPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, err
}

func derToPrivateKey(der []byte) (key interface{}, err error) {

	if key, err = x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *ecdsa.PrivateKey:
			return
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}

	return nil, errors.New("invalid key type. The DER must contain an ecdsa.PrivateKey")
}

func SignPayload(ctx contextApi.Client, payloadBytes []byte) (*fab.SignedEnvelope, error) {

	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(payloadBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of payload failed")
	}
	return &fab.SignedEnvelope{Payload: payloadBytes, Signature: signature}, nil
}
func SignPayloadByKey(payloadBytes []byte) (*fab.SignedEnvelope, error) {
	path := "./priv_sk"
	privateKey, err := loadPrivateKey(path)
	if err != nil {
		log.Fatalf("签名Proposal失败, err=%v", err)
	}
	signature, err := Sign(payloadBytes, privateKey)
	if err != nil {
		log.Fatalf("签名Proposal失败, err=%v", err)
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
