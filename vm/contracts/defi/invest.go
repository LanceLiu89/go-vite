package defi

import (
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/interfaces"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/vm/contracts/abi"
	"github.com/vitelabs/go-vite/vm/contracts/common"
	defiproto "github.com/vitelabs/go-vite/vm/contracts/defi/proto"
	"github.com/vitelabs/go-vite/vm/contracts/dex"
	"github.com/vitelabs/go-vite/vm/util"
	"github.com/vitelabs/go-vite/vm_db"
	"math/big"
)

var (
	LoanRateCardinalNum int32 = 1e6
	MinDayRate          int32 = 1 // 1/1,000,000
	MaxDayRate                = LoanRateCardinalNum

	MinSubDays int32 = 1
	MaxSubDays int32 = 7

	commonTokenPow = big.NewInt(1e18)
	minShareAmount = new(big.Int).Mul(big.NewInt(10), commonTokenPow)
)

func PrepareInvest(db vm_db.VmDb, address types.Address, bizType uint8, loanAvailable, stakeAmount *big.Int, availableHeight, stakeHeightMin, stakeSVIPHeight, deFiDayHeight uint64) (loanInvested, baseInvested *big.Int, durationHeight uint64, err error) {
	var (
		baseAvailable = new(big.Int)
		fund          *Fund
		acc           *defiproto.Account
		ok            bool
	)
	if fund, ok = GetFund(db, address); ok {
		if acc, ok = GetAccountInfo(fund, ledger.ViteTokenId); ok {
			baseAvailable.SetBytes(acc.BaseAccount.Available)
		}
	}
	totalAvailable := new(big.Int).Add(loanAvailable, baseAvailable)

	//fmt.Printf("loanAvailable %s, baseAvailable %s\n", loanAvailable.String(), baseAvailable.String())
	switch bizType {
	case InvestForMining:
		if totalAvailable.Cmp(dex.StakeForMiningMinAmount) < 0 {
			err = ExceedFundAvailableErr
			return
		} else if availableHeight < stakeHeightMin+uint64(dex.SchedulePeriods)*deFiDayHeight {
			err = AvailableHeightNotValidForInvestErr
			return
		}
		durationHeight = stakeHeightMin
		loanInvested, baseInvested = getInvestedAmount(loanAvailable, stakeAmount)
	case InvestForSVIP:
		if totalAvailable.Cmp(dex.StakeForSuperVIPAmount) < 0 {
			err = ExceedFundAvailableErr
			return
		} else if availableHeight < stakeSVIPHeight {
			err = AvailableHeightNotValidForInvestErr
			return
		}
		durationHeight = stakeSVIPHeight
		loanInvested, baseInvested = getInvestedAmount(loanAvailable, dex.StakeForSuperVIPAmount)
	case InvestForQuota:
		if totalAvailable.Cmp(dex.StakeForMiningMinAmount) < 0 {
			err = ExceedFundAvailableErr
			return
		} else if availableHeight < stakeHeightMin {
			err = AvailableHeightNotValidForInvestErr
			return
		}
		durationHeight = stakeHeightMin
		loanInvested, baseInvested = getInvestedAmount(loanAvailable, stakeAmount)
	case InvestForSBP:
		if totalAvailable.Cmp(stakeAmount) < 0 {
			err = ExceedFundAvailableErr
			return
		} else if availableHeight < stakeHeightMin {
			err = AvailableHeightNotValidForInvestErr
			return
		}
		durationHeight = stakeHeightMin
		loanInvested, baseInvested = getInvestedAmount(loanAvailable, stakeAmount)
	}
	return
}

func NewInvest(db vm_db.VmDb, gs util.GlobalStatus, address types.Address, loan *Loan, bizType uint8, beneficiary types.Address, loanInvested, baseInvested *big.Int, durationHeight uint64) *Invest {
	invest := &Invest{}
	invest.Id = NewInvestSerialNo(db)
	invest.LoanId = loan.Id
	invest.Address = address.Bytes()
	invest.LoanAmount = loanInvested.Bytes()
	invest.BaseAmount = baseInvested.Bytes()
	invest.BizType = int32(bizType)
	invest.Beneficial = beneficiary.Bytes()
	invest.CreateHeight = gs.SnapshotBlock().Height
	invest.ExpireHeight = invest.CreateHeight + durationHeight
	invest.Status = InvestPending
	invest.Created = GetDeFiTimestamp(db)
	return invest
}

