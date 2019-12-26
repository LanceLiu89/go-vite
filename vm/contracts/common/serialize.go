package common

import (
	gc "github.com/vitelabs/go-vite/common"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/vm_db"
)

type Serializable interface {
	Serialize() ([]byte, error)
	DeSerialize([]byte) error
}

func DeserializeFromDb(db vm_db.VmDb, key []byte, serializable Serializable) bool {
	if data := GetValueFromDb(db, key); len(data) > 0 {
		if err := serializable.DeSerialize(data); err != nil {
			panic(err)
		}
		return true
	} else {
		return false
	}
}

func SerializeToDb(db vm_db.VmDb, key []byte, serializable Serializable) {
	if data, err := serializable.Serialize(); err != nil {
		panic(err)
	} else {
		SetValueToDb(db, key, data)
	}
}

func GetValueFromDb(db vm_db.VmDb, key []byte) []byte {
	if data, err := db.GetValue(key); err != nil {
		panic(err)
	} else {
		return data
	}
}

func SetValueToDb(db vm_db.VmDb, key, value []byte) {
	if err := db.SetValue(key, value); err != nil {
		panic(err)
	}
}

func GetAddressFromKey(db vm_db.VmDb, key []byte) *types.Address {
	if addressBytes := GetValueFromDb(db, key); len(addressBytes) == types.AddressSize {
		address, _ := types.BytesToAddress(addressBytes)
		return &address
	} else {
		return nil
	}
}

type Event interface {
	GetTopicId() types.Hash
	ToDataBytes() []byte
	FromBytes([]byte) interface{}
}

func DoEmitEventLog(db vm_db.VmDb, event Event) {
	log := &ledger.VmLog{}
	log.Topics = append(log.Topics, event.GetTopicId())
	log.Data = event.ToDataBytes()
	db.AddLog(log)
}

func FromNameToHash(name string) types.Hash {
	hs := types.Hash{}
	hs.SetBytes(gc.RightPadBytes([]byte(name), types.HashSize))
	return hs
}
