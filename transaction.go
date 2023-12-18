package laukit

import (
    "context"
    "fmt"
    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/laukkw/kwstart/errors"
    "math/big"
    "time"
)

type TransactionReq struct {
    From       common.Address
    To         common.Address
    Nonce      *big.Int
    GasLimit   uint64
    GasPrice   *big.Int
    GasTip     *big.Int
    AccessList types.AccessList
    ETHValue   *big.Int
    Data       []byte
}
type WaitReceipt func(ctx context.Context) (*types.Receipt, error)

var errorPath = "Transaction package "

func EclNewTransaction(ctx context.Context, ecl *Ecl, req *TransactionReq) (*types.Transaction, error) {
    if ecl == nil || req == nil {
        return nil, fmt.Errorf("%s new transaction error: 请求为空", errorPath)
    }
    to := &req.To

    if req.Nonce == nil {
        nonce, err := ecl.PendingNonceAt(ctx, req.From)
        if err != nil {
            return nil, fmt.Errorf("%s ecl get pending nonce error %w", errorPath, err)
        }
        req.Nonce = big.NewInt(0).SetUint64(nonce)
    }
    if req.GasPrice == nil {
        gasPrice, err := ecl.SuggestGasPrice(ctx)
        if err != nil {
            return nil, fmt.Errorf("%s ecl get gas price error: %w", errorPath, err)
        }
        req.GasPrice = gasPrice
    }
    if req.GasLimit == 0 {
        callMsg := ethereum.CallMsg{
            From:     req.From,
            To:       to,
            Gas:      0, // estimating this value
            GasPrice: req.GasPrice,
            Value:    req.ETHValue,
            Data:     req.Data,
        }

        gasLimit, err := ecl.EstimateGas(ctx, callMsg)
        if err != nil {
            return nil, fmt.Errorf("%s ecl estimate gas error: %w", errorPath, err)
        }
        req.GasLimit = gasLimit
    }

    if to == nil && len(req.Data) == 0 {
        return nil, fmt.Errorf("%s ecl req error to Or data is nil", errorPath)
    }

    var rawTx *types.Transaction
    if req.GasTip != nil {
        chainId, err := ecl.ChainID(ctx)
        if err != nil {
            return nil, err
        }
        rawTx = types.NewTx(&types.DynamicFeeTx{
            ChainID:    chainId,
            To:         to,
            Nonce:      req.Nonce.Uint64(),
            Value:      req.ETHValue,
            GasFeeCap:  req.GasPrice,
            GasTipCap:  req.GasTip,
            Data:       req.Data,
            Gas:        req.GasLimit,
            AccessList: req.AccessList,
        })
    } else if req.AccessList != nil {
        chainId, err := ecl.ChainID(ctx)
        if err != nil {
            return nil, err
        }

        rawTx = types.NewTx(&types.AccessListTx{
            ChainID:    chainId,
            To:         to,
            Gas:        req.GasLimit,
            GasPrice:   req.GasPrice,
            Data:       req.Data,
            Nonce:      req.Nonce.Uint64(),
            Value:      req.ETHValue,
            AccessList: req.AccessList,
        })
    } else {
        rawTx = types.NewTx(&types.LegacyTx{
            To:       to,
            Gas:      req.GasLimit,
            GasPrice: req.GasPrice,
            Data:     req.Data,
            Nonce:    req.Nonce.Uint64(),
            Value:    req.ETHValue,
        })
    }
    return rawTx, nil
}

func EclSendTransaction(ctx context.Context, ecl *Ecl, signTx *types.Transaction) (*types.Transaction, WaitReceipt, error) {
    if ecl == nil {
        return nil, nil, fmt.Errorf("%s ecl client is nil", errorPath)
    }
    waitFn := func(ctx context.Context) (*types.Receipt, error) {
        return EclWaitReceipt(ctx, ecl, signTx.Hash())
    }

    return signTx, waitFn, ecl.SendTransaction(ctx, signTx)
}

func EclWaitReceipt(ctx context.Context, ecl *Ecl, txHash common.Hash) (*types.Receipt, error) {
    var clearTimeout context.CancelFunc
    if _, ok := ctx.Deadline(); !ok {
        ctx, clearTimeout = context.WithTimeout(ctx, 120*time.Second) // default timeout of 120 seconds
        defer clearTimeout()
    }

    for {
        select {
        case <-ctx.Done():
            if err := ctx.Err(); err != nil {
                return nil, fmt.Errorf("ethwallet, WaitReceipt for %v: %w", txHash, err)
            }
        default:
        }

        receipt, err := ecl.TransactionReceipt(ctx, txHash)
        if err != nil && !errors.Is(err, ethereum.NotFound) {
            return nil, err
        }

        if receipt != nil {
            return receipt, nil
        }

        time.Sleep(1 * time.Second)
    }

}
