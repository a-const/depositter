package deposit

import (
	"context"
	"depositter/manager"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
)

type Builder struct {
	pool  *sync.Pool
	wg    *sync.WaitGroup
	dc    *manager.DepositContract
	p     *Parser
	nonce *atomic.Int64

	batch [][]rpc.BatchElem
	part  int
	index int
}

func NewBuilder(ctx context.Context, dc *manager.DepositContract, p *Parser, length int) *Builder {
	b := &Builder{
		dc: dc,
		pool: &sync.Pool{
			New: func() any {
				return dc.CopyTransactor()
			},
		},
	}
	n, err := b.dc.Client.PendingNonceAt(ctx, dc.PublicCommon)
	if err != nil {
		log.Fatal("Cannot retreive nonce")
	}
	b.nonce = &atomic.Int64{}
	b.nonce.Add(int64(n))
	b.batch = make([][]rpc.BatchElem, length/500+1)
	for i := 0; i < len(b.batch); i++ {
		b.batch[i] = make([]rpc.BatchElem, 500)
	}
	b.part = 0
	b.index = 0
	b.p = p
	return b
}

func (b *Builder) BuildBatch() [][]rpc.BatchElem {
	input := b.Loader()
	MakePipe(input, b.MakeTx, b.MakeElem, b.AppendElem)
	return b.batch
}

func (b *Builder) get() *bind.TransactOpts {
	return b.pool.Get().(*bind.TransactOpts)
}

func (b *Builder) put(txer *bind.TransactOpts) {
	b.pool.Put(txer)
}

func (b *Builder) Loader() chan *Deposit {
	input := make(chan *Deposit, len(b.p.Deposits))
	go func() {
		for _, d := range b.p.Deposits {
			input <- d
		}
	}()
	return input
}

func (b *Builder) MakeTx(d *Deposit) *types.Transaction {
	txor := b.get()
	txor.NoSend = true
	txor.Value, _ = new(big.Int).SetString("8192000000000000000000", 10)
	txor.GasLimit = 2_000_000
	txor.Nonce = big.NewInt(b.nonce.Load())
	b.nonce.Add(1)
	tx, err := b.dc.Contract.Deposit(
		txor,
		d.PubKey,
		d.WithdrawalCredential,
		d.ContractAddress,
		d.Signature,
		d.DepositDataRoot,
	)
	if err != nil {
		log.Errorf("Error building batch element. Error: %s", err)
	}
	b.put(txor)
	return tx
}

func (b *Builder) MakeElem(tx *types.Transaction) rpc.BatchElem {
	bin, err := tx.MarshalBinary()
	if err != nil {
		log.Error("Error marshaling tx to binary")
	}

	elem := rpc.BatchElem{
		Method: "eth_sendRawTransaction",
		//Method: "eth_estimateGas",
		Args: []any{hexutil.Encode(bin)},
	}
	return elem
}

func (b *Builder) AppendElem(elem rpc.BatchElem) {
	b.batch[b.part][b.index] = elem
	log.Infof("Building batch. Batch[%d][%d]", b.part, b.index)
	b.index++
	if b.index >= 500 {
		b.index = 0
		b.part++
	}
}
