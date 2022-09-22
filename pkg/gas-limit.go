package pkg

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/net/context"
)

// SendTransactor 发送前先获取一遍GasLimit. 再发送交易.
func SendTransactor(fn func() (*types.Transaction, error), auth *bind.TransactOpts) (*types.Transaction, error) {
	auth.NoSend = true
	tx, err := fn()
	if err != nil {
		return nil, err
	}
	auth.NoSend = false
	auth.GasLimit = tx.Gas() + 100
	return fn()
}

func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
