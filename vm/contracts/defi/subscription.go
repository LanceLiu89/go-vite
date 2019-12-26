package defi

import (
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/interfaces"
	"github.com/vitelabs/go-vite/vm/contracts/common"
	"github.com/vitelabs/go-vite/vm/util"
	"github.com/vitelabs/go-vite/vm_db"
)


func NewSubscription(address types.Address, db vm_db.VmDb, param *ParamSubscribe, loan *Loan) *Subscription {
	sub := &Subscription{}
	sub.LoanId = param.LoanId
	sub.Address = address.Bytes()
	sub.Token = loan.Token
	sub.Shares = param.Shares
	sub.ShareAmount = loan.ShareAmount
	sub.Status = LoanOpen
	sub.Created = GetDeFiTimestamp(db)
	return sub
}

func DoSubscribe(db vm_db.VmDb, gs util.GlobalStatus, loan *Loan, shares int32, deFiDayHeight uint64) (err error) {
	loan.SubscribedShares = loan.SubscribedShares + shares
	loan.Updated = GetDeFiTimestamp(db)
	if loan.Shares == loan.SubscribedShares {
		loan.Status = LoanSuccess
		loan.ExpireHeight = GetExpireHeight(gs, loan.ExpireDays, deFiDayHeight)
		loan.StartHeight = gs.SnapshotBlock().Height
		loan.StartTime = loan.Updated
		OnAccLoanSuccess(db, loan.Address, loan)
		AddLoanAccountEvent(db, loan.Address, LoanAccNewSuccessLoan, 0, loan.Id, CalculateAmount(loan.Shares, loan.ShareAmount).Bytes())
	}
	SaveLoan(db, loan)
	AddLoanUpdateEvent(db, loan)
	if loan.Status == LoanSuccess {
		err = traverseLoanSubscriptions(db, loan, func(sub *Subscription) (err1 error) {
			sub.Status = LoanSuccess
			sub.Updated = loan.Updated
			amount := CalculateAmount(sub.Shares, sub.ShareAmount)
			SaveSubscription(db, sub)
			AddSubscriptionUpdateEvent(db, sub)
			if _, err1 = OnAccSubscribeSuccess(db, sub.Address, amount); err1 != nil {
				return err1
			}
			AddBaseAccountEvent(db, sub.Address, BaseSubscribeSuccessReduce, 0, loan.Id, amount.Bytes())
			return
		})
	}
	return
}

func GetLoanSubscriptions(db vm_db.VmDb, loanId uint64) (subs []*Subscription, err error) {
	loan := &Loan{}
	loan.Id = loanId
	err = traverseLoanSubscriptions(db, loan, func(sub *Subscription) error {
		subs = append(subs, sub)
		return nil
	})
	return
}


func GetSubscriptionList(db vm_db.VmDb, subKey []byte, count int) (infos []*Subscription, newSubKey []byte, err error) {
	var iterator interfaces.StorageIterator
	if iterator, err = db.NewStorageIterator(subscriptionKeyPrefix); err != nil {
		return
	}
	defer iterator.Release()

	if len(subKey) > 0 {
		ok := iterator.Seek(subKey)
		if !ok {
			err = fmt.Errorf("last subscription key not valid for page subscription list")
			return
		}
	}
	infos = make([]*Subscription, 0, count)
	for {
		if !iterator.Next() {
			if err = iterator.Error(); err != nil {
				return
			}
			break
		}
		data := iterator.Value()
		if len(data) > 0 {
			sub := &Subscription{}
			if err = sub.DeSerialize(data); err != nil {
				return
			} else {
				infos = append(infos, sub)
				if len(infos) == count {
					newSubKey = iterator.Key()
					return
				}
			}
		}
	}
	return infos, iterator.Key(), nil
}

func traverseLoanSubscriptions(db vm_db.VmDb, loan *Loan, traverseFunc func(sub *Subscription) error ) (err error) {
	iterator, err := db.NewStorageIterator(append(subscriptionKeyPrefix, common.Uint64ToBytes(loan.Id)...))
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
		data := iterator.Value()
		sub := &Subscription{}
		if err = sub.DeSerialize(data); err != nil {
			panic(err)
		}
		if err = traverseFunc(sub); err != nil {
			return
		}
	}
	return
}