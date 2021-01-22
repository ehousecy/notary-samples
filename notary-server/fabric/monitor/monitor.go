package monitor

import (
	"github.com/ehousecy/notary-samples/notary-server/fabric/business"
	"github.com/ehousecy/notary-samples/notary-server/fabric/client"
	"github.com/ehousecy/notary-samples/notary-server/fabric/tx"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"log"
)

type FabricMonitor struct {
	th         tx.Handler
	channelIDs []string
}

func New(th tx.Handler) *FabricMonitor {
	var fm = &FabricMonitor{th: th, channelIDs: business.New().GetSupportChannels()}
	return fm
}

func (fm *FabricMonitor) Start() {
	//todo:开始监听前,确保监听的交易id写入map
	//1.开启区块监听
	fm.BlockEventsMonitor(fm.channelIDs)
}

func (fm *FabricMonitor) BlockEventsMonitor(channelIDs []string) {
	//获取所有通道
	for _, channelID := range channelIDs {
		c, err := client.New(channelID)
		if err != nil {
			log.Fatalf("fabric BlockEvents monitor start failed: init config failed, channelID=%v, err=%v", channelID, err)
		}
		go fm.filteredBlockListener(channelID, c.ChannelProvider)
	}

}

func (fm *FabricMonitor) filteredBlockListener(channelID string, ccp contextApi.ChannelProvider) {

	blockNumber, err := fm.th.QueryLastFabricBlockNumber(channelID)
	if err != nil {
		log.Fatalf("fail registering for TxStatus events, get start block number fail, err: %s", err)
	}
	blockNumber++
	eventClient, _ := event.New(ccp, event.WithSeekType(seek.FromBlock), event.WithBlockNum(blockNumber), event.WithBlockEvents())
	reg, events, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		log.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer eventClient.Unregister(reg)

	for e := range events {
		log.Printf("Receive filterd block event: blockNumber=%v", e.FilteredBlock.Number)
		for i, transaction := range e.FilteredBlock.FilteredTransactions {
			log.Printf("transaction index %d: type: %v, txid: %v, "+
				"validation code: %v", i,
				transaction.Type, transaction.Txid,
				transaction.TxValidationCode)
		}
		go fm.th.HandleTxStatusBlock(channelID, e.FilteredBlock)
		log.Println() // Just go print empty log for easy to read
	}
}
