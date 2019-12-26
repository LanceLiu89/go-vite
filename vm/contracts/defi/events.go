package defi

import (
	"github.com/golang/protobuf/proto"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/vm/contracts/common"
	defiproto "github.com/vitelabs/go-vite/vm/contracts/defi/proto"
	"github.com/vitelabs/go-vite/vm_db"
)

const newLoanEventName = "newLoanEvent"
const loanUpdateEventName = "loanUpdateEvent"
const newSubscriptionEventName = "newSubscriptionEvent"
const subscriptionUpdateEventName = "subscriptionUpdateEvent"
const newInvestEventName = "newInvestEvent"
const investUpdateEventName = "investUpdateEvent"
const sbpNewRegistrationEventName = "sbpNewRegistrationEvent"
const sbpRegistrationUpdateEventName = "sbpRegistrationUpdate"
const baseAccountEventName = "baseAccountEvent"
const loanAccountEventName = "loanAccountEvent"

func AddNewLoanEvent(db vm_db.VmDb, ln *Loan) {
	event := &NewLoanEvent{}
	event.Loan = ln.Loan
	common.DoEmitEventLog(db, event)
}

func AddLoanUpdateEvent(db vm_db.VmDb, ln *Loan) {
	event := &LoanUpdateEvent{}
	event.Id = ln.Id
	event.SubscribedShares = ln.SubscribedShares
	event.Invested = ln.Invested
	event.Status = ln.Status
	event.ExpireHeight = ln.ExpireHeight
	event.StartHeight = ln.StartHeight
	event.SettledDays = ln.SettledDays
	event.SettledInterest = ln.SettledInterest
	event.Timestamp = ln.Updated
	common.DoEmitEventLog(db, event)
}

func AddNewSubscriptionEvent(db vm_db.VmDb, sub *Subscription) {
	event := &NewSubscriptionEvent{}
	event.Subscription = sub.Subscription
	common.DoEmitEventLog(db, event)
}

func AddSubscriptionUpdateEvent(db vm_db.VmDb, sub *Subscription) {
	event := &SubscriptionUpdateEvent{}
	event.LoanId = sub.LoanId
	event.Address = sub.Address
	event.Shares = sub.Shares
	event.Status = sub.Status
	event.Interest = sub.Interest
	common.DoEmitEventLog(db, event)
}

func AddNewInvestEvent(db vm_db.VmDb, invest *Invest) {
	event := &NewInvestEvent{}
	event.Invest = invest.Invest
	common.DoEmitEventLog(db, event)
}

func AddInvestUpdateEvent(db vm_db.VmDb, invest *Invest) {
	event := &InvestUpdateEvent{}
	event.Id = invest.Id
	event.Status = invest.Status
	common.DoEmitEventLog(db, event)
}

func AddSBPNewRegistrationEvent(db vm_db.VmDb, newREg *SBPRegistration) {
	event := &SBPNewRegistrationEvent{}
	event.SBPRegistration = newREg.SBPRegistration
	common.DoEmitEventLog(db, event)
}

func AddSBPRegistrationUpdateEvent(db vm_db.VmDb, regUpdate *SBPRegistration) {
	event := &SBPNewRegistrationEvent{}
	event.SBPRegistration = regUpdate.SBPRegistration
	common.DoEmitEventLog(db, event)
}

func AddBaseAccountEvent(db vm_db.VmDb, address []byte, bizType int32, investType uint8, loanId uint64, amount []byte) {
	event := &BaseAccountEvent{}
	event.Address = address
	event.BizType = bizType
	event.InvestType = int32(investType)
	event.LoanId = loanId
	event.Amount = amount
	common.DoEmitEventLog(db, event)
}

func AddLoanAccountEvent(db vm_db.VmDb, address []byte, bizType int32, investType uint8, loanId uint64, amount []byte) {
	event := &LoanAccountEvent{}
	event.Address = address
	event.BizType = bizType
	event.InvestType = int32(investType)
	event.LoanId = loanId
	event.Amount = amount
	common.DoEmitEventLog(db, event)
}

type NewLoanEvent struct {
	defiproto.Loan
}

func (ln NewLoanEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(newLoanEventName)
}

func (ln NewLoanEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&ln.Loan)
	return data
}

func (ln NewLoanEvent) FromBytes(data []byte) interface{} {
	event := NewLoanEvent{}
	if err := proto.Unmarshal(data, &event.Loan); err != nil {
		return nil
	} else {
		return event
	}
}

