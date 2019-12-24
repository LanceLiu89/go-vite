package common

import (
	"encoding/binary"
)

func Uint64ToBytes(value uint64) []byte {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, value)
	return bs
}

func BytesToUint64(bytes []byte) uint64 {
	return binary.BigEndian.Uint64(bytes)
}

func Uint32ToBytes(value uint32) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, value)
	return bs
}

func BytesToUint32(bytes []byte) uint32 {
	return binary.BigEndian.Uint32(bytes)
}

func BitwiseNotBytes(bytes []byte) {
	for i, b := range bytes {
		bytes[i] = ^b
	}
}

func IsOperationValidWithMask(operationCode, mask uint8) bool {
	return uint8(byte(operationCode)&byte(mask)) == mask
}