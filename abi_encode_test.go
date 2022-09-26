package laukkit

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"testing"
)

func TestAbiCoder(t *testing.T) {
	result, err := AbiCoder([]string{"uint256", "uint256"},
		[]interface{}{big.NewInt(1), big.NewInt(2)})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hexutil.Encode(result))
}
