package main

import (
	"testing"
)

func TestAbiDecodeExprAndStringify(t *testing.T) {
	t.Run("uint256", func(t *testing.T) {
		input := "0000000000000000000000008c43fbebaa2ded5a50c10766b0f03a151f2bbf17000000000000000000000000487ee5d805b3c95eb23055dc92aad29a89961f170000000000000000000000000000000000000000000000001bc16d674ec80000"
		t.Log(AbiDecodeExprAndStringify("address,address,uint256", MustDecodeString(input)))
	})
}
