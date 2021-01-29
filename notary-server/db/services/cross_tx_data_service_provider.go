package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ehousecy/notary-samples/notary-server/db/constant"
	"github.com/ehousecy/notary-samples/notary-server/db/model"
	"github.com/jmoiron/sqlx"
	"log"
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

func (cts CrossTxDataServiceProvider) ValidateEnableCreateTransferFromTx(cidStr string, txType string) error {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return errors.New("invalid cross-chain transaction id")
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
	return nil
}

func (cts CrossTxDataServiceProvider) CreateTransferFromTx(cidStr string, txID string, txType string) error {
	if err := cts.ValidateEnableCreateTransferFromTx(cidStr, txType); err != nil {
		return err
	}
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return errors.New("invalid cross-chain transaction id")
	}
	//1.获取跨链交易
	ctd, err := model.GetCrossTxDetailByID(cid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("跨链交易不存在")
		}
		return err
	}

	//3.开启事务
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

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

func (cts CrossTxDataServiceProvider) ValidateEnableBoundTransferToTx(boundTxID string, cIDChan chan string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByTxID(boundTxID)
	if err != nil {
		return err
	}
	if err = validateEnableBoundTransferToTx(ctd, td); err == nil && cIDChan != nil {
		cIDChan <- int64ToString(ctd.ID)
	}
	return err
}

