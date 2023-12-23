package laukit

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "math/big"
)

type Eauth struct {
    Private *ecdsa.PrivateKey
    *Ecl
    context.Context
}

type EauthOptions func(*Eauth)

func WithPrivateKey(private string) EauthOptions {
    key, _ := crypto.HexToECDSA(private)
    return func(eauth *Eauth) {
        eauth.Private = key
    }
}

func WithEcl(ecl *Ecl) EauthOptions {
    return func(eauth *Eauth) {
        eauth.Ecl = ecl
    }
}

func NewEAuth(opts ...EauthOptions) *Eauth {
    b := &Eauth{}
    for _, o := range opts {
        o(b)
    }
    return b
}
func (e *Eauth) NewTransactor(ctx context.Context) (*bind.TransactOpts, error) {
    gasPrice, err := e.SuggestGasPrice(ctx)
    if err != nil {
        return nil, err
    }
    auth, err := bind.NewKeyedTransactorWithChainID(e.Private, e.ChainId)
    if err != nil {
        return nil, err
    }
    auth.Context = ctx
    auth.Value = big.NewInt(0)
    auth.GasLimit = 0
    auth.GasPrice = gasPrice
    nonce, err := e.GetNonce(ctx)
    if err != nil {
        return nil, err
    }
    auth.Nonce = big.NewInt(int64(nonce))
    tip, err := e.SuggestGasTipCap(ctx)
    if err != nil {
        return auth, nil
    } else {
        auth.GasTipCap = tip
        auth.GasFeeCap = auth.GasPrice
        // auth.GasPrice =
    }
    return auth, nil
}

func (e *Eauth) GetNonce(ctx context.Context) (uint64, error) {
    return e.PendingNonceAt(ctx, crypto.PubkeyToAddress(e.Private.PublicKey))
}

func (e *Eauth) NewTransactorNotPrivateKey(ctx context.Context, from string) (*bind.TransactOpts, error) {
    if e.Rpc == nil {
        return nil, fmt.Errorf("not init client")
    }
    gasPrice, err := e.SuggestGasPrice(ctx)
    if err != nil {
        return nil, err
    }
    nonce, err := e.PendingNonceAt(ctx, common.HexToAddress(from))
    if err != nil {
        return nil, err
    }
    resp := &bind.TransactOpts{
        From:     common.HexToAddress(from),
        Nonce:    big.NewInt(0).SetUint64(nonce),
        GasPrice: gasPrice,
        Context:  ctx,
        NoSend:   false,
    }
    tip, err := e.SuggestGasTipCap(ctx)
    if err != nil {
        // 获取失败的话就默认
        return resp, nil
    } else {
        resp.GasTipCap = tip
        resp.GasFeeCap = resp.GasPrice
        //resp.GasPrice = nil
    }
    return resp, nil
}
