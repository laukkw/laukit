package laukit

import (
    "context"
    "crypto/ecdsa"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
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

func NewMcsHub(opts ...EauthOptions) *Eauth {
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
    return auth, nil
}

func (e *Eauth) GetNonce(ctx context.Context) (uint64, error) {
    return e.PendingNonceAt(ctx, crypto.PubkeyToAddress(e.Private.PublicKey))
}
