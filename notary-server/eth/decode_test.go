package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"math/rand"
	"time"
)

func constructTx() *types.Transaction {
	randPriv := fmt.Sprintf("%x", generatePrivKey())
	privateKey, err := crypto.HexToECDSA(randPriv)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1000000000000000000) // in wei (1 eth)
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
	var data []byte
	tx := types.NewTransaction(1, toAddress, value, 210000, big.NewInt(1), data)

	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(1)), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	return signedTx
}

//generate random bytes as account private key
func generatePrivKey() []byte {
	priv := make([]byte, 32)
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	for i := 0; i < 31; i++ {
		priv[i] = byte(r.Intn(256))
	}
	return priv
}
