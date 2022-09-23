package main

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestAbiDecodeExprAndStringify(t *testing.T) {
	t.Run("uint256", func(t *testing.T) {
		input := "0000000000000000000000008c43fbebaa2ded5a50c10766b0f03a151f2bbf17000000000000000000000000487ee5d805b3c95eb23055dc92aad29a89961f170000000000000000000000000000000000000000000000001bc16d674ec80000"
		t.Log(AbiDecodeExprAndStringify("address,address,uint256", MustDecodeString(input)))
	})

	t.Run("decode", func(t *testing.T) {
		input := "00000000000000000000000052ea2cc6746a7c2421e23b6f129e44201896d3ef0000000000000000000000008c43fbebaa2ded5a50c10766b0f03a151f2bbf170000000000000000000000000000000000000000000000008ac7230489e80000000000000000000000000000000000000000000000000000813385d2ced36c52000000000000000000000000000000000000000000000000000009184e729fff00000000000000000000000000000000000000000000000000038d7ea4c6800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006cae1bbeabbe40eee243ccf6b3d3a96d5579f237000000000000000000000000000000000000000000000000813385d2ced36c52000000000000000000000000000000000000000000000000000009184e729fff000000000000000000000000000000000000000000000000024f231efe124ee2000000000000000000000000000000000000000000000000000000000000001b"

		requset, err := AbiDecodeExprAndStringify("address,address,uint256,uint256,uint256,uint256,uint256,address,uint256,uint256,uint256,uint256", MustDecodeString(input))
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range requset {
			t.Logf("index %v --> %v", k, v)
		}

	})

}

func TestAbiDecoder(t *testing.T) {
	input := "0000000000000000000000008c43fbebaa2ded5a50c10766b0f03a151f2bbf17000000000000000000000000487ee5d805b3c95eb23055dc92aad29a89961f170000000000000000000000000000000000000000000000001bc16d674ec80000"
	var from, to common.Address
	var num *big.Int
	if err := AbiDecoder([]string{"address", "address", "uint256"},
		MustDecodeString(input), []interface{}{&from, &to, &num}); err != nil {
		t.Fatal(err)
	}
	t.Log(from.String(), to.String(), num.String())
}
