package util

import (
	"github.com/vitelabs/go-vite/common/fork"
	"github.com/vitelabs/go-vite/common/helper"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"math/big"
)

type dbInterface interface {
	GetBalance(tokenTypeId *types.TokenTypeId) (*big.Int, error)
	SetBalance(tokenTypeId *types.TokenTypeId, amount *big.Int)

	GetValue(key []byte) ([]byte, error)
	SetValue(key []byte, value []byte) error

	GetContractMeta() (*ledger.ContractMeta, error)
	GetContractMetaInSnapshot(contractAddress types.Address, snapshotBlock *ledger.SnapshotBlock) (meta *ledger.ContractMeta, err error)
	GetConfirmSnapshotHeader(blockHash types.Hash) (*ledger.SnapshotBlock, error)
	LatestSnapshotBlock() (*ledger.SnapshotBlock, error)
}

func AddBalance(db dbInterface, id *types.TokenTypeId, amount *big.Int) {
	b, err := db.GetBalance(id)
	DealWithErr(err)
	b.Add(b, amount)
	db.SetBalance(id, b)
}

func SubBalance(db dbInterface, id *types.TokenTypeId, amount *big.Int) {
	b, err := db.GetBalance(id)
	DealWithErr(err)
	if b.Cmp(amount) >= 0 {
		b.Sub(b, amount)
		db.SetBalance(id, b)
	}
}

func GetValue(db dbInterface, key []byte) []byte {
	v, err := db.GetValue(key)
	DealWithErr(err)
	return v
}

func SetValue(db dbInterface, key []byte, value []byte) {
	err := db.SetValue(key, value)
	DealWithErr(err)
}

// For normal send block:
// 1. toAddr is user, quota ratio is 1;
// 2. toAddr is contract, contract is created in latest snapshot block, return quota ratio
// 3. toAddr is contract, contract is not created in latest snapshot block, return error
func GetQuotaRatioForS(db dbInterface, toAddr types.Address) (uint8, error) {
	if !types.IsContractAddr(toAddr) {
		return CommonQuotaRatio, nil
	}
	sb, err := db.LatestSnapshotBlock()
	DealWithErr(err)
	return GetQuotaRatioBySnapshotBlock(db, toAddr, sb)
}

func CheckContractAddrInSAfterNewFork(db dbInterface, toAddr types.Address) error {
	sb, err := db.LatestSnapshotBlock()
	DealWithErr(err)
	if !fork.IsNewFork(sb.Height) {
		return nil
	}
	_, err = GetQuotaRatioForS(db, toAddr)
	return err
}

// For send block generated by contract receive block:
// 1. toAddr is user, quota ratio is 1;
// 2. toAddr is contract, send block is confirmed, contract is created in confirm status, return quota ratio
// 3. toAddr is contract, send block is confirmed, contract is not created in confirm status, return error
// 4. toAddr is contract, send block is not confirmed, contract is created in latest block, return quota ratio
// 5. toAddr is contract, send block is not confirmed, contract is not created in latest block, wait for a reliable status
func GetQuotaRatioForRS(db dbInterface, toAddr types.Address, fromHash types.Hash, status GlobalStatus) (uint8, error) {
	if !types.IsContractAddr(toAddr) {
		return CommonQuotaRatio, nil
	}
	if !helper.IsNil(status) && status.SnapshotBlock() != nil {
		return GetQuotaRatioBySnapshotBlock(db, toAddr, status.SnapshotBlock())
	}
	confirmSb, err := db.GetConfirmSnapshotHeader(fromHash)
	DealWithErr(err)
	if confirmSb != nil {
		return GetQuotaRatioBySnapshotBlock(db, toAddr, confirmSb)
	}
	sb, err := db.LatestSnapshotBlock()
	DealWithErr(err)
	meta, err := db.GetContractMetaInSnapshot(toAddr, sb)
	DealWithErr(err)
	if meta != nil && meta.IsDeleted() {
		return 0, ErrContractNotExists
	}
	if meta != nil {
		return meta.QuotaRatio, nil
	}
	return 0, ErrNoReliableStatus
}

func CheckContractAddrInRSAfterNewFork(db dbInterface, toAddr types.Address, fromHash types.Hash, status GlobalStatus) error {
	sb, err := db.LatestSnapshotBlock()
	DealWithErr(err)
	if !fork.IsNewFork(sb.Height) {
		return nil
	}
	_, err = GetQuotaRatioForRS(db, toAddr, fromHash, status)
	return err
}

func GetQuotaRatioBySnapshotBlock(db dbInterface, toAddr types.Address, snapshotBlock *ledger.SnapshotBlock) (uint8, error) {
	meta, err := db.GetContractMetaInSnapshot(toAddr, snapshotBlock)
	DealWithErr(err)
	if meta == nil || meta.IsDeleted() {
		return 0, ErrContractNotExists
	}
	return meta.QuotaRatio, nil
}
