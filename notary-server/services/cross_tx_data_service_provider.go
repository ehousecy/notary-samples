package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ehousecy/notary-samples/notary-server/constant"
	"github.com/ehousecy/notary-samples/notary-server/model"
	"strconv"
)

type CrossTxDataServiceProvider struct {
}

func NewCrossTxDataServiceProvider() CrossTxDataServiceProvider {
	return CrossTxDataServiceProvider{}
}

func (cts CrossTxDataServiceProvider) CreateCrossTx(ctxBase CrossTxBase) (string, error) {

	ctd := convertCrossTxBase2CrossTxDetail(ctxBase)
	ctd.Status = "created"
	cid, err := ctd.Save()
	return int64ToString(cid), err
}

func (cts CrossTxDataServiceProvider) CreateTransferFromTx(cidStr string, txID string, txType string) error {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return errors.New("cid 异常")
	}
	//1.判断ctxID是否存在
	ctd, err := model.GetCrossTxDetailByID(cid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("跨链交易不存在")
		}
		return err
	}
	//2.判断cID对应的相同type TX是否存在
	td := NewTransferFromTx(ctd, txType, txID)
	existed, err := td.VerifyExistedValidCIDAndType()
	if err != nil {
		return fmt.Errorf("创建转账交易失败,err:%v", err)
	}
	if existed {
		return errors.New("交易已存在")
	}

	//3.开启事务
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//4.创建tx，更新ctx
	_, err = td.Save(tx)
	if err != nil {
		return err
	}
	err = ctd.CreateTransferFromTxInfo(txID, txType, tx)
	if err != nil {
		return err
	}

	//5.提交事务
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func NewTransferFromTx(ctd *model.CrossTxDetail, txType string, txID string) model.TxDetail {
	amount := ctd.FabricAmount
	if txType == constant.TypeEthereum {
		amount = ctd.EthAmount
	}
	td := model.TxDetail{
		BaseTxDetail: model.BaseTxDetail{
			TxFrom:    txID,
			Amount:    amount,
			TxStatus:  "created",
			Type:      txType,
			CrossTxID: ctd.ID,
		},
	}
	return td
}

func (cts CrossTxDataServiceProvider) CompleteTransferFromTx(cidStr string, txID string) error {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return errors.New("cid 异常")
	}
	//1.查询跨链交易和需要完成的交易
	ctd, err := model.GetCrossTxDetailByID(cid)
	if err != nil {
		return err
	}
	td, err := model.GetCrossTxByFromTxID(txID, cid)
	if err != nil {
		return err
	}
	//2.校验需要完成的交易状态
	//todo:交易状态替换
	if td.TxStatus != "..." {
		return errors.New("当前交易不能完成")
	}
	//3.开启事务完成交易
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = td.CompleteTransferFromTx(tx)
	if err != nil {
		return err
	}
	err = ctd.CompleteTransferFromTx(txID, td.Type, tx)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (cts CrossTxDataServiceProvider) BoundTransferToTx(cidStr string, boundTxID, txID string) error {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return errors.New("cid 异常")
	}
	//1.查询跨链交易和需要绑定的交易
	ctd, err := model.GetCrossTxDetailByID(cid)
	if err != nil {
		return err
	}
	td, err := model.GetCrossTxByFromTxID(boundTxID, cid)
	if err != nil {
		return err
	}
	//2.校验需要绑定的交易状态
	if td.TxStatus != "..." {
		return errors.New("当前交易不能代理转账")
	}

	//3.开启事务
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = td.BoundTransferToTx(txID, tx)
	if err != nil {
		return err
	}

	err = ctd.BoundTransferToTxInfo(txID, td.Type, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (cts CrossTxDataServiceProvider) CompleteTransferToTx(cidStr string, txID string) error {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return errors.New("cid 异常")
	}
	//1.查询跨链交易和需要完成的交易
	ctd, err := model.GetCrossTxDetailByID(cid)
	if err != nil {
		return err
	}
	td, err := model.GetCrossTxByFromTxID(txID, cid)
	if err != nil {
		return err
	}

	//开启事务完成交易
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = td.CompleteTransferToTx(tx)
	if err != nil {
		return err
	}
	err = ctd.CompleteTransferFromTx(txID, td.Type, tx)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
func (cts CrossTxDataServiceProvider) QueryCrossTxInfoByCID(cidStr string) (*CrossTxInfo, error) {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return nil, errors.New("cid 异常")
	}
	ctd, err := model.GetCrossTxDetailByID(cid)
	if err != nil {
		return nil, err
	}
	tds, err := model.GetTxDetailByCTxID(cid)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	info := convertCrossTxInfo(*ctd, tds...)
	return &info, nil
}

