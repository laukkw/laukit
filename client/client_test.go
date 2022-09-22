package client

import "testing"

func TestNewProvider(t *testing.T) {
	var bsc = []string{"https://data-seed-prebsc-1-s2.binance.org:8545",
		"https://bsc-dataseed1.ninicoin.io"}

	p, err := NewProvider(bsc, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p.ChainID(p.Ctx))

}
