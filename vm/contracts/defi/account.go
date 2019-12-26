package defi

import (
	"bytes"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/vm/contracts/common"
	defiproto "github.com/vitelabs/go-vite/vm/contracts/defi/proto"
	"github.com/vitelabs/go-vite/vm_db"
	"math/big"
)

func GetAccounts(fund *Fund, token *types.TokenTypeId) ([]*defiproto.Account, bool) {
	if fund == nil || len(fund.Accounts) == 0 {
		return nil, false
	}
	if token != nil {
		acc, ok := GetAccountInfo(fund, *token)
		return []*defiproto.Account{acc}, ok
	} else {
		return fund.Accounts, true
	}
}

func GetAccountInfo(fund *Fund, token types.TokenTypeId) (*defiproto.Account, bool) {
	for _, a := range fund.Accounts {
		if bytes.Equal(token.Bytes(), a.Token) {
			return a, true
		}
	}
	return newAccount(token.Bytes()), false
}

func OnAccDeposit(db vm_db.VmDb, address types.Address, token types.TokenTypeId, amount *big.Int) (updatedAcc *defiproto.Account) {
	fund, _ := GetFund(db, address)
	var foundToken bool
	for _, acc := range fund.Accounts {
		if bytes.Equal(acc.Token, token.Bytes()) {
			acc.BaseAccount.Available = common.AddBigInt(acc.BaseAccount.Available, amount.Bytes())
			updatedAcc = acc
			foundToken = true
			break
		}
	}
	if !foundToken {
		updatedAcc = newAccount(token.Bytes())
		updatedAcc.BaseAccount.Available = amount.Bytes()
		fund.Accounts = append(fund.Accounts, updatedAcc)
	}
	SaveFund(db, address, fund)
	return
}

func OnAccWithdraw(db vm_db.VmDb, address types.Address, tokenId []byte, amount *big.Int) (*defiproto.Account, error) {
	return updateFund(db, address, tokenId, func(acc *defiproto.Account) (*defiproto.Account, error) {
		available := new(big.Int).SetBytes(acc.BaseAccount.Available)
		if available.Cmp(amount) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.BaseAccount.Available = available.Sub(available, amount).Bytes()
		}
		return acc, nil
	})
}

func OnAccNewLoan(db vm_db.VmDb, address types.Address, interest *big.Int) (*defiproto.Account, error) {
	return updateFund(db, address, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		available := new(big.Int).SetBytes(acc.BaseAccount.Available)
		if available.Cmp(interest) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.BaseAccount.Available = available.Sub(available, interest).Bytes()
			acc.BaseAccount.Locked = common.AddBigInt(acc.BaseAccount.Locked, interest.Bytes())
		}
		return acc, nil
	})
}

func OnAccLoanFailed(db vm_db.VmDb, address types.Address, interest []byte) (*defiproto.Account, error) {
	return updateFund(db, address, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		if common.CmpForBigInt(acc.BaseAccount.Locked, interest) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.BaseAccount.Available = common.AddBigInt(acc.BaseAccount.Available, interest)
			acc.BaseAccount.Locked = common.SubBigIntAbs(acc.BaseAccount.Locked, interest)
		}
		return acc, nil
	})
}

func OnAccLoanSuccess(db vm_db.VmDb, address []byte, loan *Loan) (*defiproto.Account, error) {
	addr, _ := types.BytesToAddress(address)
	return updateFund(db, addr, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		acc.LoanAccount.Available = common.AddBigInt(acc.LoanAccount.Available, CalculateAmount(loan.Shares, loan.ShareAmount).Bytes())
		return acc, nil
	})
}

func OnAccLoanSettleInterest(db vm_db.VmDb, address []byte, interest []byte) (*defiproto.Account, error) {
	addr, _ := types.BytesToAddress(address)
	return updateFund(db, addr, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		if common.CmpForBigInt(acc.BaseAccount.Locked, interest) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.BaseAccount.Locked = common.SubBigInt(acc.BaseAccount.Locked, interest).Bytes()
		}
		return acc, nil
	})
}

func OnAccLoanExpired(db vm_db.VmDb, address types.Address, amount *big.Int) (*defiproto.Account, error) {
	return updateFund(db, address, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		available := new(big.Int).SetBytes(acc.LoanAccount.Available)
		if available.Cmp(amount) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.LoanAccount.Available = available.Sub(available, amount).Bytes()
		}
		return acc, nil
	})
}

func OnAccSubscribe(db vm_db.VmDb, address types.Address, amount *big.Int) (*defiproto.Account, error) {
	return updateFund(db, address, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		available := new(big.Int).SetBytes(acc.BaseAccount.Available)
		if available.Cmp(amount) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.BaseAccount.Available = available.Sub(available, amount).Bytes()
			acc.BaseAccount.Subscribing = common.AddBigInt(acc.BaseAccount.Subscribing, amount.Bytes())
		}
		return acc, nil
	})
}

func OnAccSubscribeSuccess(db vm_db.VmDb, address []byte, amount *big.Int) (*defiproto.Account, error) {
	addr, _ := types.BytesToAddress(address)
	return updateFund(db, addr, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		subscribing := new(big.Int).SetBytes(acc.BaseAccount.Subscribing)
		if subscribing.Cmp(amount) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.BaseAccount.Subscribing = common.SubBigIntAbs(acc.BaseAccount.Subscribing, amount.Bytes())
			acc.BaseAccount.Subscribed = common.AddBigInt(acc.BaseAccount.Subscribed, amount.Bytes())
		}
		return acc, nil
	})
}

