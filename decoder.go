package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

// BytesToBytes32 将byte32 转换为 bytes32
func BytesToBytes32(slice []byte) [32]byte {
	var bytes32 [32]byte
	copy(bytes32[:], slice)
	return bytes32
}

// AddressPadding 转换为64位
func AddressPadding(input string) string {
	if strings.HasPrefix(input, "0x") {
		input = input[2:]
	}
	if len(input) < 64 {
		input = strings.Repeat("0", 64-len(input)) + input
	}
	return input[0:64]
}

// FunctionSignature 获取一个函数的signature
func FunctionSignature(functionExpr string) string {
	return hexutil.Encode(crypto.Keccak256([]byte(functionExpr))[0:4])
}

func MustDecodeString(h string) []byte {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic(fmt.Errorf("ethcoder: must hex decode but failed due to, %v", err))
	}
	return b
}
