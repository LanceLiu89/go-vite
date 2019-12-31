package defi

import (
	"github.com/vitelabs/go-vite/interfaces"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/vm/contracts/common"
	"github.com/vitelabs/go-vite/vm/contracts/dex"
	"github.com/vitelabs/go-vite/vm/util"
	"github.com/vitelabs/go-vite/vm_db"
	"math/big"
)

func UpdateLoans(db vm_db.VmDb, data []byte, gs util.GlobalStatus, deFiDayHeight uint64) (blocks []*ledger.AccountBlock, err error) {
	estTime, time := GetDeFiEstimateTimestamp(db, gs)
	var bls []*ledger.AccountBlock
	if len(data) > 0 {
		for i := 0; i < len(data)/8; i++ {
			loanId := common.BytesToUint64(data[i*8 : (i+1)*8])
			if loan, ok := GetLoan(db, loanId); ok {
				if bls, err = innerUpdateLoan(db, loan, estTime, time, gs, deFiDayHeight); err != nil {
					return
				} else {
					blocks = append(blocks, bls...)
				}
			}
		}
	} else {
		err = traverseLoan(db, func(loan *Loan) error {
			if bls, err = innerUpdateLoan(db, loan, estTime, time, gs, deFiDayHeight); err != nil {
				return err
			} else {
				blocks = append(blocks, bls...)
			}
			return nil
		})
	}
	return
}

func UpdateInvests(db vm_db.VmDb, data []byte, gs util.GlobalStatus, confirmSeconds int64, deFiDayHeight uint64) (blocks []*ledger.AccountBlock, err error) {
	var bls []*ledger.AccountBlock
	if len(data) > 0 {
		for i := 0; i < len(data)/8; i++ {
			investId := common.BytesToUint64(data[i*8 : (i+1)*8])
			if invest, ok := GetInvest(db, investId); ok {
				if bls, err = innerUpdateInvest(db, invest, gs, confirmSeconds, deFiDayHeight); err != nil {
					return
				} else {
					blocks = append(blocks, bls...)
				}
			}
		}
	} else {
		var iterator interfaces.StorageIterator
		iterator, err = db.NewStorageIterator(investKeyPrefix)
		if err != nil {
			panic(err)
		}
		defer iterator.Release()
		for {
			if !iterator.Next() {
				if iterator.Error() != nil {
					panic(iterator.Error())
				}
				break
			}
			investValue := iterator.Value()
			invest := &Invest{}
			if err = invest.DeSerialize(investValue); err != nil {
				panic(err)
			}
			if bls, err = innerUpdateInvest(db, invest, gs, confirmSeconds, deFiDayHeight); err != nil {
				return
			} else {
				blocks = append(blocks, bls...)
			}
		}
	}
	return
}

func SettleInterest(db vm_db.VmDb, data []byte, gs util.GlobalStatus, dayHeight uint64) (err error) {
	if len(data) > 0 {
		for i := 0; i < len(data)/8; i++ {
			loanId := common.BytesToUint64(data[i*8 : (i+1)*8])
			if loan, ok := GetLoan(db, loanId); ok {
				if err = innerSettleLoanInterest(db, loan, gs, dayHeight); err != nil {
					return
				}
			}
		}
	} else {
		return traverseLoan(db, func(loan *Loan) error {
			return innerSettleLoanInterest(db, loan, gs, dayHeight)
		})
	}
	return
}

func innerUpdateLoan(db vm_db.VmDb, loan *Loan, estTime, time int64, gs util.GlobalStatus, deFiDayHeight uint64) (blocks []*ledger.AccountBlock, err error) {
	switch loan.Status {
	case LoanOpen:
		if loan.Created+int64(loan.SubscribeDays)*int64(deFiDayHeight) < estTime {
			loan.Status = LoanFailed
			loan.Updated = time
			DoRefundLoan(db, loan)
		}
	case LoanSuccess:
		if gs.SnapshotBlock().Height > loan.ExpireHeight {
			loan.Updated = time
			innerSettleLoanInterest(db, loan, gs, deFiDayHeight)
			if new(big.Int).SetBytes(loan.Invested).Sign() == 0 {
				loan.Status = LoanExpiredRefunded
				DoRefundLoan(db, loan)
			} else {
				loan.Status = LoanExpired
				expireLoanSubscriptions(db, loan)
				SaveLoan(db, loan)
				AddLoanUpdateEvent(db, loan)
				blocks, err = DoCancelExpiredLoanInvests(db, loan)
			}
		}
	case LoanExpired:
		if new(big.Int).SetBytes(loan.Invested).Sign() == 0 {
			loan.Updated = time
			loan.Status = LoanExpiredRefunded
			DoRefundLoan(db, loan)
		} else {
			if time-loan.Updated > 300 { //retry cancel invests
				loan.Updated = time
				SaveLoan(db, loan)
				blocks, err = DoCancelExpiredLoanInvests(db, loan)
			}
		}
	}
	return
}

