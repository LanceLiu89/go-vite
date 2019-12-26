package abi

import (
	"github.com/vitelabs/go-vite/vm/abi"
	"strings"
)

const (
	jsonDeFi = `
	[
		{"type":"function","name":"Deposit","inputs":[]},
        {"type":"function","name":"Withdraw", "inputs":[{"name":"token","type":"tokenId"},{"name":"amount","type":"uint256"}]},
        {"type":"function","name":"NewLoan", "inputs":[{"name":"token","type":"tokenId"},{"name":"dayRate","type":"int32"},{"name":"shareAmount","type":"uint256"},{"name":"shares","type":"int32"},{"name":"subscribeDays","type":"int32"},{"name":"expireDays","type":"int32"}]},
        {"type":"function","name":"CancelLoan", "inputs":[{"name":"loanId","type":"uint64"}]},
        {"type":"function","name":"Subscribe", "inputs":[{"name":"loanId","type":"uint64"},{"name":"shares","type":"int32"}]},
        {"type":"function","name":"Invest", "inputs":[{"name":"loanId","type":"uint64"},{"name":"bizType","type":"uint8"},{"name":"amount","type":"uint256"},{"name":"beneficiary","type":"address"}]},
        {"type":"function","name":"RegisterSBP", "inputs":[{"name":"loanId","type":"uint64"},{"name":"sbpName","type":"string"},{"name":"blockProducingAddress","type":"address"},{"name":"rewardWithdrawAddress","type":"address"}]},
		{"type":"function","name":"UpdateSBPRegistration", "inputs":[{"name":"investId","type":"uint64"},{"name":"operationCode","type":"uint8"},{"name":"blockProducingAddress","type":"address"},{"name":"rewardWithdrawAddress","type":"address"}]},
        {"type":"function","name":"CancelInvest", "inputs":[{"name":"investId","type":"uint64"}]},
		{"type":"function","name":"RefundInvest", "inputs":[{"name":"InvestHashes","type":"bytes"},{"name":"reason","type":"uint8"}]},
		{"type":"function","name":"StakeForQuotaWithCallbackCallback", "inputs":[{"name":"id","type":"bytes32"},{"name":"success","type":"bool"}]},
		{"type":"function","name":"CancelQuotaStakingWithCallbackCallback", "inputs":[{"name":"id","type":"bytes32"},{"name":"success","type":"bool"}]},
		{"type":"function","name":"DeFiAdminConfig", "inputs":[{"name":"operationCode","type":"uint8"},{"name":"owner","type":"address"},{"name":"timeOracle","type":"address"},{"name":"jobTrigger","type":"address"},{"name":"operator","type":"address"}]},
		{"type":"function","name":"TriggerJob", "inputs":[{"name":"bizType","type":"uint8"},{"name":"data","type":"bytes"}]},
		{"type":"function","name":"NotifyTime", "inputs":[{"name":"timestamp","type":"int64"}]}

]`

	MethodNameDeFiDeposit                     = "Deposit"
	MethodNameDeFiWithdraw                    = "Withdraw"
	MethodNameDeFiNewLoan                     = "NewLoan"
	MethodNameDeFiCancelLoan                  = "CancelLoan"
	MethodNameDeFiSubscribe                   = "Subscribe"
	MethodNameDeFiInvest                      = "Invest"
	MethodNameDeFiRegisterSBP                 = "RegisterSBP"
	MethodNameDeFiUpdateSBPRegistration       = "UpdateSBPRegistration"
	MethodNameDeFiCancelInvest                = "CancelInvest"
	MethodNameDeFiRefundInvest                = "RefundInvest"
	MethodNameDeFiDelegateStakeCallback       = "StakeForQuotaWithCallbackCallback"
	MethodNameDeFiCancelDelegateStakeCallback = "CancelQuotaStakingWithCallbackCallback"
	MethodNameDeFiAdminConfig                 = "DeFiAdminConfig"
	MethodNameDeFiTriggerJob                  = "TriggerJob"
	MethodNameDeFiNotifyTime                  = "NotifyTime"
)

var (
	ABIDeFi, _ = abi.JSONToABIContract(strings.NewReader(jsonDeFi))
)
