package monitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/laukkw/kwstart/errors"
	"sync"
)

var (
	ErrFatal                 = errors.New("ethmonitor: fatal error, stopping.")
	ErrUnexpectedParentHash  = errors.New("ethmonitor: unexpected parent hash")
	ErrUnexpectedBlockNumber = errors.New("ethmonitor: unexpected block number")
	ErrQueueFull             = errors.New("ethmonitor: publish queue is full")
	ErrMaxAttempts           = errors.New("ethmonitor: max attempts hit")
)

type Blocks []*Block

type Chain struct {
	blocks         Blocks
	retentionLimit int
	mu             sync.Mutex
}

func newChain(retentionLimit int) *Chain {
	return &Chain{
		blocks:         make(Blocks, 0, retentionLimit),
		retentionLimit: retentionLimit,
	}
}

func (c *Chain) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blocks = c.blocks[:0]
}
func (c *Chain) push(nextBlock *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := len(c.blocks)
	if n > 0 {
		headBlock := c.blocks[n-1]
		if nextBlock.ParentHash() != headBlock.Hash() {
			return ErrUnexpectedParentHash
		}
		if nextBlock.NumberU64() != headBlock.NumberU64()+1 {
			return ErrUnexpectedBlockNumber
		}
	}
	c.blocks = append(c.blocks, nextBlock)
	if len(c.blocks) > c.retentionLimit {
		c.blocks[0] = nil
		c.blocks = c.blocks[1:]
	}
	return nil
}
func (c *Chain) pop() *Block {
	c.mu.Lock()
	defer c.mu.Unlock()

	n := len(c.blocks) - 1
	block := c.blocks[n]
	c.blocks[n] = nil
	c.blocks = c.blocks[:n]
	return block
}

func (c *Chain) Head() *Block {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.blocks.Head()
}

func (c *Chain) Tail() *Block {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.blocks.Tail()
}

func (c *Chain) Blocks() Blocks {
	c.mu.Lock()
	defer c.mu.Unlock()
	blocks := make(Blocks, len(c.blocks))
	copy(blocks, c.blocks)
	return blocks
}

func (c *Chain) GetBlock(hash common.Hash) *Block {
	c.mu.Lock()
	defer c.mu.Unlock()
	block, _ := c.blocks.FindBlock(hash)
	return block
}
func (c *Chain) GetBlockByNumber(blockNum uint64, event Event) *Block {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.blocks) - 1; i >= 0; i-- {
		if c.blocks[i].NumberU64() == blockNum && c.blocks[i].Event == event {
			return c.blocks[i]
		}
	}
	return nil
}

func (c *Chain) GetTransaction(hash common.Hash) *types.Transaction {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.blocks) - 1; i >= 0; i-- {
		for _, txn := range c.blocks[i].Transactions() {
			if txn.Hash() == hash {
				return txn
			}
		}
	}
	return nil
}

func (c *Chain) PrintAllBlocks() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, b := range c.blocks {
		fmt.Printf("<- [%d] %s\n", b.NumberU64(), b.Hash().Hex())
	}
}

type Block struct {
	*types.Block
	Event Event
	Logs  []types.Log
	OK    bool
}

type Event uint32

const (
	Added Event = iota
	Removed
)

func (blocks Blocks) LatestBlock() *Block {
	for i := len(blocks) - 1; i >= 0; i-- {
		if blocks[i].Event == Added {
			return blocks[i]
		}
	}
	return nil
}

func (blocks Blocks) Head() *Block {
	if len(blocks) == 0 {
		return nil
	}
	return blocks[len(blocks)-1]
}

func (blocks Blocks) Tail() *Block {
	if len(blocks) == 0 {
		return nil
	}
	return blocks[0]
}

func (blocks Blocks) IsOK() bool {
	for _, block := range blocks {
		if !block.OK {
			return false
		}
	}
	return true
}

func (blocks Blocks) Reorg() bool {
	for _, block := range blocks {
		if block.Event == Removed {
			return true
		}
	}
	return false
}

func (blocks Blocks) FindBlock(hash common.Hash, optEvent ...Event) (*Block, bool) {
	for i := len(blocks) - 1; i >= 0; i-- {
		if blocks[i].Hash() == hash {
			if optEvent == nil {
				return blocks[i], true
			} else if len(optEvent) > 0 && blocks[i].Event == optEvent[0] {
				return blocks[i], true
			}
		}
	}
	return nil, false
}

func (blocks Blocks) EventExists(block *types.Block, event Event) bool {
	b, ok := blocks.FindBlock(block.Hash(), event)
	if !ok {
		return false
	}
	if b.ParentHash() == block.ParentHash() && b.NumberU64() == block.NumberU64() {
		return true
	}
	return false
}
func (blocks Blocks) Copy() Blocks {
	nb := make(Blocks, len(blocks))

	for i, b := range blocks {
		var logs []types.Log
		if b.Logs != nil {
			copy(logs, b.Logs)
		}
		nb[i] = &Block{
			Block: b.Block,
			Event: b.Event,
			Logs:  logs,
			OK:    b.OK,
		}
	}

	return nb
}
