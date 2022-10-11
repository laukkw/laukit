package monitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/laukkw/kwstart/errors"
	"github.com/laukkw/kwstart/kwlog"
	"github.com/laukkw/laukit/client"
	"golang.org/x/net/context"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
)

type Options struct {
	Logger                   kwlog.Logger
	PollingInterval          time.Duration
	Timeout                  time.Duration
	StartBlockNumber         *big.Int
	TrailNumBlocksBehindHead int
	BlockRetentionLimit      int
	WithLogs                 bool
	LogTopics                []common.Hash
	DebugLogging             bool
}

var DefaultOptions = Options{
	Logger:                   kwlog.New(kwlog.NewOptions()),
	PollingInterval:          1000 * time.Millisecond,
	Timeout:                  60 * time.Second,
	StartBlockNumber:         nil,
	TrailNumBlocksBehindHead: 0,
	BlockRetentionLimit:      200,
	WithLogs:                 false,
	LogTopics:                []common.Hash{},
	DebugLogging:             false,
}

type Monitor struct {
	options         Options
	log             kwlog.Logger
	provider        *client.Provider
	chain           *Chain
	nextBlockNumber *big.Int
	publishCh       chan Blocks
	publishQueue    *queue
	subscribers     []*subscriber
	ctx             context.Context
	ctxStop         context.CancelFunc
	running         int32
	mu              sync.RWMutex
}

func NewMonitor(provider *client.Provider, opts ...Options) (*Monitor, error) {
	options := DefaultOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	if options.Logger == nil {
		return nil, fmt.Errorf("monitor: logger is nil")
	}

	options.BlockRetentionLimit += options.TrailNumBlocksBehindHead

	return &Monitor{
		options:      options,
		log:          options.Logger,
		provider:     provider,
		chain:        newChain(options.BlockRetentionLimit),
		publishCh:    make(chan Blocks),
		publishQueue: newQueue(options.BlockRetentionLimit * 2),
		subscribers:  make([]*subscriber, 0),
	}, nil
}
func (m *Monitor) Run(ctx context.Context) error {
	if m.IsRunning() {
		return fmt.Errorf("monitor: already running")
	}

	m.ctx, m.ctxStop = context.WithCancel(ctx)

	atomic.StoreInt32(&m.running, 1)
	defer atomic.StoreInt32(&m.running, 0)

	if m.chain.Head() != nil {
		m.nextBlockNumber = m.chain.Head().Number()
	} else if m.options.StartBlockNumber != nil {
		m.nextBlockNumber = m.options.StartBlockNumber
	}

	if m.nextBlockNumber == nil {
		m.log.Info("monitor: starting from block=latest")
	} else {
		if m.chain.Head() == nil {
			m.log.Infof("monitor: starting from block=%d", m.nextBlockNumber)
		} else {
			m.log.Infof("monitor: starting from block=%d", m.nextBlockNumber.Uint64()+1)
		}
	}

	// Broadcast published events to all subscribers
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case blocks := <-m.publishCh:
				if m.options.DebugLogging {
					m.log.Debug(fmt.Sprintf("monitor: publishing block %v %v %v", blocks.LatestBlock().NumberU64(), "# events:", len(blocks)))
				}

				// broadcast to subscribers
				m.broadcast(blocks)
			}
		}
	}()

	// Monitor the chain for canonical representation
	return m.monitor()
}
func (m *Monitor) Stop() {
	m.log.Info("monitor: stop")
	m.ctxStop()
}

func (m *Monitor) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
}

func (m *Monitor) Options() Options {
	return m.options
}

