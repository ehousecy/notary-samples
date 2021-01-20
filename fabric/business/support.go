package business

import (
	"fmt"
	"github.com/ehousecy/notary-samples/fabric/business/impl"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"sync"
)

func init() {
	once.Do(func() {
		Support = support{businessMap: make(map[string]Business)}
		Support.Register(impl.BasicBusiness{})
	})
}

//定义business support全局实例
var Support support
var once sync.Once

type support struct {
	businessMap map[string]Business
}

func (s support) Register(b Business) {
	channelID := b.GetChannelID()
	if channelID == "" {
		panic("register channel business impl fail")
	}
	s.businessMap[channelID] = b
}

func (s support) GetSupportChannels() []string {
	keys := make([]string, 0, len(s.businessMap))
	for k := range s.businessMap {
		keys = append(keys, k)
	}
	return keys
}

func (s support) InitSDK(channelID string) (*fabsdk.FabricSDK, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return nil, err
	}
	return business.InitSDK()
}

func (s support) GetContextOptions(channelID string) ([]fabsdk.ContextOption, error) {
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

func (s support) CreateFromRequest(channelID string, rp RequestParams) (*channel.Request, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return nil, err
	}
	return business.CreateFromRequest(rp.ChaincodeName, rp.AssetType, rp.Asset, rp.From)
}
func (s support) CreateToRequest(channelID string, rp RequestParams) (*channel.Request, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return nil, err
	}
	return business.CreateToRequest(rp.ChaincodeName, rp.AssetType, rp.Asset, rp.To)
}
func (s support) ValidateEnableSupport(channelID, chaincodeName, assetType, asset string) (bool, error) {
	business, err := s.getBusiness(channelID)
	if err != nil {
		return false, err
	}
	return business.ValidateEnableSupport(chaincodeName, assetType, asset)
}

func (s support) getBusiness(channelID string) (Business, error) {
	b, ok := s.businessMap[channelID]
	if !ok {
		return nil, fmt.Errorf("the specified channel is not supported, channelID=%v", channelID)
	}
	return b, nil
}
