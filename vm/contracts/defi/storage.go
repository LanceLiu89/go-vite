package defi

import (
	"bytes"
	"github.com/golang/protobuf/proto"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/vm/contracts/common"
	defiproto "github.com/vitelabs/go-vite/vm/contracts/defi/proto"
	"github.com/vitelabs/go-vite/vm/contracts/dex"
	"github.com/vitelabs/go-vite/vm/util"
	"github.com/vitelabs/go-vite/vm_db"
	"math/big"
)

var (
	fundKeyPrefix              = []byte("fd:") //fd:address
	loanKeyPrefix              = []byte("ln:") //ln:loanId 3+8
	loanSerialNoKey            = []byte("lnSn:")
	subscriptionKeyPrefix      = []byte("sb:")    //sub:loanId+address 3+8+20 = 31
	investKeyPrefix            = []byte("ivt:")   // ivt:Id 4+8 = 19
	investToLoanIndexKeyPrefix = []byte("iv2LI:") // iv2LI:loanId,investId
	investSerialNoKey          = []byte("ivtSn:") //invest:
	investQuotaInfoKeyPrefix   = []byte("ivQ:")
	sbpRegistrationKeyPrefix = []byte("ivSp:")

	ownerKey = []byte("own:")
	timeOracleKey = []byte("tmr:")
	defiTimestampKey = []byte("tmst:")
	jobTriggerKey = []byte("jbtg:")

	initOwner, _ = types.HexToAddress("vite_a8a00b3a2f60f5defb221c68f79b65f3620ee874f951a825db")
)

const (
	LoanOpen = iota + 1
	LoanSuccess
	LoanCancelled
	LoanFailed
	LoanExpired
	LoanExpiredRefunded
)

//invest type
const (
	InvestForMining = iota + 1
	InvestForSVIP
	InvestForSBP
	InvestForQuota
)

//invest status
const (
	InvestPending = iota + 1
	InvestSuccess
	InvestCancelling
	InvestRefunded
)

const (
	RefundFailedInvest = iota + 1
	RefundCancelledInvest
)

const (
	UpdateBlockProducingAddress    = 1
	UpdateSBPRewardWithdrawAddress = 2
)

const (
	JobUpdateLoan     = 1
	JobUpdateInvest   = 2
	JobSettleInterest = 3
)

//baseAccount update bizType
const (
	BaseDeposit = iota + 1
	BaseWithdraw
	BaseSubscribeLock
	BaseSubscribeSuccessReduce
	BaseSubscribeFailedRelease
	BaseSubscribeExpiredRefund
	BaseLoanInterestLocked
	BaseLoanInterestReduce
	BaseLoanInterestRelease
	BaseSubscribeInterestIncome
	BaseInvestReduce
	BaseInvestRefund
)

//loanAccount update bizType
const (
	LoanAccNewSuccessLoan = iota + 1
	LoanAccExpiredRefund
	LoanAccInvestReduce
	LoanAccInvestRefund
)

//MethodNameDeFiAdminConfig
const (
	AdminConfigOwner      = 1
	AdminConfigTimeOracle = 2
	AdminConfigJobTrigger = 4
	AdminConfigOperator   = 8
)

type ParamWithdraw struct {
	Token  types.TokenTypeId
	Amount *big.Int
}

type ParamNewLoan struct {
	Token         types.TokenTypeId
	DayRate       int32
	ShareAmount   *big.Int
	Shares        int32
	SubscribeDays int32
	ExpireDays    int32
}

type ParamSubscribe struct {
	LoanId uint64
	Shares int32
}

type ParamInvest struct {
	LoanId      uint64
	BizType     uint8
	Amount      *big.Int
	Beneficiary types.Address
}

type ParamRefundInvest struct {
	InvestHashes []byte
	Reason       uint8
}

type ParamRegisterSBP struct {
	LoanId                uint64
	SbpName               string
	BlockProducingAddress types.Address
	RewardWithdrawAddress types.Address
}

type ParamUpdateSBPRegistration struct {
	InvestId              uint64
	OperationCode         uint8
	BlockProducingAddress types.Address
	RewardWithdrawAddress types.Address
}

type ParamDeFiAdminConfig struct {
	OperationCode uint8
	Owner         types.Address // 1 owner
	TimeOracle    types.Address // 2 timeOracle
	JobTrigger    types.Address // 4 jobTrigger
	Operator      types.Address // 8 operator
}

type ParamTriggerJob struct {
	BizType uint8
	Data    []byte
}

