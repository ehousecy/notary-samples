package model

import "testing"

func TestCrossTxDetail_Save(t *testing.T) {
	ctxBase := BaseCrossTxDetail{
		ID:              4,
		EthFrom:         "123",
		EthTo:           "456",
		EthAmount:       "789",
		FabricFrom:      "ffrom",
		FabricTo:        "fto",
		FabricAmount:    "fa",
		FabricChannel:   "channel",
		FabricChaincode: "chaincode",
	}
	detail := CrossTxDetail{BaseCrossTxDetail: ctxBase}
	save, err := detail.Save()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("insert id=%v", save)
}
