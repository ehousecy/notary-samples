package impl

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type BasicBusiness struct {
}

func (b BasicBusiness) InitSDK() (*fabsdk.FabricSDK, error) {
	home := os.Getenv("HOME")
	var filename = "config.yaml"
	appConfigPath := filepath.Join(home, ".notary-samples", filename)
	fmt.Println(appConfigPath)
	if !FileExists(appConfigPath) {
		panic("config file not exist")
	}
	ccpPath := filepath.Join(appConfigPath)
	return fabsdk.New(config.FromFile(filepath.Clean(ccpPath)))
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
	//todo: 添加业务逻辑
	request := &channel.Request{ChaincodeID: chaincodeName, Fcn: "TransferAsset", Args: [][]byte{[]byte(asset), []byte("Tom")}}
	return request, nil
}
func (b BasicBusiness) CreateToRequest(chaincodeName, assetType, asset, to string) (*channel.Request, error) {
	//todo: 添加业务逻辑
	request := &channel.Request{ChaincodeID: chaincodeName, Fcn: "TransferAsset", Args: [][]byte{[]byte(asset), []byte(to)}}
	return request, nil
}

func (b BasicBusiness) ValidateEnableSupport(chaincodeName, assetType, asset string) (bool, error) {
	//todo: 添加业务逻辑
	return true, nil
}

func (b BasicBusiness) GetChannelID() string {
	return "mychannel"
}

func (b BasicBusiness)  CreateQueryAssertRequest(chaincodeName, assetType, asset string) (*channel.Request, error) {
	return &channel.Request{ChaincodeID: chaincodeName, Fcn: "ReadAsset", Args: [][]byte{[]byte(asset)}}, nil
}
