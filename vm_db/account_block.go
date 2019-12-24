package vm_db

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
)

func (vdb *vmDb) GetUnconfirmedBlocks(address types.Address) []*ledger.AccountBlock {
	return vdb.chain.GetUnconfirmedBlocks(address)
}
func (vdb *vmDb) GetLatestAccountBlock(addr types.Address) (*ledger.AccountBlock, error) {
	return vdb.chain.GetLatestAccountBlock(addr)
}

func (vdb *vmDb) GetCompleteBlockByHash(blockHash types.Hash) (*ledger.AccountBlock, error) {
	return vdb.chain.GetCompleteBlockByHash(blockHash)
}

func (vdb *vmDb) GetAccountBlockByHash(blockHash types.Hash) (*ledger.AccountBlock, error) {
	return vdb.chain.GetAccountBlockByHash(blockHash)
}
