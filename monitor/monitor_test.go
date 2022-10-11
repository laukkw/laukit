package monitor

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/laukkw/kwstart/kwlog"
	"github.com/laukkw/laukit/client"
	"golang.org/x/net/context"
	"math/big"
	"testing"
)

func TestMonitor(t *testing.T) {
	//测试Monitor 扫区块数据
	var bsc = []string{"https://data-seed-prebsc-1-s2.binance.org:8545",
		"https://bsc-dataseed1.ninicoin.io"}

	p, err := client.NewProvider(bsc, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p.ChainID(p.Ctx))
	logops := kwlog.Options{
		OutputPaths:       []string{"./test_monitor_log", "stdout"},
		ErrorOutputPaths:  []string{"./test_monitor_err"},
		Level:             kwlog.DebugLevel.String(),
		Format:            "console",
		DisableCaller:     false,
		DisableStacktrace: false,
		EnableColor:       false,
		Development:       true,
	}
	nowBlockNumber, _ := p.BlockNumber(context.Background())

	opts := DefaultOptions
	opts.Logger = kwlog.New(&logops)
	opts.StartBlockNumber = big.NewInt(23224753)
	opts.BlockRetentionLimit = 10
	opts.LogTopics = []common.Hash{common.HexToHash("0x856999cd0ef0485cf4c165e3c3d3cb68b168d825a77eb23eeed123f97d583ce1")}
	opts.WithLogs = true

	monitor, err := NewMonitor(p, opts)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := monitor.Run(p.Ctx)
		if err != nil {
			return
		}
	}()

	defer monitor.Stop()
	sub := monitor.Subscribe()
	defer sub.Unsubscribe()
	for {
		select {
		case blocks := <-sub.Blocks():
			//_ = blocks
			for _, b := range blocks {
				if b.NumberU64() == nowBlockNumber+2 {
					return
				}
				/*for _, v := range b.Transactions() {
					from, err := types.Sender(types.NewLondonSigner(v.ChainId()), v)
					if err != nil {
						t.Fatal(err)
					}
					t.Log("Hash --> ", v.Hash(), "User --> ", from)
				}*/
				for _, v := range b.Logs {
					//t.Logf(" %+v ", v)
					t.Log(hex.EncodeToString(v.Data))
					t.Log(v.Data)
				}
			}
		case <-sub.Done():
			return
		}
	}
}
