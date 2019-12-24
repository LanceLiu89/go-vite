package common

import (
	"fmt"
	"github.com/vitelabs/go-vite/common/helper"
	"math/big"
	"strconv"
)

const BigFloatPrec = 120

func AddBigInt(a []byte, b []byte) []byte {
	return new(big.Int).Add(new(big.Int).SetBytes(a), new(big.Int).SetBytes(b)).Bytes()
}

func SubBigIntAbs(a []byte, b []byte) []byte {
	return new(big.Int).Sub(new(big.Int).SetBytes(a), new(big.Int).SetBytes(b)).Bytes()
}

func SafeSubBigInt(amt []byte, sub []byte) (res, actualSub []byte,  exceed bool) {
	if CmpForBigInt(sub, amt) > 0 {
		res = nil
		actualSub = amt
		exceed = true
	} else {
		res = SubBigIntAbs(amt, sub)
		actualSub = sub
	}
	return
}

func SubBigInt(a []byte, b []byte) *big.Int {
	return new(big.Int).Sub(new(big.Int).SetBytes(a), new(big.Int).SetBytes(b))
}

func MinBigInt(a []byte, b []byte) []byte {
	if new(big.Int).SetBytes(a).Cmp(new(big.Int).SetBytes(b)) > 0 {
		return b
	} else {
		return a
	}
}

func CmpToBigZero(a []byte) int {
	return new(big.Int).SetBytes(a).Sign()
}

func CmpForBigInt(a []byte, b []byte) int {
	if len(a) == 0 && len(b) == 0 {
		return 0
	} else if len(a) == 0 {
		return -1
	} else if len(b) == 0 {
		return 1
	}
	return new(big.Int).SetBytes(a).Cmp(new(big.Int).SetBytes(b))
}

func GetAbs(v int32) (int32, int32) { //abs, sign
	if v < 0 {
		return -v, -1
	} else {
		return v, 1
	}
}

func NegativeAmount(amount []byte) *big.Int {
	return new(big.Int).Neg(new(big.Int).SetBytes(amount))
}

func CardinalRateToString(rawRate, rateCardinalNum int32) string {
	rateStr := strconv.Itoa(int(rawRate))
	cardinalStr := strconv.Itoa(int(rateCardinalNum))
	rateLen := len(rateStr)
	cardinalLen := len(cardinalStr)
	if rateLen >= cardinalLen {
		return fmt.Sprintf("%s.%s", rateStr[0:rateLen-cardinalLen+1], rateStr[rateLen-cardinalLen+1:cardinalLen])
	} else {
		decimalPartArr := make([]byte, cardinalLen-1)
		if rateLen == cardinalLen-1 {
			copy(decimalPartArr, rateStr)
		} else {
			//left pad 0
			for i := 0; i < cardinalLen-rateLen-1; i++ {
				decimalPartArr[i] = '0'
			}
			copy(decimalPartArr[cardinalLen-1-rateLen:], rateStr)
		}
		return fmt.Sprintf("0.%s", decimalPartArr)
	}
}

func CalculateAmountForRate(amount []byte, rate, rateCardinalNum int32) []byte {
	if rate > 0 {
		amtF := new(big.Float).SetPrec(BigFloatPrec).SetInt(new(big.Int).SetBytes(amount))
		rateF, _ := new(big.Float).SetPrec(BigFloatPrec).SetString(CardinalRateToString(rate, rateCardinalNum))
		return RoundAmount(new(big.Float).SetPrec(BigFloatPrec).Mul(amtF, rateF)).Bytes()
	} else {
		return nil
	}
}

func AdjustForDecimalsDiff(sourceAmountF *big.Float, decimalsDiff int32) *big.Float {
	if decimalsDiff == 0 {
		return sourceAmountF
	}
	dcDiffAbs, dcDiffSign := GetAbs(decimalsDiff)
	decimalDiffInt := new(big.Int).Exp(helper.Big10, new(big.Int).SetUint64(uint64(dcDiffAbs)), nil)
	decimalDiffFloat := new(big.Float).SetPrec(BigFloatPrec).SetInt(decimalDiffInt)
	if dcDiffSign > 0 {
		return sourceAmountF.Quo(sourceAmountF, decimalDiffFloat)
	} else {
		return sourceAmountF.Mul(sourceAmountF, decimalDiffFloat)
	}
}

func AdjustAmountForDecimalsDiff(amount []byte, decimalsDiff int32) *big.Int {
	return RoundAmount(AdjustForDecimalsDiff(new(big.Float).SetPrec(BigFloatPrec).SetInt(new(big.Int).SetBytes(amount)), decimalsDiff))
}

func RoundAmount(amountF *big.Float) *big.Int {
	amount, _ := new(big.Float).SetPrec(BigFloatPrec).Add(amountF, big.NewFloat(0.5)).Int(nil)
	return amount
}
