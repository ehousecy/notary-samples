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
	th tx.Handler
}

func (fm *FabricMonitor) Start() {
	channelIDs := business.Support.GetSupportChannels()
	//1.开启区块监听
	fm.BlockEventsMonitor(channelIDs)
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
	//todo: 获取开始区块
	var startNum uint64 = 0
	eventClient, _ := event.New(ccp, event.WithSeekType(seek.FromBlock), event.WithBlockNum(startNum), event.WithBlockEvents())
	reg, events, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		log.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer eventClient.Unregister(reg)

	for e := range events {
		log.Printf("Receive filterd block event: number=%v", e.FilteredBlock.Number)
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
