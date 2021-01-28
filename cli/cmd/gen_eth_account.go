package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"math/rand"
	"time"
)

var genAccountCmd = &cobra.Command{
	Use: "gen-account",
	Run: execGenAccountCmd,
}

func execGenAccountCmd(cmd *cobra.Command, args []string)  {
	priv := genPrivateKey()
	address := crypto.PubkeyToAddress(priv.PublicKey).Hex()
	privString := fmt.Sprintf("%x", priv.D)
	fmt.Println("Generated Account:")
	fmt.Println("Address: ", address)
	fmt.Println("Private Key: ",privString)
}

func genPrivateKey() *ecdsa.PrivateKey {
	priv := make([]byte, 32)
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	for i := 0; i < 31; i ++ {
		priv[i] = byte(r.Intn(256))
	}
	return  crypto.ToECDSAUnsafe(priv)
}