func (cts CrossTxDataServiceProvider) QueryAllCrossTxInfo() ([]CrossTxInfo, error) {
	ctds, err := model.GetCrossTxDetails()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var cIDs []int64
	for _, ctd := range ctds {
		cIDs = append(cIDs, ctd.ID)
	}
	tds, err := model.GetTxDetailByCTxID(cIDs...)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	m := make(map[int64][]*model.TxDetail)
	for _, td := range tds {
		m[td.CrossTxID] = append(m[td.CrossTxID], td)
	}
	var infos []CrossTxInfo
	for _, ctd := range ctds {
		infos = append(infos, convertCrossTxInfo(*ctd, m[ctd.ID]...))
	}
	return infos, nil
}

func convertCrossTxBase2CrossTxDetail(ctb CrossTxBase) model.CrossTxDetail {
	return model.CrossTxDetail{
		BaseCrossTxDetail: model.BaseCrossTxDetail{
			EthFrom:         ctb.EthFrom,
			EthTo:           ctb.EthTo,
			EthAmount:       ctb.EthAmount,
			FabricFrom:      ctb.FabricFrom,
			FabricTo:        ctb.FabricTo,
			FabricAmount:    ctb.FabricAmount,
			FabricChannel:   ctb.FabricChannel,
			FabricChaincode: ctb.FabricChaincode,
		},
	}
}

func convertCrossTxInfo(ctd model.CrossTxDetail, tds ...*model.TxDetail) CrossTxInfo {
	info := CrossTxInfo{
		CrossTxBase: CrossTxBase{
			ID:              int64ToString(ctd.ID),
			EthFrom:         ctd.EthFrom,
			EthTo:           ctd.EthTo,
			EthAmount:       ctd.EthAmount,
			FabricFrom:      ctd.FabricFrom,
			FabricTo:        ctd.FabricTo,
			FabricAmount:    ctd.FabricAmount,
			FabricChannel:   ctd.FabricChannel,
			FabricChaincode: ctd.FabricChaincode,
		},
		Status:    ctd.Status,
		CreatedAt: ctd.CreatedAt,
		UpdatedAt: ctd.UpdatedAt,
	}
	if len(tds) != 0 {
		for _, td := range tds {
			if td.Type == constant.TypeFabric {
				info.FabricTx = convertTxDetail(*td)
			} else if td.Type == constant.TypeEthereum {
				info.EthTx = convertTxDetail(*td)
			}
		}
	}
	return info
}

func convertTxDetail(td model.TxDetail) *TxDetail {
	return &TxDetail{
		TxFrom:         td.TxFrom,
		TxTo:           td.TxTo,
		Amount:         td.Amount,
		TxStatus:       td.TxStatus,
		Type:           td.Type,
		CrossTxID:      int64ToString(td.CrossTxID),
		FromTxID:       td.FromTxID,
		ToTxID:         td.ToTxID,
		FromTxCreateAt: td.FromTxCreateAt,
		ToTxCreateAt:   td.ToTxCreateAt,
		FromTxFinishAt: td.FromTxFinishAt,
		ToTxFinishAt:   td.ToTxFinishAt,
	}
}

func int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func stringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
