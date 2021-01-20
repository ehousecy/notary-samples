package model

import "testing"

func TestInsertFabricBlockLog(t *testing.T) {
	InsertFabricBlockLog(1, "mychannel")
	InsertFabricBlockLog(3, "mychannel")
	InsertFabricBlockLog(2, "mychannel")
}

func TestQueryLastFabricBlockNumber(t *testing.T) {
	blockNumber, err := QueryLastFabricBlockNumber("mychannel")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("query result blockNumber=%v", blockNumber)
}
