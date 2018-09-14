package vm_context

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/contracts"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/trie"
)

type Chain interface {
	GetAccount(address *types.Address) (*ledger.Account, error)
	GetSnapshotBlockByHash(hash *types.Hash) (*ledger.SnapshotBlock, error)
	GetAccountBlockByHash(blockHash *types.Hash) (*ledger.AccountBlock, error)
	GetStateTrie(hash *types.Hash) *trie.Trie
	GetTokenInfoById(tokenId *types.TokenTypeId) (*contracts.TokenInfo, error)

	NewStateTrie() *trie.Trie
	GetConfirmAccountBlock(snapshotHeight uint64, address *types.Address) (*ledger.AccountBlock, error)
}
