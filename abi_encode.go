package laukit

import (
	"fmt"
	"github.com/laukkw/kwstart/errors"
)

// AbiCoder 将数据按类型 encode 返回bytes
func AbiCoder(argTypes []string, argValues []interface{}) ([]byte, error) {
	if len(argTypes) != len(argValues) {
		return nil, errors.New("invalid arguments - types and values do not match")
	}
	args, err := buildArgumentsFromTypes(argTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to build abi: %v", err)
	}
	return args.Pack(argValues...)
}
