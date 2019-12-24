package defi

import (
	"github.com/vitelabs/go-vite/interfaces"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/vm/contracts/common"
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
			if bls, err = innerUpdateLoan(db, loan, estTime, time, gs, deFiDayHeight); err != nil {
				return
			} else {
				blocks = append(blocks, bls...)
			}
		}
	}
	return
}

func UpdateInvests(db vm_db.VmDb, data []byte, confirmSeconds int64) {
	if len(data) > 0 {
		for i := 0; i < len(data)/8; i++ {
			investId := common.BytesToUint64(data[i*8 : (i+1)*8])
			if invest, ok := GetInvest(db, investId); ok {
				innerUpdateInvest(db, invest, GetDeFiTimestamp(db), confirmSeconds)
			}
		}
	} else {
		iterator, err := db.NewStorageIterator(investKeyPrefix)
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
			innerUpdateInvest(db, invest, GetDeFiTimestamp(db), confirmSeconds)
		}
	}
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

func innerUpdateInvest(db vm_db.VmDb, invest *Invest, time int64, confirmSeconds int64) {
	if invest.Status == InvestPending && invest.BizType != InvestForQuota && time-invest.Created > confirmSeconds {
		//InvestForMining, InvestForSVIP, InvestForSBP
		ConfirmInvest(db, invest)
	}
}
