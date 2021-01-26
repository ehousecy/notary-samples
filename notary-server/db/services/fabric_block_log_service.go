package services

import "github.com/ehousecy/notary-samples/notary-server/db/model"

type FabricBlockLogService interface {
	AddFabricBlockLog(blockNumber uint64, channelID string)
	QueryLastFabricBlockNumber(channelID string) (uint64, error)
}

type FabricBlockLogServiceProvider struct {
}

func NewFabricBlockLogServiceProvider() FabricBlockLogServiceProvider {
	return FabricBlockLogServiceProvider{}
}

func (f FabricBlockLogServiceProvider) AddFabricBlockLog(blockNumber uint64, channelID string) {
	model.InsertFabricBlockLog(blockNumber, channelID)
}

func (f FabricBlockLogServiceProvider) QueryLastFabricBlockNumber(channelID string) (uint64, error) {
	return model.QueryLastFabricBlockNumber(channelID)
}