func (m *Monitor) Provider() *client.Provider {
	return m.provider
}
func (m *Monitor) monitor() error {
	ctx := m.ctx
	events := Blocks{}

	// pollInterval is used for adaptive interval
	pollInterval := m.options.PollingInterval

	// monitor run loop
	for {
		select {

		case <-m.ctx.Done():
			return nil

		case <-time.After(pollInterval):
			headBlock := m.chain.Head()
			if headBlock != nil {
				m.nextBlockNumber = big.NewInt(0).Add(headBlock.Number(), big.NewInt(1))
			}

			nextBlock, err := m.fetchBlockByNumber(ctx, m.nextBlockNumber)
			if err == ethereum.NotFound {
				// reset poll interval as by config
				pollInterval = m.options.PollingInterval
				continue
			} else {
				// speed up the poll interval if we found the next block
				pollInterval /= 2
			}
			if err != nil {
				m.log.Warnf("monitor: [retrying] failed to fetch next block # %d, due to: %v", m.nextBlockNumber, err)
				pollInterval = m.options.PollingInterval // reset poll interval
				continue
			}

			events, err = m.buildCanonicalChain(ctx, nextBlock, events)
			if err != nil {
				m.log.Warnf("monitor: error reported '%v', failed to build chain for next blockNum:%d blockHash:%s, retrying..",
					err, nextBlock.NumberU64(), nextBlock.Hash().Hex())

				// pause, then retry
				time.Sleep(m.options.PollingInterval)
				continue
			}

			if m.options.WithLogs {
				m.addLogs(ctx, events)
				m.backfillChainLogs(ctx)
			} else {
				for _, b := range events {
					b.Logs = nil // nil it out to be clear to subscribers
					b.OK = true
				}
			}

			// publish events
			err = m.publish(ctx, events)
			if err != nil {
				return errors.New(fmt.Sprintf(ErrFatal.Error(), err))
			}

			// clear events sink
			events = Blocks{}
		}
	}
}

func (m *Monitor) buildCanonicalChain(ctx context.Context, nextBlock *types.Block, events Blocks) (Blocks, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	headBlock := m.chain.Head()

	m.log.Debugf("monitor: new block #%d hash:%s prevHash:%s numTens:%d",
		nextBlock.NumberU64(), nextBlock.Hash().String(), nextBlock.ParentHash().String(), len(nextBlock.Transactions()))

	if headBlock == nil || nextBlock.ParentHash() == headBlock.Hash() {
		// block-chaining it up
		block := &Block{Event: Added, Block: nextBlock}
		events = append(events, block)
		return events, m.chain.push(block)
	}

	poppedBlock := *m.chain.pop()
	poppedBlock.Event = Removed
	poppedBlock.OK = true

	m.log.Debugf("monitor: block reorg, reverting block #%d hash:%s prevHash:%s", poppedBlock.NumberU64(), poppedBlock.Hash().Hex(), poppedBlock.ParentHash().Hex())
	events = append(events, &poppedBlock)

	pause := m.options.PollingInterval * time.Duration(len(events))
	time.Sleep(pause)

	nextParentBlock, err := m.fetchBlockByHash(ctx, nextBlock.ParentHash())
	if err != nil {
		return events, err
	}

	events, err = m.buildCanonicalChain(ctx, nextParentBlock, events)
	if err != nil {
		return events, err
	}

	block := &Block{Event: Added, Block: nextBlock}
	err = m.chain.push(block)
	if err != nil {
		return events, err
	}
	events = append(events, block)

	return events, nil
}

func (m *Monitor) addLogs(ctx context.Context, blocks Blocks) {
	tctx, cancel := context.WithTimeout(ctx, m.options.Timeout)
	defer cancel()

	for _, block := range blocks {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if block.OK {
			continue
		}

		if block.Event == Removed {
			block.OK = true
			continue
		}

		blockHash := block.Hash()

		var topics [][]common.Hash
		if len(m.options.LogTopics) > 0 {
			topics = append(topics, m.options.LogTopics)
		}

		logs, err := m.provider.FilterLogs(tctx, ethereum.FilterQuery{
			BlockHash: &blockHash,
			Topics:    topics,
		})

		if err == nil {
			if len(logs) > 0 || block.Bloom() == (types.Bloom{}) {
				if logs == nil {
					block.Logs = []types.Log{}
				} else {
					block.Logs = logs
				}
				block.OK = true
				continue
			}
		}
		block.Logs = nil
		block.OK = false

		m.log.Infof("monitor: [getLogs failed -- marking block %s for log backfilling] %v", blockHash.Hex(), err)
	}
}

func (m *Monitor) backfillChainLogs(ctx context.Context) {
	blocks := m.chain.Blocks()

	for i := len(blocks) - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !blocks[i].OK {
			m.addLogs(ctx, Blocks{blocks[i]})
			if blocks[i].Event == Added && blocks[i].OK {
				m.log.Infof("monitor: [getLogs backfill successful for block:%d %s]", blocks[i].NumberU64(), blocks[i].Hash().Hex())
			}
		}
	}
}

