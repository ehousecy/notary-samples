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
	if err = ctd.ValidateHostingStatus(); err != nil {
		return err
	}
	existed, err := model.ValidateExistedValidTxDetailCIDAndType(cid, txType)
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
	td := model.NewTransferFromTx(ctd, txType, txID)
	if _, err = td.Save(tx); err != nil {
		return err
	}
	if err = ctd.CreateTransferFromTxInfo(txID, txType, tx); err != nil {
		return err
	}

	//5.提交事务
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (cts CrossTxDataServiceProvider) ValidateEnableCompleteTransferFromTx(cidStr string, txID string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByFromTxID(cidStr, txID)
	if err != nil {
		return err
	}
	return validateEnableCompleteTransferFromTx(ctd, td)
}

func (cts CrossTxDataServiceProvider) CompleteTransferFromTx(cidStr string, txID string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByFromTxID(cidStr, txID)
	if err != nil {
		return err
	}
	//2.校验需要完成的交易状态
	//todo:交易状态替换
	if td.TxStatus != constant.TxStatusFromCreated {
		return errors.New("当前交易不能完成")
	}
	if err = ctd.ValidateHostingStatus(); err != nil {
		return err
	}
	//3.开启事务完成交易
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = td.CompleteTransferFromTx(tx); err != nil {
		return err
	}
	if err = ctd.CompleteTransferFromTx(txID, td.Type, tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (cts CrossTxDataServiceProvider) ValidateEnableBoundTransferToTx(cidStr string, boundTxID string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByFromTxID(cidStr, boundTxID)
	if err != nil {
		return err
	}
	return validateEnableBoundTransferToTx(ctd, td)
}

func (cts CrossTxDataServiceProvider) BoundTransferToTx(cidStr string, boundTxID, txID string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByFromTxID(cidStr, boundTxID)
	if err != nil {
		return err
	}
	//2.校验需要绑定的交易状态
	if err = validateEnableBoundTransferToTx(ctd, td); err != nil {
		return err
	}
	//3.开启事务
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = td.BoundTransferToTx(txID, tx); err != nil {
		return err
	}
	if err = ctd.BoundTransferToTxInfo(txID, td.Type, tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (cts CrossTxDataServiceProvider) ValidateEnableCompleteTransferToTx(cidStr string, txID string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByToTxID(cidStr, txID)
	if err != nil {
		return err
	}
	return validateEnableCompleteTransferToTx(ctd, td)
}

func (cts CrossTxDataServiceProvider) CompleteTransferToTx(cidStr string, txID string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByToTxID(cidStr, txID)
	if err != nil {
		return err
	}
	if err = validateEnableCompleteTransferToTx(ctd, td); err != nil {
		return err
	}

	//开启事务完成交易
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = td.CompleteTransferToTx(tx); err != nil {
		return err
	}

	if err = ctd.CompleteTransferToTx(txID, td.Type, tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
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
			if td.TxStatus != constant.TxStatusFromFailed {
				if td.Type == constant.TypeFabric {
					info.FabricTx = convertTxDetail(*td)
				} else if td.Type == constant.TypeEthereum {
					info.EthTx = convertTxDetail(*td)
				}
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

func getCrossTxDetailAndTxDetailByFromTxID(cidStr string, txID string) (*model.CrossTxDetail, *model.TxDetail, error) {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return nil, nil, errors.New("cid 异常")
	}
	//1.查询跨链交易和需要绑定的交易
	ctd, err := model.GetCrossTxDetailByID(cid)
	if err != nil {
		return nil, nil, err
	}
	td, err := model.GetCrossTxByFromTxID(txID, cid)
	return ctd, td, err
}

func getCrossTxDetailAndTxDetailByToTxID(cidStr string, txID string) (*model.CrossTxDetail, *model.TxDetail, error) {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return nil, nil, errors.New("cid 异常")
	}
	//1.查询跨链交易和需要绑定的交易
	ctd, err := model.GetCrossTxDetailByID(cid)
	if err != nil {
		return nil, nil, err
	}
	td, err := model.GetCrossTxByToTxID(txID, cid)
	return ctd, td, err
}

func validateEnableCompleteTransferFromTx(ctd *model.CrossTxDetail, td *model.TxDetail) error {
	if td.TxStatus != constant.TxStatusToCreated || ctd.Status != constant.StatusCreated {
		return errors.New("当前交易不能代理转账")
	}
	return nil
}

func validateEnableBoundTransferToTx(ctd *model.CrossTxDetail, td *model.TxDetail) error {
	if td.TxStatus != constant.TxStatusFromFinished || ctd.Status != constant.StatusHosted {
		return errors.New("当前交易不能代理转账")
	}
	return nil
}

func validateEnableCompleteTransferToTx(ctd *model.CrossTxDetail, td *model.TxDetail) error {
	if td.TxStatus != constant.TxStatusToCreated || ctd.Status != constant.StatusHosted {
		return errors.New("当前交易不能代理转账")
	}
	return nil
}
