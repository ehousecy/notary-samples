package business

import (
	"fmt"
	"github.com/ehousecy/notary-samples/notary-server/fabric/business/impl"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"sync"
)

//定义business support全局实例
var support Support
var once sync.Once

type Support struct {
	businessMap map[string]Business
}

func New() Support {
	once.Do(func() {
		support = Support{businessMap: make(map[string]Business)}
		support.register(impl.BasicBusiness{})
	})
	return support
}

func (s Support) register(b Business) {
	channelID := b.GetChannelID()
	if channelID == "" {
		panic("register channel business impl fail")
	}
	s.businessMap[channelID] = b
}

func (s Support) GetSupportChannels() []string {
	keys := make([]string, 0, len(s.businessMap))
	for k := range s.businessMap {
		keys = append(keys, k)
	}
	return keys
}

func (s Support) InitSDK(channelID string) (*fabsdk.FabricSDK, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return nil, err
	}
	return business.InitSDK()
}

func (s Support) GetContextOptions(channelID string) ([]fabsdk.ContextOption, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return nil, err
	}
	return business.GetContextOptions()
}

type RequestParams struct {
	ChaincodeName string
	AssetType     string
	Asset         string
	From          string
	To            string
}

func (s Support) CreateFromRequest(channelID string, rp RequestParams) (*channel.Request, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return nil, err
	}
	return business.CreateFromRequest(rp.ChaincodeName, rp.AssetType, rp.Asset, rp.From)
}
func (s Support) CreateToRequest(channelID string, rp RequestParams) (*channel.Request, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return nil, err
	}
	return business.CreateToRequest(rp.ChaincodeName, rp.AssetType, rp.Asset, rp.To)
}
func (s Support) ValidateEnableSupport(channelID, chaincodeName, assetType, asset string) (bool, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return false, err
	}
	return business.ValidateEnableSupport(chaincodeName, assetType, asset)
}

func (s Support) CreateQueryAssertRequest(channelID string, rp RequestParams) (*channel.Request, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return nil, err
	}
	return business.CreateQueryAssertRequest(rp.ChaincodeName, rp.AssetType, rp.Asset)
}

func (s Support) getBusiness(channelID string) (Business, error) {
	b, ok := s.businessMap[channelID]
	if !ok {
		return nil, fmt.Errorf("the specified channel is not supported, channelID=%v", channelID)
	}
	return b, nil
}