type Fund struct {
	defiproto.Fund
}

func (fd *Fund) Serialize() (data []byte, err error) {
	return proto.Marshal(&fd.Fund)
}

func (fd *Fund) DeSerialize(data []byte) (err error) {
	protoFund := defiproto.Fund{}
	if err := proto.Unmarshal(data, &protoFund); err != nil {
		return err
	} else {
		fd.Fund = protoFund
		return nil
	}
}

type Loan struct {
	defiproto.Loan
}

func (ln *Loan) Serialize() (data []byte, err error) {
	return proto.Marshal(&ln.Loan)
}

func (ln *Loan) DeSerialize(data []byte) error {
	loan := defiproto.Loan{}
	if err := proto.Unmarshal(data, &loan); err != nil {
		return err
	} else {
		ln.Loan = loan
		return nil
	}
}

type Subscription struct {
	defiproto.Subscription
}

func (sb *Subscription) Serialize() (data []byte, err error) {
	return proto.Marshal(&sb.Subscription)
}

func (sb *Subscription) DeSerialize(data []byte) error {
	subscription := defiproto.Subscription{}
	if err := proto.Unmarshal(data, &subscription); err != nil {
		return err
	} else {
		sb.Subscription = subscription
		return nil
	}
}

type Invest struct {
	defiproto.Invest
}

func (iv *Invest) Serialize() (data []byte, err error) {
	return proto.Marshal(&iv.Invest)
}

func (iv *Invest) DeSerialize(data []byte) error {
	invest := defiproto.Invest{}
	if err := proto.Unmarshal(data, &invest); err != nil {
		return err
	} else {
		iv.Invest = invest
		return nil
	}
}

type InvestQuotaInfo struct {
	defiproto.InvestQuotaInfo
}

func (ivq *InvestQuotaInfo) Serialize() (data []byte, err error) {
	return proto.Marshal(&ivq.InvestQuotaInfo)
}

func (ivq *InvestQuotaInfo) DeSerialize(data []byte) error {
	info := defiproto.InvestQuotaInfo{}
	if err := proto.Unmarshal(data, &info); err != nil {
		return err
	} else {
		ivq.InvestQuotaInfo = info
		return nil
	}
}

type SBPRegistration struct {
	defiproto.SBPRegistration
}

func (sbpr *SBPRegistration) Serialize() (data []byte, err error) {
	return proto.Marshal(&sbpr.SBPRegistration)
}

func (sbpr *SBPRegistration) DeSerialize(data []byte) error {
	sbpRegistration := defiproto.SBPRegistration{}
	if err := proto.Unmarshal(data, &sbpRegistration); err != nil {
		return err
	} else {
		sbpr.SBPRegistration = sbpRegistration
		return nil
	}
}

type TimeWithHeight struct {
	defiproto.TimeWithHeight
}

func (twh *TimeWithHeight) Serialize() (data []byte, err error) {
	return proto.Marshal(&twh.TimeWithHeight)
}

func (twh *TimeWithHeight) DeSerialize(data []byte) error {
	timeWithHeight := defiproto.TimeWithHeight{}
	if err := proto.Unmarshal(data, &timeWithHeight); err != nil {
		return err
	} else {
		twh.TimeWithHeight = timeWithHeight
		return nil
	}
}

func SaveLoan(db vm_db.VmDb, loan *Loan) {
	common.SerializeToDb(db, getLoanKey(loan.Id), loan)
}

func DeleteLoan(db vm_db.VmDb, loan *Loan) {
	common.SetValueToDb(db, getLoanKey(loan.Id), nil)
}

func GetLoan(db vm_db.VmDb, loanId uint64) (loan *Loan, ok bool) {
	loan = &Loan{}
	ok = common.DeserializeFromDb(db, getLoanKey(loanId), loan)
	return
}

func GetLoanAvailable(gs util.GlobalStatus, loan *Loan) (available *big.Int, availableHeight uint64) {
	available = new(big.Int).Sub(CalculateAmount(loan.Shares, loan.ShareAmount), new(big.Int).SetBytes(loan.Invested))
	availableHeight = loan.ExpireHeight - gs.SnapshotBlock().Height
	return
}

func IsOpenLoanSubscribeFail(db vm_db.VmDb, loan *Loan, deFiDayHeight uint64) bool {
	return loan.Status == LoanOpen && GetDeFiTimestamp(db) > GetLoanFailTime(loan, deFiDayHeight)
}

