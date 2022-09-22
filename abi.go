package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

func AbiDecodeExprAndStringify(expr string, input []byte) ([]string, error) {
	argsList := parseArgumentExpr(expr)
	var argTypes []string
	for _, v := range argsList {
		argTypes = append(argTypes, v.Type)
	}

	return AbiMarshalStringValues(argTypes, input)
}

func AbiMarshalStringValues(argTypes []string, input []byte) ([]string, error) {
	values, err := AbiDecoderWithReturnedValues(argTypes, input)
	if err != nil {
		return nil, err
	}
	return StringifyValues(values)
}
func AbiDecoderWithReturnedValues(argTypes []string, input []byte) ([]interface{}, error) {
	args, err := buildArgumentsFromTypes(argTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to build abi: %v", err)
	}
	return args.UnpackValues(input)
}

func parseArgumentExpr(expr string) []abiArgument {
	var args []abiArgument
	expr = strings.Trim(expr, "() ")
	p := strings.Split(expr, ",")

	if expr == "" {
		return args
	}
	for _, v := range p {
		v = strings.Trim(v, " ")
		n := strings.Split(v, " ")
		arg := abiArgument{Type: n[0]}
		if len(n) > 1 {
			arg.Name = n[1]
		}
		args = append(args, arg)
	}
	return args
}
func buildArgumentsFromTypes(argTypes []string) (abi.Arguments, error) {
	args := abi.Arguments{}
	for _, argType := range argTypes {
		abiType, err := abi.NewType(argType, "", nil)
		if err != nil {
			return nil, err
		}
		args = append(args, abi.Argument{Type: abiType})
	}
	return args, nil
}

type abiArgument struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
}

func StringifyValues(values []interface{}) ([]string, error) {
	strs := []string{}

	for _, value := range values {
		stringer, ok := value.(fmt.Stringer)
		if ok {
			strs = append(strs, stringer.String())
			continue
		}

		switch v := value.(type) {
		case nil:
			strs = append(strs, "")
			break

		case string:
			strs = append(strs, v)
			break

		default:
			strs = append(strs, fmt.Sprintf("%v", value))
			break
		}
	}

	return strs, nil
}