func (m *Monitor) fetchBlockByNumber(ctx context.Context, num *big.Int) (*types.Block, error) {
	maxErrAttempts, errAttempts := 10, 0 // in case of node connection failures

	var block *types.Block
	var err error

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if errAttempts >= maxErrAttempts {
			m.log.Warnf("monitor: fetchBlockByNumber hit maxErrAttempts after %d tries for block num %v due to %v", errAttempts, num, err)
			return nil, errors.New(fmt.Sprintf(ErrMaxAttempts.Error(), err))
		}

		tctx, cancel := context.WithTimeout(ctx, m.options.Timeout)
		defer cancel()

		block, err = m.provider.BlockByNumber(tctx, num)
		if err != nil {
			if err == ethereum.NotFound {
				return nil, ethereum.NotFound
			} else {
				errAttempts++
				time.Sleep(m.options.PollingInterval * time.Duration(errAttempts) * 2)
				continue
			}
		}
		return block, nil
	}
}

func (m *Monitor) fetchBlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	maxNotFoundAttempts, notFoundAttempts := 4, 0
	maxErrAttempts, errAttempts := 10, 0

	var block *types.Block
	var err error

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if notFoundAttempts >= maxNotFoundAttempts {
			return nil, ethereum.NotFound
		}
		if errAttempts >= maxErrAttempts {
			m.log.Warnf("Monitor: fetchBlockByHash hit maxErrAttempts after %d tries for block hash %s due to %v", errAttempts, hash.Hex(), err)
			return nil, errors.New(fmt.Sprintf(ErrMaxAttempts.Error(), err))
		}

		block, err = m.provider.BlockByHash(ctx, hash)
		if err != nil {
			if err == ethereum.NotFound {
				notFoundAttempts++
				time.Sleep(m.options.PollingInterval * time.Duration(notFoundAttempts) * 2)
				continue
			} else {
				errAttempts++
				time.Sleep(m.options.PollingInterval * time.Duration(errAttempts) * 2)
				continue
			}
		}
		if block != nil {
			return block, nil
		}
	}
}

func (m *Monitor) publish(ctx context.Context, events Blocks) error {
	maxBlockNum := uint64(0)
	if m.options.TrailNumBlocksBehindHead > 0 {
		maxBlockNum = m.LatestBlock().NumberU64() - uint64(m.options.TrailNumBlocksBehindHead)
	}

	// Enqueue
	err := m.publishQueue.enqueue(events)
	if err != nil {
		return err
	}

	// Publish events existing in the queue
	pubEvents, ok := m.publishQueue.dequeue(maxBlockNum)
	if ok {
		m.publishCh <- pubEvents
	}

	return nil
}

func (m *Monitor) broadcast(events Blocks) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sub := range m.subscribers {
		select {
		case <-sub.done:
		case sub.sendCh <- events:
		}
	}
}

func (m *Monitor) Subscribe() Subscription {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan Blocks)
	subscriber := &subscriber{
		ch:     ch,
		sendCh: makeUnboundedBuffered(ch, m.log, 100),
		done:   make(chan struct{}),
	}

	subscriber.unsubscribe = func() {
		close(subscriber.done)
		m.mu.Lock()
		defer m.mu.Unlock()
		close(subscriber.sendCh)

		for ok := true; ok; _, ok = <-subscriber.ch {
		}

		for i, sub := range m.subscribers {
			if sub == subscriber {
				m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
				return
			}
		}
	}

	m.subscribers = append(m.subscribers, subscriber)

	return subscriber
}

func (m *Monitor) Chain() *Chain {
	m.chain.mu.Lock()
	defer m.chain.mu.Unlock()
	blocks := make(Blocks, len(m.chain.blocks))
	copy(blocks, m.chain.blocks)
	return &Chain{
		blocks: blocks,
	}
}

func (m *Monitor) LatestBlock() *Block {
	return m.chain.Head()
}

func (m *Monitor) GetBlock(hash common.Hash) *Block {
	return m.chain.GetBlock(hash)
}

func (m *Monitor) GetTransaction(hash common.Hash) *types.Transaction {
	return m.chain.GetTransaction(hash)
}