func DoDexInvest(invest *Invest, bizType uint8, amount *big.Int) ([]*ledger.AccountBlock, error) {
	if data, err := abi.ABIDexFund.PackMethod(abi.MethodNameDexFundDelegateInvest, invest.Id, invest.Address, bizType, invest.Beneficial); err != nil {
		return nil, err
	} else {
		return []*ledger.AccountBlock{
			{
				AccountAddress: types.AddressDeFi,
				ToAddress:      types.AddressDexFund,
				BlockType:      ledger.BlockTypeSendCall,
				Amount:         amount,
				TokenId:        ledger.ViteTokenId,
				Data:           data,
			},
		}, nil
	}
}

func DoCancelDexInvest(investId []byte) (blocks []*ledger.AccountBlock, err error) {
	if data, err := abi.ABIDexFund.PackMethod(abi.MethodNameDexFundCancelDelegateInvest, investId); err != nil {
		return nil, err
	} else {
		return []*ledger.AccountBlock{
			{
				AccountAddress: types.AddressDeFi,
				ToAddress:      types.AddressDexFund,
				BlockType:      ledger.BlockTypeSendCall,
				Amount:         big.NewInt(0),
				TokenId:        ledger.ViteTokenId,
				Data:           data,
			},
		}, nil
	}
}

func DoQuotaInvest(db vm_db.VmDb, beneficiary types.Address, invest *Invest, amount *big.Int, stakeHeight uint64, block *ledger.AccountBlock) (blocks []*ledger.AccountBlock, err error) {
	var (
		data    []byte
		stakeId types.Hash
	)
	if data, err = abi.ABIQuota.PackMethod(abi.MethodNameStakeWithCallback, beneficiary, stakeHeight); err != nil {
		return nil, err
	}
	blocks = []*ledger.AccountBlock{
		{
			AccountAddress: types.AddressDeFi,
			ToAddress:      types.AddressQuota,
			BlockType:      ledger.BlockTypeSendCall,
			Amount:         amount,
			TokenId:        ledger.ViteTokenId,
			Data:           data,
		},
	}
	stakeId = util.ComputeSendBlockHash(block, blocks[0], 0)
	invest.InvestHash = stakeId.Bytes()
	SaveInvest(db, invest)
	SaveInvestQuotaInfo(db, stakeId, invest, amount)
	return
}

func DoCancelQuotaInvest(investId []byte) (blocks []*ledger.AccountBlock, err error) {
	stakeId, _ := types.BytesToHash(investId)
	if data, err := abi.ABIQuota.PackMethod(abi.MethodNameCancelStakeWithCallback, stakeId); err != nil {
		return nil, err
	} else {
		return []*ledger.AccountBlock{
			{
				AccountAddress: types.AddressDeFi,
				ToAddress:      types.AddressQuota,
				BlockType:      ledger.BlockTypeSendCall,
				Amount:         big.NewInt(0),
				TokenId:        ledger.ViteTokenId,
				Data:           data,
			},
		}, nil
	}
}

func DoRegisterSBP(db vm_db.VmDb, invest *Invest, param *ParamRegisterSBP, amount *big.Int, block *ledger.AccountBlock) (blocks []*ledger.AccountBlock, err error) {
	var (
		data           []byte
		registrationId types.Hash
	)
	if data, err = abi.ABIGovernance.PackMethod(abi.MethodNameRegisterV3, param.SbpName, param.BlockProducingAddress, param.RewardWithdrawAddress); err != nil {
		return nil, err
	} else {
		blocks = []*ledger.AccountBlock{
			{
				AccountAddress: types.AddressDeFi,
				ToAddress:      types.AddressGovernance,
				BlockType:      ledger.BlockTypeSendCall,
				Amount:         amount,
				TokenId:        ledger.ViteTokenId,
				Data:           data,
			},
		}
		registrationId = util.ComputeSendBlockHash(block, blocks[0], 0)
		invest.InvestHash = registrationId.Bytes()
		SaveInvest(db, invest)
		regInfo := NewSBPRegistration(param, invest)
		SaveSBPRegistration(db, registrationId.Bytes(), regInfo)
		AddSBPNewRegistrationEvent(db, regInfo)
		return
	}
}

