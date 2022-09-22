package monitor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/laukkw/kwstart/kwlog"
	"github.com/laukkw/laukit/client"
	"golang.org/x/net/context"
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
	//opts.StartBlockNumber = big.NewInt(23028092)
	opts.BlockRetentionLimit = 10
	opts.LogTopics = []common.Hash{common.HexToHash("0x627059660ea01c4733a328effb2294d2f86905bf806da763a89cee254de8bee5")}
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
				for _, v := range b.Transactions() {
					from, err := types.Sender(types.NewLondonSigner(v.ChainId()), v)
					if err != nil {
						t.Fatal(err)
					}
					t.Log("Hash --> ", v.Hash(), "User --> ", from)
				}
				for _, v := range b.Logs {
					t.Logf(" %+v ", v)
				}
			}
		case <-sub.Done():
			return
		}
	}
}