type LoanUpdateEvent struct {
	defiproto.LoanUpdate
}

func (lu LoanUpdateEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(loanUpdateEventName)
}

func (lu LoanUpdateEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&lu.LoanUpdate)
	return data
}

func (lu LoanUpdateEvent) FromBytes(data []byte) interface{} {
	event := LoanUpdateEvent{}
	if err := proto.Unmarshal(data, &event.LoanUpdate); err != nil {
		return nil
	} else {
		return event
	}
}

type NewSubscriptionEvent struct {
	defiproto.Subscription
}

func (sb NewSubscriptionEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(newSubscriptionEventName)
}

func (sb NewSubscriptionEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&sb.Subscription)
	return data
}

func (sb NewSubscriptionEvent) FromBytes(data []byte) interface{} {
	event := NewSubscriptionEvent{}
	if err := proto.Unmarshal(data, &event.Subscription); err != nil {
		return nil
	} else {
		return event
	}
}

type SubscriptionUpdateEvent struct {
	defiproto.SubscriptionUpdate
}

func (su SubscriptionUpdateEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(subscriptionUpdateEventName)
}

func (su SubscriptionUpdateEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&su.SubscriptionUpdate)
	return data
}

func (su SubscriptionUpdateEvent) FromBytes(data []byte) interface{} {
	event := SubscriptionUpdateEvent{}
	if err := proto.Unmarshal(data, &event.SubscriptionUpdate); err != nil {
		return nil
	} else {
		return event
	}
}

type NewInvestEvent struct {
	defiproto.Invest
}

func (nie NewInvestEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(newInvestEventName)
}

func (nie NewInvestEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&nie.Invest)
	return data
}

func (nie NewInvestEvent) FromBytes(data []byte) interface{} {
	event := NewInvestEvent{}
	if err := proto.Unmarshal(data, &event.Invest); err != nil {
		return nil
	} else {
		return event
	}
}

type InvestUpdateEvent struct {
	defiproto.InvestUpdate
}

func (iv InvestUpdateEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(investUpdateEventName)
}

func (iv InvestUpdateEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&iv.InvestUpdate)
	return data
}

func (iv InvestUpdateEvent) FromBytes(data []byte) interface{} {
	event := InvestUpdateEvent{}
	if err := proto.Unmarshal(data, &event.InvestUpdate); err != nil {
		return nil
	} else {
		return event
	}
}

type SBPNewRegistrationEvent struct {
	defiproto.SBPRegistration
}

func (snr SBPNewRegistrationEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(sbpNewRegistrationEventName)
}

func (snr SBPNewRegistrationEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&snr)
	return data
}

func (snr SBPNewRegistrationEvent) FromBytes(data []byte) interface{} {
	event := SBPNewRegistrationEvent{}
	if err := proto.Unmarshal(data, &event.SBPRegistration); err != nil {
		return nil
	} else {
		return event
	}
}

type SBPRegistrationUpdateEvent struct {
	defiproto.SBPRegistration
}

func (sru SBPRegistrationUpdateEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(sbpRegistrationUpdateEventName)
}

func (sru SBPRegistrationUpdateEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&sru.SBPRegistration)
	return data
}

func (sru SBPRegistrationUpdateEvent) FromBytes(data []byte) interface{} {
	event := SBPRegistrationUpdateEvent{}
	if err := proto.Unmarshal(data, &event.SBPRegistration); err != nil {
		return nil
	} else {
		return event
	}
}

type BaseAccountEvent struct {
	defiproto.BaseAccountUpdate
}

func (ba BaseAccountEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(baseAccountEventName)
}

func (ba BaseAccountEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&ba.BaseAccountUpdate)
	return data
}

func (ba BaseAccountEvent) FromBytes(data []byte) interface{} {
	event := BaseAccountEvent{}
	if err := proto.Unmarshal(data, &event.BaseAccountUpdate); err != nil {
		return nil
	} else {
		return event
	}
}

type LoanAccountEvent struct {
	defiproto.LoanAccountUpdate
}

func (la LoanAccountEvent) GetTopicId() types.Hash {
	return common.FromNameToHash(loanAccountEventName)
}

func (la LoanAccountEvent) ToDataBytes() []byte {
	data, _ := proto.Marshal(&la.LoanAccountUpdate)
	return data
}

func (la LoanAccountEvent) FromBytes(data []byte) interface{} {
	event := LoanAccountEvent{}
	if err := proto.Unmarshal(data, &event.LoanAccountUpdate); err != nil {
		return nil
	} else {
		return event
	}
}
