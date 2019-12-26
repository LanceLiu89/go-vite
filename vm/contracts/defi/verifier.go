package defi

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/vm/contracts/common"
	"github.com/vitelabs/go-vite/vm_db"
	"math/big"
)

type DeFiVerifyRes struct {
	UserCount      int                                   `json:"userCount"`
	BalanceMatched bool                                  `json:"balanceMatched"`
	VerifyItems    map[types.TokenTypeId]*DeFiVerifyItem `json:"balances"`
}

type DeFiVerifyItem struct {
	TokenId        types.TokenTypeId `json:"tokenId"`
	Balance        string            `json:"balance"`
	Amount         string            `json:"amount"`
	Invested       string            `json:"invested"`
	BalanceMatched bool              `json:"balanceMatched"`
}

type DeFiAccumulateAccount struct {
	TokenId  types.TokenTypeId
	Amount   []byte
	Invested []byte
}

func VerifyDeFiBalance(db vm_db.VmDb) *DeFiVerifyRes {
	userAccountMap := make(map[types.TokenTypeId]*DeFiAccumulateAccount)
	verifyItems := make(map[types.TokenTypeId]*DeFiVerifyItem)
	count, _ := accumulateUserAccount(db, userAccountMap)
	allBalanceMatched := true
	var (
		balanceMatched bool
		balance                     *big.Int
	)
	for tokenId, acc := range userAccountMap {
		amount := new(big.Int).SetBytes(acc.Amount)
		invested := new(big.Int).SetBytes(acc.Invested)
		balance, _ = db.GetBalance(&tokenId)
		if balanceMatched = amount.Cmp(balance) == 0; !balanceMatched {
			allBalanceMatched = false
		}
		verifyItems[tokenId] = &DeFiVerifyItem{TokenId: tokenId, Balance: balance.String(), Amount: amount.String(), Invested:invested.String(), BalanceMatched: balanceMatched}
	}
	return &DeFiVerifyRes{
		count,
		allBalanceMatched,
		verifyItems,
	}
}

func accumulateUserAccount(db vm_db.VmDb, accumulateRes map[types.TokenTypeId]*DeFiAccumulateAccount) (int, error) {
	var (
		userAccountValue []byte
		userFund         *Fund
		ok               bool
	)
	var count = 0
	iterator, err := db.NewStorageIterator(fundKeyPrefix)
	if err != nil {
		return 0, err
	}
	defer iterator.Release()
	for {
		if ok = iterator.Next(); ok {
			userAccountValue = iterator.Value()
			if len(userAccountValue) == 0 {
				continue
			}
		} else {
			break
		}
		userFund = &Fund{}
		if err = userFund.DeSerialize(userAccountValue); err != nil {
			return 0, err
		}
		for _, acc := range userFund.Accounts {
			tokenId, _ := types.BytesToTokenTypeId(acc.Token)
			total := common.AddBigInt(acc.BaseAccount.Available, acc.BaseAccount.Subscribing)
			total = common.AddBigInt(total, acc.BaseAccount.Locked)
			total = common.AddBigInt(total, acc.LoanAccount.Available)
			invested := common.AddBigInt(acc.BaseAccount.Invested, acc.LoanAccount.Invested)
			accAccount(tokenId, total, invested, accumulateRes)
		}
		count++
	}
	return count, nil
}

func accAccount(tokenId types.TokenTypeId, amount, invested []byte, accAccount map[types.TokenTypeId]*DeFiAccumulateAccount) {
	if acc, ok := accAccount[tokenId]; ok {
		acc.Amount = common.AddBigInt(acc.Amount, amount)
		acc.Invested = common.AddBigInt(acc.Invested, invested)
	} else {
		accAccount[tokenId] = &DeFiAccumulateAccount{
			tokenId, amount, invested,
		}
	}
}