func OnAccSubscribeSettleInterest(db vm_db.VmDb, address []byte, interest *big.Int) (*defiproto.Account, error) {
	addr, _ := types.BytesToAddress(address)
	return updateFund(db, addr, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		acc.BaseAccount.Available = common.AddBigInt(acc.BaseAccount.Available, interest.Bytes())
		return acc, nil
	})
}

func OnAccRefundFailedSubscription(db vm_db.VmDb, address []byte, amount *big.Int) (*defiproto.Account, error) {
	addr, _ := types.BytesToAddress(address)
	return updateFund(db, addr, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		subscribing := new(big.Int).SetBytes(acc.BaseAccount.Subscribing)
		if subscribing.Cmp(amount) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.BaseAccount.Available = common.AddBigInt(acc.BaseAccount.Available, amount.Bytes())
			acc.BaseAccount.Subscribing = common.SubBigIntAbs(acc.BaseAccount.Subscribing, amount.Bytes())
		}
		return acc, nil
	})
}

func OnAccRefundSuccessSubscription(db vm_db.VmDb, address []byte, amount *big.Int) (*defiproto.Account, error) {
	addr, _ := types.BytesToAddress(address)
	return updateFund(db, addr, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		subscribed := new(big.Int).SetBytes(acc.BaseAccount.Subscribed)
		if subscribed.Cmp(amount) < 0 {
			return nil, ExceedFundAvailableErr
		} else {
			acc.BaseAccount.Available = common.AddBigInt(acc.BaseAccount.Available, amount.Bytes())
			acc.BaseAccount.Subscribed = common.SubBigInt(acc.BaseAccount.Subscribed, amount.Bytes()).Bytes()
		}
		return acc, nil
	})
}

func OnAccInvest(db vm_db.VmDb, address types.Address, loanAmount, baseAmount *big.Int) (*defiproto.Account, error) {
	return updateFund(db, address, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		if loanAmount.Sign() != 0 {
			loanAvailable := new(big.Int).SetBytes(acc.LoanAccount.Available)
			if loanAvailable.Cmp(loanAmount) < 0 {
				return nil, ExceedFundAvailableErr
			} else {
				acc.LoanAccount.Available = loanAvailable.Sub(loanAvailable, loanAmount).Bytes()
				acc.LoanAccount.Invested = common.AddBigInt(acc.LoanAccount.Invested, loanAmount.Bytes())
			}
		}
		if baseAmount.Sign() != 0 {
			baseAvailable := new(big.Int).SetBytes(acc.BaseAccount.Available)
			if baseAvailable.Cmp(baseAmount) < 0 {
				return nil, ExceedFundAvailableErr
			} else {
				acc.BaseAccount.Available = baseAvailable.Sub(baseAvailable, baseAmount).Bytes()
				acc.BaseAccount.Invested = common.AddBigInt(acc.BaseAccount.Invested, baseAmount.Bytes())
			}
		}
		return acc, nil
	})
}

func OnAccRefundInvest(db vm_db.VmDb, address []byte, loanAmount, baseAmount []byte) (*defiproto.Account, error) {
	addr, _ := types.BytesToAddress(address)
	return updateFund(db, addr, ledger.ViteTokenId.Bytes(), func(acc *defiproto.Account) (*defiproto.Account, error) {
		if len(loanAmount) != 0 {
			if common.CmpForBigInt(acc.LoanAccount.Invested, loanAmount) < 0 {
				return nil, ExceedFundAvailableErr
			} else {
				acc.LoanAccount.Invested = common.SubBigIntAbs(acc.LoanAccount.Invested, loanAmount)
				acc.LoanAccount.Available = common.AddBigInt(acc.LoanAccount.Available, loanAmount)
			}
		}
		if len(baseAmount) > 0 {
			if common.CmpForBigInt(acc.BaseAccount.Invested, baseAmount) < 0 {
				return nil, ExceedFundAvailableErr
			} else {
				acc.BaseAccount.Invested = common.SubBigInt(acc.BaseAccount.Invested, baseAmount).Bytes()
				acc.BaseAccount.Available = common.AddBigInt(acc.BaseAccount.Available, baseAmount)
			}
		}
		return acc, nil
	})
}

func updateFund(db vm_db.VmDb, address types.Address, tokenId []byte, updateAccFunc func(*defiproto.Account) (*defiproto.Account, error)) (updatedAcc *defiproto.Account, err error) {
	if fund, ok := GetFund(db, address); ok {
		var foundAcc bool
		for _, acc := range fund.Accounts {
			if bytes.Equal(acc.Token, tokenId) {
				foundAcc = true
				if updatedAcc, err = updateAccFunc(acc); err != nil {
					return
				}
				break
			}
		}
		if foundAcc {
			SaveFund(db, address, fund)
		} else {
			err = ExceedFundAvailableErr
		}
	} else {
		err = ExceedFundAvailableErr
	}
	return
}

func GetFund(db vm_db.VmDb, address types.Address) (fund *Fund, ok bool) {
	fund = &Fund{}
	ok = common.DeserializeFromDb(db, GetFundKey(address), fund)
	return
}

func SaveFund(db vm_db.VmDb, address types.Address, fund *Fund) {
	common.SerializeToDb(db, GetFundKey(address), fund)
}

func GetFundKey(address types.Address) []byte {
	return append(fundKeyPrefix, address.Bytes()...)
}
