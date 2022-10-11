package laukit

import (
	"encoding/hex"
	"math/big"
	"testing"
)

func TestAddressPadding(t *testing.T) {
	p := AddressPadding("0x8c43fbebaa2ded5a50c10766b0f03a151f2bbf17")
	t.Log(p)
	t.Log(hex.DecodeString(p))
}

func TestFunctionSignature(t *testing.T) {
	t.Log(FunctionSignature("swapExactTokensForTokensSupportingFeeOnTransferTokens(uint256,uint256,address[],address,uint256)"))
}

func TestBytesToBytes32(t *testing.T) {
	address := "8c43fbebaa2ded5a50c10766b0f03a151f2bbf17"
	addressbyte, err := hex.DecodeString(address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(BytesToBytes32(addressbyte))

	a := big.NewInt(123141)
	t.Log(BytesToBytes32(a.Bytes()))
}