func DoRevokeSBP(db vm_db.VmDb, traceHash []byte) (blocks []*ledger.AccountBlock, err error) {
	if info, ok := GetSBPRegistration(db, traceHash); !ok {
		panic(SBPRegistrationNotExistsErr)
	} else {
		if data, err := abi.ABIGovernance.PackMethod(abi.MethodNameRevokeV3, info.Name); err != nil {
			return nil, err
		} else {
			return []*ledger.AccountBlock{
				{
					AccountAddress: types.AddressDeFi,
					ToAddress:      types.AddressGovernance,
					BlockType:      ledger.BlockTypeSendCall,
					Amount:         big.NewInt(0),
					TokenId:        ledger.ViteTokenId,
					Data:           data,
				},
			}, nil
		}
	}
}

func DoUpdateSBP(db vm_db.VmDb, traceHash []byte, param *ParamUpdateSBPRegistration) (blocks []*ledger.AccountBlock, err error) {
	if info, ok := GetSBPRegistration(db, traceHash); !ok {
		panic(SBPRegistrationNotExistsErr)
	} else {
		if common.IsOperationValidWithMask(param.OperationCode, UpdateBlockProducingAddress) {
			info.ProducingAddress = param.BlockProducingAddress.Bytes()
			if data, err := abi.ABIGovernance.PackMethod(abi.MethodNameUpdateBlockProducintAddressV3, info.Name, param.BlockProducingAddress); err != nil {
				return nil, err
			} else {
				blocks = []*ledger.AccountBlock{
					{
						AccountAddress: types.AddressDeFi,
						ToAddress:      types.AddressGovernance,
						BlockType:      ledger.BlockTypeSendCall,
						Amount:         big.NewInt(0),
						TokenId:        ledger.ViteTokenId,
						Data:           data,
					},
				}
			}
		}
		if common.IsOperationValidWithMask(param.OperationCode, UpdateSBPRewardWithdrawAddress) {
			info.RewardWithdrawAddress = param.RewardWithdrawAddress.Bytes()
			if data, err := abi.ABIGovernance.PackMethod(abi.MethodNameUpdateSBPRewardWithdrawAddress, info.Name, param.RewardWithdrawAddress); err != nil {
				return nil, err
			} else {
				blocks = append(blocks, &ledger.AccountBlock{
					AccountAddress: types.AddressDeFi,
					ToAddress:      types.AddressGovernance,
					BlockType:      ledger.BlockTypeSendCall,
					Amount:         big.NewInt(0),
					TokenId:        ledger.ViteTokenId,
					Data:           data,
				})
			}
		}
		SaveSBPRegistration(db, traceHash, info)
		AddSBPRegistrationUpdateEvent(db, info)
		return
	}
}

func DoRefundInvest(db vm_db.VmDb, invest *Invest) {
	OnAccRefundInvest(db, invest.Address, invest.LoanAmount, invest.BaseAmount)
	if loan, ok := GetLoan(db, invest.LoanId); !ok {
		panic(LoanNotExistsErr)
	} else {
		OnLoanCancelInvest(db, loan, invest.LoanAmount)
		AddLoanUpdateEvent(db, loan)
	}
	invest.Status = InvestRefunded
	DeleteInvest(db, invest.Id)
	DeleteInvestToLoanIndex(db, invest)
	AddInvestUpdateEvent(db, invest)
	AddLoanAccountEvent(db, invest.Address, LoanAccInvestRefund, uint8(invest.BizType), invest.Id, invest.LoanAmount)
	if len(invest.BaseAmount) > 0 {
		AddBaseAccountEvent(db, invest.Address, BaseInvestRefund, uint8(invest.BizType), invest.Id, invest.BaseAmount)
	}
}

func DoRefundQuotaInvest(db vm_db.VmDb, invest *Invest) {
	DoRefundInvest(db, invest)
	DeleteInvestQuotaInfo(db, invest.InvestHash)
}

func DoRefundSBPInvest(db vm_db.VmDb, invest *Invest) {
	DoRefundInvest(db, invest)
	DeleteSBPRegistration(db, invest.InvestHash)
}

func HandleDexRefundOnFail(db vm_db.VmDb, block *ledger.AccountBlock) (blocks []*ledger.AccountBlock, err error) {
	param := new(dex.ParamDelegateInvest)
	if err = abi.ABIDexFund.UnpackMethod(param, abi.MethodNameDexFundDelegateInvest, block.Data); err != nil {
		return
	}
	if invest, ok := GetInvest(db, param.InvestId); ok && invest.Status == InvestPending {
		DoRefundInvest(db, invest)
	} else {
		err = InvestNotExistsErr
	}
	return
}

