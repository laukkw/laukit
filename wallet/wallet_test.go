package wallet

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/laukkw/laukit/client"
	"github.com/laukkw/laukit/pkg"
	"github.com/laukkw/laukit/test/erc20"
	"math/big"
	"testing"
)

func TestInterface(t *testing.T) {
	var bsc = []string{"https://data-seed-prebsc-1-s2.binance.org:8545",
		"https://bsc-dataseed1.ninicoin.io"}

	p, err := client.NewProvider(bsc, true)
	if err != nil {
		t.Fatal(err)
	}
	wallet, err := NewWalletProvider("e7b609a399b0f88fe6054c4810dbaebec670643c16e43a1aea7e7ee8952b6206", p)
	if err != nil {
		t.Fatal(err)
	}
	ercInstance, err := erc20.NewStore(common.HexToAddress("0xC032904F1688e04F25a6918dFEe17c407E7F1c9f"), p.Client)
	if err != nil {
		t.Fatal(err)
	}
	auth, err := wallet.NewTransactor(p.Ctx)
	if err != nil {
		return
	}
	tx, err := pkg.SendTransactor(func() (*types.Transaction, error) {
		return ercInstance.Approve(auth, common.HexToAddress("0x4Ac19Ef38DB893a9128a49C654680A5DdC3F8202"), big.NewInt(1000000000000000))
	}, auth)
	t.Log(tx.Hash())

}
