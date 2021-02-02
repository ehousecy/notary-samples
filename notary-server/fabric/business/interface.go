package business

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type Business interface {
	InitSDK() (*fabsdk.FabricSDK, error)
	GetContextOptions() ([]fabsdk.ContextOption, error)
	CreateFromRequest(chaincodeName, assetType, asset, from string) (*channel.Request, error)
	CreateToRequest(chaincodeName, assetType, asset, to string) (*channel.Request, error)
	GetChannelID() string
	ValidateEnableSupport(chaincodeName, assetType, asset string) (bool, error)
	CreateQueryAssertRequest(chaincodeName, assetType, asset string) (*channel.Request, error)
}