func innerUpdateInvest(db vm_db.VmDb, invest *Invest, gs util.GlobalStatus, confirmSeconds int64, deFiDayHeight uint64) ([]*ledger.AccountBlock, error){
	if invest.Status == InvestPending && invest.BizType != InvestForQuota && GetDeFiTimestamp(db)-invest.Created > confirmSeconds {
		//InvestForMining, InvestForSVIP, InvestForSBP
		ConfirmInvest(db, invest)
	} else if invest.Status == InvestSuccess && invest.BizType == InvestForMining {
		if invest.ExpireHeight <= gs.SnapshotBlock().Height {//can be cancelled
			loan, _ := GetLoan(db, invest.LoanId)
			if (loan.ExpireHeight - gs.SnapshotBlock().Height)/deFiDayHeight <= uint64(dex.SchedulePeriods) {
				CancellingInvest(db, invest)
				return DoCancelDexInvest(common.Uint64ToBytes(invest.Id))
			}
		}
	}
	return nil, nil
}

func innerSettleLoanInterest(db vm_db.VmDb, loan *Loan, gs util.GlobalStatus, deFiDayHeight uint64) (err error) {
	if loan.Status == LoanSuccess && loan.SettledDays < loan.ExpireDays {
		toSettleDays := int32((gs.SnapshotBlock().Height - loan.StartHeight) / deFiDayHeight)
		if loan.SettledDays >= toSettleDays {
			return
		}
		interest := new(big.Int).SetBytes(loan.Interest)
		settledInterest := new(big.Int).SetBytes(loan.SettledInterest)
		if settledInterest.Cmp(interest) >= 0 {
			loan.SettledDays = loan.ExpireDays
		} else {
			leavedInterest := new(big.Int).Sub(interest, settledInterest)
			loanNewInterest := new(big.Int)
			for i := loan.SettledDays; i < toSettleDays && i < loan.ExpireDays && leavedInterest.Sign() > 0; i++ {
				if err = traverseLoanSubscriptions(db, loan, func(sub *Subscription) (err1 error) {
					subNewInterest := CalculateInterest(sub.Shares, new(big.Int).SetBytes(sub.ShareAmount), loan.DayRate, 1)
					if leavedInterest.Cmp(subNewInterest) <= 0 {
						subNewInterest = new(big.Int).Set(leavedInterest)
					}
					leavedInterest.Sub(leavedInterest, subNewInterest)
					loanNewInterest.Add(loanNewInterest, subNewInterest)
					sub.Interest = common.AddBigInt(sub.Interest, subNewInterest.Bytes())
					SaveSubscription(db, sub)
					AddSubscriptionUpdateEvent(db, sub)
					if _, err1 = OnAccSubscribeSettleInterest(db, sub.Address, subNewInterest); err1 != nil {
						return
					}
					AddBaseAccountEvent(db, sub.Address, BaseSubscribeInterestIncome, 0, loan.Id, subNewInterest.Bytes())
					return
				}); err != nil {
					return
				}
			}
			loan.SettledDays = toSettleDays
			loan.SettledInterest = common.AddBigInt(loan.SettledInterest, loanNewInterest.Bytes())
			if _, err = OnAccLoanSettleInterest(db, loan.Address, loanNewInterest.Bytes()); err != nil {
				return
			}
			AddBaseAccountEvent(db, loan.Address, BaseLoanInterestReduce, 0, loan.Id, loanNewInterest.Bytes())
		}
		loan.Updated = GetDeFiTimestamp(db)
		SaveLoan(db, loan)
		AddLoanUpdateEvent(db, loan)
	}
	return
}

func traverseLoan(db vm_db.VmDb, loanFunc func(loan *Loan) error) (err error) {
	var iterator interfaces.StorageIterator
	iterator, err = db.NewStorageIterator(loanKeyPrefix)
	if err != nil {
		panic(err)
	}
	defer iterator.Release()
	for {
		if !iterator.Next() {
			if iterator.Error() != nil {
				panic(iterator.Error())
			}
			break
		}
		loanValue := iterator.Value()
		loan := &Loan{}

		if err = loan.DeSerialize(loanValue); err != nil {
			panic(err)
		}
		if err = loanFunc(loan); err != nil {
			return
		}
	}
	return
}