func getLoanKey(loanId uint64) []byte {
	return append(loanKeyPrefix, common.Uint64ToBytes(loanId)...)
}

func SaveSubscription(db vm_db.VmDb, sub *Subscription) {
	common.SerializeToDb(db, getSubscriptionKey(sub.LoanId, sub.Address), sub)
}

func GetSubscription(db vm_db.VmDb, loanId uint64, address []byte) (sub *Subscription, ok bool) {
	sub = &Subscription{}
	ok = common.DeserializeFromDb(db, getSubscriptionKey(loanId, address), sub)
	return
}

func DeleteSubscription(db vm_db.VmDb, sub *Subscription) {
	common.SetValueToDb(db, getSubscriptionKey(sub.LoanId, sub.Address), nil)
}

func getSubscriptionKey(loanId uint64, address []byte) []byte {
	return append(subscriptionKeyPrefix, append(common.Uint64ToBytes(loanId), address...)...)
}

func SaveInvest(db vm_db.VmDb, invest *Invest) {
	common.SerializeToDb(db, getInvestKey(invest.Id), invest)
}

func GetInvest(db vm_db.VmDb, investId uint64) (invest *Invest, ok bool) {
	invest = &Invest{}
	ok = common.DeserializeFromDb(db, getInvestKey(investId), invest)
	return
}

func ConfirmInvest(db vm_db.VmDb, invest *Invest) {
	invest.Status = InvestSuccess
	invest.Updated = GetDeFiTimestamp(db)
	common.SerializeToDb(db, getInvestKey(invest.Id), invest)
	AddInvestUpdateEvent(db, invest)
}

func CancellingInvest(db vm_db.VmDb, invest *Invest) {
	invest.Status = InvestCancelling
	invest.Updated = GetDeFiTimestamp(db)
	common.SerializeToDb(db, getInvestKey(invest.Id), invest)
	AddInvestUpdateEvent(db, invest)
}

func DeleteInvest(db vm_db.VmDb, investId uint64) {
	common.SetValueToDb(db, getInvestKey(investId), nil)
}

func getInvestKey(investId uint64) []byte {
	return append(investKeyPrefix, common.Uint64ToBytes(investId)...)
}

func SaveInvestToLoanIndex(db vm_db.VmDb, invest *Invest) {
	common.SetValueToDb(db, getInvestToLoanIndexKey(invest.LoanId, invest.Id), []byte{1})
}

func DeleteInvestToLoanIndex(db vm_db.VmDb, invest *Invest) {
	common.SetValueToDb(db, getInvestToLoanIndexKey(invest.LoanId, invest.Id), nil)
}

func getInvestToLoanIndexKey(loanId, investId uint64) []byte {
	return append(append(investToLoanIndexKeyPrefix, common.Uint64ToBytes(loanId)...), common.Uint64ToBytes(investId)...)
}

func NewLoanSerialNo(db vm_db.VmDb) (serialNo uint64) {
	return newSerialNo(db, loanSerialNoKey)
}

func NewInvestSerialNo(db vm_db.VmDb) (serialNo uint64) {
	return newSerialNo(db, investSerialNoKey)
}

func GetInvestQuotaInfo(db vm_db.VmDb, hash []byte) (info *InvestQuotaInfo, ok bool) {
	info = &InvestQuotaInfo{}
	ok = common.DeserializeFromDb(db, GetInvestQuotaInfoKey(hash), info)
	return
}

func SaveInvestQuotaInfo(db vm_db.VmDb, hash types.Hash, invest *Invest, amount *big.Int) {
	info := &InvestQuotaInfo{}
	info.Address = invest.Address
	info.Amount = amount.Bytes()
	info.InvestId = invest.Id
	common.SerializeToDb(db, GetInvestQuotaInfoKey(hash.Bytes()), info)
}

func DeleteInvestQuotaInfo(db vm_db.VmDb, hash []byte) {
	common.SetValueToDb(db, GetInvestQuotaInfoKey(hash), nil)
}

func GetInvestQuotaInfoKey(hash []byte) []byte {
	return append(investQuotaInfoKeyPrefix, hash[len(investQuotaInfoKeyPrefix):]...)
}

func GetSBPRegistration(db vm_db.VmDb, hash []byte) (info *SBPRegistration, ok bool) {
	info = &SBPRegistration{}
	ok = common.DeserializeFromDb(db, GetSBPRegistrationKey(hash), info)
	return
}

