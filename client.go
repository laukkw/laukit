package laukit

import (
    "context"
    "fmt"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/rpc"
    "math/big"
)

type Ecl struct {
    *ethclient.Client
    Rpc     *rpc.Client
    Ctx     context.Context
    ChainId *big.Int
}

func NewEcl(ctx context.Context, url string) (*Ecl, error) {
    if ctx == nil {
        ctx = context.Background()
    }
    if url == "" {
        return nil, fmt.Errorf("rpc不可为空")
    }

    rpc, err := rpc.Dial(url)
    if err != nil {
        return nil, fmt.Errorf("rpc url 连接出错")
    }
    cli := ethclient.NewClient(rpc)
    chainId, err := cli.ChainID(ctx)
    if err != nil {
        return nil, fmt.Errorf("rpc url 连接出错")
    }

    ecl := &Ecl{
        Client:  cli,
        Rpc:     rpc,
        Ctx:     ctx,
        ChainId: chainId,
    }

    return ecl, nil

}
