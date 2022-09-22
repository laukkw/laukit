package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/laukkw/kwstart/errors"
	"github.com/laukkw/laukit/client"
	"golang.org/x/net/context"
	"math/big"
)

type WalletProvider struct {
	PrivateKey *ecdsa.PrivateKey
	provider   *client.Provider
}

func NewWalletProvider(privateKey string, provider *client.Provider) (*WalletProvider, error) {
	private, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("private key err %v", err.Error()))
	}

	return &WalletProvider{private, provider}, nil
}

func (w *WalletProvider) Backend() *client.Provider {
	return w.provider
}

func (w *WalletProvider) NewTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	gasPrice, err := w.provider.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(w.PrivateKey, w.provider.ChainId)
	if err != nil {
		return nil, err
	}
	auth.Context = ctx
	auth.Value = big.NewInt(0)
	auth.GasLimit = 0
	auth.GasPrice = gasPrice
	nonce, err := w.GetNonce(ctx)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	return auth, nil
}

func (w *WalletProvider) GetNonce(ctx context.Context) (uint64, error) {
	return w.provider.PendingNonceAt(ctx, crypto.PubkeyToAddress(w.PrivateKey.PublicKey))
}
