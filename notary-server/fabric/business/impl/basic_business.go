package impl

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var accountID string
type BasicBusiness struct {
}

func (b BasicBusiness) InitSDK() (*fabsdk.FabricSDK, error) {
	home := os.Getenv("HOME")
	var filename = "config.yaml"
	appConfigPath := filepath.Join(home, ".notary-samples", filename)
	if !FileExists(appConfigPath) {
		panic("config file not exist")
	}
	ccpPath := filepath.Join(appConfigPath)
	sdk, err := fabsdk.New(config.FromFile(filepath.Clean(ccpPath)))
	if err != nil {
		return nil, err
	}
	if accountID == "" {
		b.setClientAccountID(sdk)
	}
	return sdk, err
}
func FileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}
func (b BasicBusiness) GetContextOptions() ([]fabsdk.ContextOption, error) {
	var contextOptions []fabsdk.ContextOption
	contextOptions = append(contextOptions, fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))
	return contextOptions, nil
}
func (b BasicBusiness) CreateFromRequest(chaincodeName, assetType, asset, from string) (*channel.Request, error) {
	fmt.Println("request account id:", accountID)
	request := &channel.Request{ChaincodeID: chaincodeName, Fcn: "Transfer", Args: [][]byte{[]byte(accountID),[]byte(asset)}}
	return request, nil
}
func (b BasicBusiness) CreateToRequest(chaincodeName, assetType, asset, to string) (*channel.Request, error) {
	request := &channel.Request{ChaincodeID: chaincodeName, Fcn: "Transfer", Args: [][]byte{[]byte(to),[]byte(asset)}}
	return request, nil
}

func (b BasicBusiness) ValidateEnableSupport(chaincodeName, assetType, asset string) (bool, error) {
	if chaincodeName == "basic" {
		return true, nil
	}
	return false, fmt.Errorf("fabric [%s] channel Only supports  chaincode name: [basic]", b.GetChannelID())
}

func (b BasicBusiness) GetChannelID() string {
	return "mychannel"
}

func (b BasicBusiness) CreateQueryAssertRequest(chaincodeName, assetType, asset string) (*channel.Request, error) {
	return &channel.Request{ChaincodeID: chaincodeName, Fcn: "BalanceOf", Args: [][]byte{[]byte(asset)}}, nil
}

func (b BasicBusiness) setClientAccountID(sdk *fabsdk.FabricSDK) string {
	if accountID != "" {
		return accountID
	}
	options, _ := b.GetContextOptions()
	client, _ := sdk.Context(options...)()
	block, _ := pem.Decode(client.EnrollmentCertificate())
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(fmt.Sprintf("obtain notary fabric account err:%v", err))
	}
	accountID = strings.ToUpper(cert.SerialNumber.Text(16))
	log.Printf("fabric notary account:%s",accountID)
	return accountID
}
