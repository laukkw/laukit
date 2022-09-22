package client

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
)

type Provider struct {
	*ethclient.Client
	Rpc     *rpc.Client
	Config  *Config
	Ctx     context.Context
	ChainId *big.Int
}

var _ bind.ContractBackend = &Provider{}

func NewProvider(url []string, test bool) (*Provider, error) {
	config := Config{
		Nodes: url,
		Test:  test,
	}
	p := &Provider{
		Config: &config,
		Ctx:    context.Background(),
	}
	err := p.Dial()
	if err != nil {
		return nil, err
	}
	return p, err
}

func (s *Provider) Dial() (err error) {
	switch s.Config.Test {
	case true:
		// test index = 0
		s.Rpc, err = rpc.Dial(s.Config.Nodes[0])
		if err != nil {
			return
		}
		s.Client = ethclient.NewClient(s.Rpc)
	case false:
		s.Rpc, err = rpc.Dial(s.Config.Nodes[1])
		if err != nil {
			return
		}
		s.Client = ethclient.NewClient(s.Rpc)
	}
	s.ChainId, err = s.Client.ChainID(context.Background())
	if err != nil {
		return
	}
	return
}
