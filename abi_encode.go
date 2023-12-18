package laukit

import (
    "fmt"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/laukkw/kwstart/errors"
    "strings"
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

func ParseABI(abiJSON string) (abi.ABI, error) {
    parsed, err := abi.JSON(strings.NewReader(abiJSON))
    if err != nil {
        return abi.ABI{}, fmt.Errorf("unable to parse abi json: %w", err)
    }
    return parsed, nil
}

func MustParseABI(abiJSON string) abi.ABI {
    parsed, err := ParseABI(abiJSON)
    if err != nil {
        panic(err)
    }
    return parsed
}

func EncodeInputData(abi abi.ABI, method string, args ...interface{}) ([]byte, error) {
    m, ok := abi.Methods[method]
    if !ok {
        return nil, fmt.Errorf("contract method %s not found", method)
    }
    input, err := m.Inputs.Pack(args...)
    if err != nil {
        return nil, err
    }
    input = append(m.ID, input...)
    return input, nil
}
