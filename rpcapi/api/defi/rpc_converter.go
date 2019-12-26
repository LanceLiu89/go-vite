package defi

import (
	"encoding/hex"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/rpcapi/api/dex"
	"github.com/vitelabs/go-vite/vm/contracts/defi"
)

type RpcLoan struct {
	Id               uint64 `json:"id"`
	Address          string `json:"address"`
	Token            string `json:"token"`
	ShareAmount      string `json:"shareAmount"`
	Shares           int32  `json:"shares"`
	Amount           string `json:"amount"`
	Interest         string `json:"interest,omitempty"`
	DayRate          int32  `json:"dayRate"`
	SubscribeDays    int32  `json:"subscribeDays"`
	ExpireDays       int32  `json:"expireDays"`
	SubscribedShares int32  `json:"subscribedShares,omitempty"`
	StartHeight      uint64 `json:"startHeight,omitempty"`
	ExpireHeight     uint64 `json:"expireHeight,omitempty"`
	Invested         string `json:"invested,omitempty"`
	Status           int32  `json:"status,omitempty"`
	SettledInterest  string `json:"settledInterest,omitempty"`
	SettledDays      int32  `json:"settledDays,omitempty"`
	Created          int64  `json:"created"`
	StartTime        int64  `json:"startTime,omitempty"`
	Updated          int64  `json:"updated,omitempty"`
}

type RpcLoanPage struct {
	Loans      []*RpcLoan `json:"loans"`
	LastLoanId uint64     `json:"lastLoanId"`
	Count      int        `json:"count"`
}

func LoanToRpc(loan *defi.Loan) *RpcLoan {
	var rl *RpcLoan = nil
	if loan != nil {
		address, _ := types.BytesToAddress(loan.Address)
		token, _ := types.BytesToTokenTypeId(loan.Token)
		rl = &RpcLoan{
			Id:               loan.Id,
			Address:          address.String(),
			Token:            token.String(),
			ShareAmount:      dex.AmountBytesToString(loan.ShareAmount),
			Shares:           loan.Shares,
			Amount:           defi.CalculateAmount(loan.Shares, loan.ShareAmount).String(),
			Interest:         dex.AmountBytesToString(loan.Interest),
			DayRate:          loan.DayRate,
			SubscribeDays:    loan.SubscribeDays,
			ExpireDays:       loan.ExpireDays,
			SubscribedShares: loan.SubscribedShares,
			StartHeight:      loan.StartHeight,
			ExpireHeight:     loan.ExpireHeight,
			Invested:         dex.AmountBytesToString(loan.Invested),
			Status:           loan.Status,
			SettledInterest:  dex.AmountBytesToString(loan.SettledInterest),
			SettledDays:      loan.SettledDays,
			Created:          loan.Created,
			StartTime:        loan.StartTime,
			Updated:          loan.Updated,
		}
	}
	return rl
}

type RpcSubscription struct {
	LoanId      uint64 `json:"loanId"`
	Address     string `json:"address"`
	Token       string `json:"token"`
	ShareAmount string `json:"shareAmount"`
	Shares      int32  `json:"shares"`
	Amount      string `json:"amount"`
	Interest    string `json:"interest,omitempty"`
	Status      int32  `json:"status,omitempty"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated,omitempty"`
}

type RpcSubscriptionPage struct {
	Subscriptions []*RpcSubscription `json:"subscriptions"`
	LastSubKey    string             `json:"lastSubKey"`
	Count         int                `json:"count"`
}

func SubscriptionToRpc(sub *defi.Subscription) *RpcSubscription {
	var rs *RpcSubscription = nil
	if sub != nil {
		address, _ := types.BytesToAddress(sub.Address)
		token, _ := types.BytesToTokenTypeId(sub.Token)
		rs = &RpcSubscription{
			LoanId:      sub.LoanId,
			Address:     address.String(),
			Token:       token.String(),
			ShareAmount: dex.AmountBytesToString(sub.ShareAmount),
			Shares:      sub.Shares,
			Amount:      defi.CalculateAmount(sub.Shares, sub.ShareAmount).String(),
			Interest:    dex.AmountBytesToString(sub.Interest),
			Status:      sub.Status,
			Created:     sub.Created,
			Updated:     sub.Updated,
		}
	}
	return rs
}

type RpcInvest struct {
	Id           uint64 `json:"id"`
	LoanId       uint64 `json:"loanId"`
	Address      string `json:"address"`
	LoanAmount   string `json:"loanAmount"`
	BaseAmount   string `json:"baseAmount"`
	BizType      int32  `json:"bizType,omitempty"`
	Beneficial   string `json:"beneficial,omitempty"`
	CreateHeight uint64 `json:"createHeight"`
	ExpireHeight uint64 `json:"expireHeight,omitempty"`
	Status       int32  `json:"status,omitempty"`
	InvestHash   string `json:"investHash"`
	Created      int64  `json:"created"`
	Updated      int64  `json:"updated,omitempty"`
}

type RpcInvestPage struct {
	Invests      []*RpcInvest `json:"invests"`
	LastInvestId uint64       `json:"lastInvestId"`
	Count        int          `json:"count"`
}

func InvestToRpc(invest *defi.Invest) *RpcInvest {
	var ri *RpcInvest = nil
	if invest != nil {
		address, _ := types.BytesToAddress(invest.Address)
		ri = &RpcInvest{
			Id:           invest.Id,
			LoanId:       invest.LoanId,
			Address:      address.String(),
			LoanAmount:   dex.AmountBytesToString(invest.LoanAmount),
			BaseAmount:   dex.AmountBytesToString(invest.BaseAmount),
			BizType:      invest.BizType,
			CreateHeight: invest.CreateHeight,
			ExpireHeight: invest.ExpireHeight,
			Status:       invest.Status,
			InvestHash:   hex.EncodeToString(invest.InvestHash),
			Created:      invest.Created,
			Updated:      invest.Updated,
		}
		if len(invest.Beneficial) > 0 {
			beneficial, _ := types.BytesToAddress(invest.Beneficial)
			ri.Beneficial = beneficial.String()
		}
	}
	return ri
}

type RpcSBPRegistration struct {
	InvestId              uint64 `json:"investId"`
	Name                  string `json:"name"`
	ProducingAddress      string `json:"producingAddress"`
	RewardWithdrawAddress string `json:"rewardWithdrawAddress"`
}

func SBPRegistrationToRpc(sbp *defi.SBPRegistration) *RpcSBPRegistration {
	var sbpR *RpcSBPRegistration = nil
	if sbp != nil {
		produceAddress, _ := types.BytesToAddress(sbp.ProducingAddress)
		rewardAddress, _ := types.BytesToAddress(sbp.RewardWithdrawAddress)
		sbpR = &RpcSBPRegistration{
			InvestId:              sbp.InvestId,
			Name:                  sbp.Name,
			ProducingAddress:      produceAddress.String(),
			RewardWithdrawAddress: rewardAddress.String(),
		}
	}
	return sbpR
}

type RpcInvestQuotaInfo struct {
	Address  string `json:"address"`
	Amount   string `json:"amount"`
	InvestId uint64 `json:"investId"`
}

func InvestQuotaInfoToRpc(iq *defi.InvestQuotaInfo) *RpcInvestQuotaInfo {
	var res *RpcInvestQuotaInfo = nil
	if iq != nil {
		address, _ := types.BytesToAddress(iq.Address)
		res = &RpcInvestQuotaInfo{
			Address:  address.String(),
			Amount:   dex.AmountBytesToString(iq.Amount),
			InvestId: iq.InvestId,
		}
	}
	return res
}