func HandleGovernanceFeedback(db vm_db.VmDb, block *ledger.AccountBlock) (blocks []*ledger.AccountBlock, err error) {
	var (
		isRevoke bool
		param    = new(abi.ParamRegister)
		sbpName  = new(string)
	)
	if err = abi.ABIGovernance.UnpackMethod(param, abi.MethodNameRegisterV3, block.Data); err == nil && block.Amount.Sign() > 0 { //governance.RegisterSBP failed
		*sbpName = param.SbpName
	} else if err = abi.ABIGovernance.UnpackMethod(sbpName, abi.MethodNameRevokeV3, block.Data); err == nil { //governance.RevokeSBP success
		isRevoke = true
	} else {
		err = InvalidInputParamErr
		return
	}
	var iterator interfaces.StorageIterator
	iterator, err = db.NewStorageIterator(sbpRegistrationKeyPrefix)
	if err != nil {
		panic(err)
	}
	defer iterator.Release()
	for {
		var regValue []byte
		if !iterator.Next() {
			if iterator.Error() != nil {
				panic(iterator.Error())
			}
			break
		}
		regValue = iterator.Value()
		sbpRegistration := &SBPRegistration{}
		sbpRegistration.DeSerialize(regValue)
		if sbpRegistration.Name == *sbpName {
			if invest, ok := GetInvest(db, sbpRegistration.InvestId); !ok {
				err = InvestNotExistsErr
				return
			} else {
				if isRevoke && invest.Status != InvestCancelling || !isRevoke && invest.Status != InvestPending {
					err = InvalidInvestStatusErr
				} else {
					DoRefundSBPInvest(db, invest)
				}
			}
			return
		}
	}
	err = SBPRegistrationNotExistsErr
	return
}

func GetLoanInvests(db vm_db.VmDb, loanId uint64) (invests []*Invest, err error) {
	err = traverseLoanInvests(db, loanId, func(investId uint64) error {
		if invest, ok := GetInvest(db, investId); !ok {
			return InvestNotExistsErr
		} else {
			invests = append(invests, invest)
		}
		return nil
	})
	return
}

func GetInvestList(db vm_db.VmDb, investId uint64, count int) (infos []*Invest, newLastInvestId uint64, err error) {
	var iterator interfaces.StorageIterator
	if iterator, err = db.NewStorageIterator(investKeyPrefix); err != nil {
		return
	}
	defer iterator.Release()

	if investId > 0 {
		ok := iterator.Seek(getInvestKey(investId))
		if !ok {
			err = fmt.Errorf("last investId not valid for page invest list")
			return
		}
	}
	infos = make([]*Invest, 0, count)
	for {
		if !iterator.Next() {
			if err = iterator.Error(); err != nil {
				return
			}
			break
		}
		data := iterator.Value()
		if len(data) > 0 {
			invest := &Invest{}
			if err = invest.DeSerialize(data); err != nil {
				return
			} else {
				infos = append(infos, invest)
				newLastInvestId = common.BytesToUint64(iterator.Key()[len(investKeyPrefix):])
				if len(infos) == count {
					return
				}
			}
		}
	}
	return
}

func traverseLoanInvests(db vm_db.VmDb, loanId uint64, handler func(investId uint64) error) error {
	iterator, err := db.NewStorageIterator(append(investToLoanIndexKeyPrefix, common.Uint64ToBytes(loanId)...))
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
		investId := common.BytesToUint64(iterator.Key()[len(investToLoanIndexKeyPrefix)+8:])
		if err = handler(investId); err != nil {
			return err
		}
	}
	return nil
}

func IsInvestExpired(gs util.GlobalStatus, invest *Invest) bool {
	return gs.SnapshotBlock().Height >= invest.ExpireHeight
}

func getInvestedAmount(leavedLoanAmount *big.Int, needInvestAmount *big.Int) (loanInvested, baseInvested *big.Int) {
	if leavedLoanAmount.Cmp(needInvestAmount) < 0 {
		loanInvested = new(big.Int).Set(leavedLoanAmount)
		baseInvested = new(big.Int).Sub(needInvestAmount, loanInvested)
	} else {
		loanInvested = new(big.Int).Set(needInvestAmount)
		baseInvested = new(big.Int)
	}
	return
}

func CalculateAmount(shares int32, shareAmount []byte) *big.Int {
	return CalculateAmount1(shares, new(big.Int).SetBytes(shareAmount))
}

func CalculateAmount1(shares int32, shareAmount *big.Int) *big.Int {
	return new(big.Int).Mul(big.NewInt(int64(shares)), shareAmount)
}