func NewSBPRegistration(param *ParamRegisterSBP, invest *Invest) *SBPRegistration {
	info := &SBPRegistration{}
	info.InvestId = invest.Id
	info.Name = param.SbpName
	info.ProducingAddress = param.BlockProducingAddress.Bytes()
	info.RewardWithdrawAddress = param.RewardWithdrawAddress.Bytes()
	return info
}

func SaveSBPRegistration(db vm_db.VmDb, hash []byte, info *SBPRegistration) {
	common.SerializeToDb(db, GetSBPRegistrationKey(hash), info)
}

func DeleteSBPRegistration(db vm_db.VmDb, hash []byte) {
	common.SetValueToDb(db, GetSBPRegistrationKey(hash), nil)
}

func GetSBPRegistrationKey(hash []byte) []byte {
	return append(sbpRegistrationKeyPrefix, hash[len(sbpRegistrationKeyPrefix):]...)
}

func IsOwner(db vm_db.VmDb, address types.Address) bool {
	if storeOwner := common.GetValueFromDb(db, ownerKey); len(storeOwner) == types.AddressSize {
		return bytes.Equal(storeOwner, address.Bytes())
	} else {
		return address == initOwner
	}
}

func GetOwner(db vm_db.VmDb) *types.Address {
	return common.GetAddressFromKey(db, ownerKey)
}

func SetOwner(db vm_db.VmDb, address types.Address) {
	common.SetValueToDb(db, ownerKey, address.Bytes())
}

func ValidTimeOracle(db vm_db.VmDb, address types.Address) bool {
	return bytes.Equal(common.GetValueFromDb(db, timeOracleKey), address.Bytes())
}

func GetTimeOracle(db vm_db.VmDb) *types.Address {
	return common.GetAddressFromKey(db, timeOracleKey)
}

func SetTimeOracle(db vm_db.VmDb, address types.Address) {
	common.SetValueToDb(db, timeOracleKey, address.Bytes())
}

func ValidTriggerAddress(db vm_db.VmDb, address types.Address) bool {
	return bytes.Equal(common.GetValueFromDb(db, jobTriggerKey), address.Bytes())
}

func GetJobTrigger(db vm_db.VmDb) *types.Address {
	return common.GetAddressFromKey(db, jobTriggerKey)
}

func SetPeriodJobTrigger(db vm_db.VmDb, address types.Address) {
	common.SetValueToDb(db, jobTriggerKey, address.Bytes())
}

func SetDeFiTimestamp(db vm_db.VmDb, timestamp int64, gs util.GlobalStatus) error {
	twh, _ := GetDeFiTimeWithHeight(db)
	if timestamp > twh.Timestamp {
		twh.Timestamp = timestamp
		twh.Height = gs.SnapshotBlock().Height
		common.SerializeToDb(db, defiTimestampKey, twh)
		return nil
	} else {
		return dex.InvalidTimestampFromTimeOracleErr
	}
}

func GetDeFiTimeWithHeight(db vm_db.VmDb) (twh *TimeWithHeight, ok bool) {
	twh = &TimeWithHeight{}
	ok = common.DeserializeFromDb(db, defiTimestampKey, twh)
	return
}

func GetDeFiTimestamp(db vm_db.VmDb) int64 {
	twh, _ := GetDeFiTimeWithHeight(db)
	return twh.Timestamp
}

func GetDeFiEstimateTimestamp(db vm_db.VmDb, gs util.GlobalStatus) (int64, int64) {
	twh, _ := GetDeFiTimeWithHeight(db)
	return twh.Timestamp + int64(gs.SnapshotBlock().Height-twh.Height), twh.Timestamp
}

func GetExpireHeight(gs util.GlobalStatus, days int32, deFiDayHeight uint64) uint64 {
	return gs.SnapshotBlock().Height + uint64(days)*deFiDayHeight
}

func GetLoanFailTime(loan *Loan, deFiDayHeight uint64) int64 {
	return loan.Created + int64(loan.SubscribeDays)*int64(deFiDayHeight)
}

func newAccount(token []byte) *defiproto.Account {
	account := &defiproto.Account{}
	account.Token = token
	account.BaseAccount = &defiproto.BaseAccount{}
	account.LoanAccount = &defiproto.LoanAccount{}
	return account
}

func newSerialNo(db vm_db.VmDb, key []byte) (serialNo uint64) {
	if data := common.GetValueFromDb(db, key); len(data) == 8 {
		serialNo = common.BytesToUint64(data)
		serialNo++
	} else {
		serialNo = 1
	}
	common.SetValueToDb(db, key, common.Uint64ToBytes(serialNo))
	return
}