func (cts CrossTxDataServiceProvider) BoundTransferToTx(boundTxID, txID string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByTxID(boundTxID)
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
	defer rollbackTx(tx)

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

func (cts CrossTxDataServiceProvider) QueryCrossTxInfoByCID(cidStr string) (*CrossTxInfo, error) {
	cid, err := stringToInt64(cidStr)
	if err != nil {
		return nil, errors.New("invalid cross-chain transaction id")
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

func (cts CrossTxDataServiceProvider) QueryConfirmingTxInfo(txType string) ([]ConfirmingTxInfo, error) {
	tds, err := model.GetConfirmingTxDetailByType(txType)
	if err != nil {
		log.Printf("failed Query Confirming Tx Info, err=%v", err)
		return nil, err
	}
	if len(tds) < 1 {
		return nil, nil
	}
	ctis := make([]ConfirmingTxInfo, 0, len(tds))
	for _, td := range tds {
		ctis = append(ctis, convert2ConfirmingTxInfo(td))
	}
	return ctis, nil
}

func (cts CrossTxDataServiceProvider) CompleteTransferTx(txID string) error {
	td, err := model.GetTxDetailByTxID(txID)
	if err != nil {
		return err
	}
	ctd, err := model.GetCrossTxDetailByID(td.CrossTxID)
	if err != nil {
		return err
	}
	if ctd.Status == constant.StatusCreated {
		//from
		err = completeTransferFromTx(td, ctd, txID)
	} else if ctd.Status == constant.StatusHosted {
		//to
		err = completeTransferToTx(ctd, td, txID)
	}
	return err
}

func (cts CrossTxDataServiceProvider) ValidateEnableCompleteTransferTx(txID string) error {
	ctd, td, err := getCrossTxDetailAndTxDetailByTxID(txID)
	if err != nil {
		return err
	}
	if td.TxStatus == constant.TxStatusFromCreated && ctd.Status == constant.StatusCreated {
		//from
		return nil
	} else if td.TxStatus == constant.TxStatusToCreated && ctd.Status == constant.StatusHosted {
		//to
		return nil
	}
	return fmt.Errorf("the tx unable complete, tx=%s, cid=%v", txID, ctd.ID)
}

func (cts CrossTxDataServiceProvider) CancelTransferTx(txID string) {
	td, err := model.GetTxDetailByTxID(txID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("取消交易失败, txID=%v, err=%v", txID, err)
		}
		return
	}
	ctd, err := model.GetCrossTxDetailByID(td.CrossTxID)
	if err != nil {
		log.Printf("取消交易失败, txID=%v, err=%v", txID, err)
	}
	if ctd.Status == constant.StatusCreated {
		//from
		if err = cancelTransferFromTx(ctd, td, txID); err != nil {
			log.Printf("取消交易[%s]失败, err=%v", txID, err)
		}
	} else if ctd.Status == constant.StatusHosted {
		//to
		if err = cancelTransferToTx(ctd, td, txID); err != nil {
			log.Printf("取消代理交易[%s]失败, err=%v", txID, err)
		}
	}

}

func (cts CrossTxDataServiceProvider) FailTransferTx(txID string) {
	td, err := model.GetTxDetailByTxID(txID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("更新无效交易失败, txID=%v, err=%v", txID, err)
		}
		return
	}
	ctd, err := model.GetCrossTxDetailByID(td.CrossTxID)
	if err != nil {
		log.Printf("更新无效交易失败, txID=%v, err=%v", txID, err)
	}
	if ctd.Status == constant.StatusCreated {
		//from
		if err = failTransferFromTx(ctd, td, txID); err != nil {
			log.Printf("更新无效交易[%s]失败, err=%v", txID, err)
		}
	} else if ctd.Status == constant.StatusHosted {
		//to
		if err = failTransferToTx(ctd, td, txID); err != nil {
			log.Printf("更新无效代理交易[%s]失败, err=%v", txID, err)
		}
	}
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

func convert2ConfirmingTxInfo(td *model.TxDetail) ConfirmingTxInfo {
	var cti ConfirmingTxInfo
	cti.ChannelID = td.ChannelID
	cti.ID = int64ToString(td.CrossTxID)
	if td.TxStatus == constant.TxStatusFromCreated {
		cti.TxID = td.FromTxID
		cti.IsOfflineTx = true
	} else {
		cti.TxID = td.ToTxID
		cti.IsOfflineTx = false
	}
	return cti
}

func int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func stringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func getCrossTxDetailAndTxDetailByTxID(txID string) (*model.CrossTxDetail, *model.TxDetail, error) {
	td, err := model.GetTxDetailByTxID(txID)
	if err != nil {
		return nil, nil, err
	}
	//1.查询跨链交易和需要绑定的交易
	ctd, err := model.GetCrossTxDetailByID(td.CrossTxID)
	if err != nil {
		return nil, nil, err
	}
	return ctd, td, err
}

func validateEnableCompleteTransferFromTx(ctd *model.CrossTxDetail, td *model.TxDetail) error {
	if td.TxStatus != constant.TxStatusFromCreated || ctd.Status != constant.StatusCreated {
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

func completeTransferFromTx(td *model.TxDetail, ctd *model.CrossTxDetail, txID string) error {
	//2.校验需要完成的交易状态
	if td.TxStatus != constant.TxStatusFromCreated {
		return errors.New("当前交易不能完成")
	}
	if err := ctd.ValidateHostingStatus(); err != nil {
		return err
	}
	//3.开启事务完成交易
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

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
func completeTransferToTx(ctd *model.CrossTxDetail, td *model.TxDetail, txID string) error {
	if err := validateEnableCompleteTransferToTx(ctd, td); err != nil {
		return err
	}

	//开启事务完成交易
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

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

func cancelTransferFromTx(ctd *model.CrossTxDetail, td *model.TxDetail, txID string) error {
	//校验状态
	if td.TxStatus != constant.TxStatusFromCreated && ctd.Status != constant.StatusCreated {
		return errors.New("取消交易失败，状态校验失败")
	}
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)
	if err := td.Delete(tx); err != nil {
		return fmt.Errorf("取消交易失败: %v", err.Error())
	}
	if ctd.FabricFromTxID == txID {
		ctd.FabricFromTxID = ""
	} else if ctd.EthFromTxID == txID {
		ctd.EthFromTxID = ""
	}
	if err = ctd.Update(tx); err != nil {
		return fmt.Errorf("取消交易失败: %v", err.Error())
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func cancelTransferToTx(ctd *model.CrossTxDetail, td *model.TxDetail, txID string) error {
	//校验状态
	if td.TxStatus != constant.TxStatusToCreated && ctd.Status != constant.StatusHosted {
		return errors.New("取消交易失败，状态校验失败")
	}
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)
	if ctd.FabricToTxID == txID {
		ctd.FabricToTxID = ""
	} else if ctd.EthToTxID == txID {
		ctd.EthToTxID = ""
	}
	if err = ctd.Update(tx); err != nil {
		return fmt.Errorf("取消交易失败: %v", err.Error())
	}
	td.TxStatus = constant.TxStatusFromFinished
	td.ToTxID = ""
	if err = td.Update(tx); err != nil {
		return fmt.Errorf("取消交易失败: %v", err.Error())
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func failTransferFromTx(ctd *model.CrossTxDetail, td *model.TxDetail, txID string) error {
	//校验状态
	if td.TxStatus != constant.TxStatusFromCreated && ctd.Status != constant.StatusCreated {
		return errors.New("更新无效交易失败，状态校验失败")
	}
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)
	td.TxStatus = constant.TxStatusFromFailed
	if err := td.Update(tx); err != nil {
		return fmt.Errorf("更新无效交易失败: %v", err.Error())
	}
	if ctd.FabricFromTxID == txID {
		ctd.FabricFromTxID = ""
	} else if ctd.EthFromTxID == txID {
		ctd.EthFromTxID = ""
	}
	if err := ctd.Update(tx); err != nil {
		return fmt.Errorf("更新无效交易失败: %v", err.Error())
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func failTransferToTx(ctd *model.CrossTxDetail, td *model.TxDetail, txID string) error {
	if td.TxStatus != constant.TxStatusToCreated && ctd.Status != constant.StatusHosted {
		return errors.New("更新无效代理交易失败，状态校验失败")
	}
	tx, err := model.DB.Beginx()
	if err != nil {
		return err
	}
	defer rollbackTx(tx)
	if ctd.FabricToTxID == txID {
		ctd.FabricToTxID = ""
	} else if ctd.EthToTxID == txID {
		ctd.EthToTxID = ""
	}
	if err = ctd.Update(tx); err != nil {
		return fmt.Errorf("更新无效代理交易失败: %v", err.Error())
	}
	td.TxStatus = constant.TxStatusFromFinished
	td.ToTxID = ""
	if err = td.Update(tx); err != nil {
		return fmt.Errorf("更新无效代理交易失败: %v", err.Error())
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func rollbackTx(tx *sqlx.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.Printf("unable to rollback: %v", err)
	}
}
