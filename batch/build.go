package batch

import (
	"context"
	"depositter/deposit"
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
	pool *sync.Pool
	wg   *sync.WaitGroup
	dc   *manager.DepositContract

	depChan  chan *deposit.Deposit
	txChan   chan *types.Transaction
	elemChan chan rpc.BatchElem

	nonce *atomic.Int64

	batch [][]rpc.BatchElem
	part  int
	index int
}

func NewBuilder(ctx context.Context, dc *manager.DepositContract, length int) *Builder {
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
	b.nonce.Add(int64(n))
	b.batch = make([][]rpc.BatchElem, length/500+1)
	for i := 0; i < len(b.batch); i++ {
		b.batch[i] = make([]rpc.BatchElem, 500)
	}
	b.part = 0
	b.index = 0
	return b
}

func (b *Builder) get() *bind.TransactOpts {
	return b.pool.Get().(*bind.TransactOpts)
}

func (b *Builder) put(txer *bind.TransactOpts) {
	b.pool.Put(txer)
}

func (b *Builder) Worker() {
	for in := range b.input {
		b.MakeTx(in)
	}
}

func (b *Builder) MakeTx(d *deposit.Deposit) {
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
	b.txChan <- tx
}

func (b *Builder) MakeElem() {
	tx := <-b.txChan
	bin, err := tx.MarshalBinary()
	if err != nil {
		log.Error("Error marshaling tx to binary")
	}

	elem := rpc.BatchElem{
		Method: "eth_sendRawTransaction",
		//Method: "eth_estimateGas",
		Args: []any{hexutil.Encode(bin)},
	}
	b.elemChan <- elem
}

func (b *Builder) AppendElem() {
	for elem := range b.elemChan {
		b.batch[b.part][b.index] = elem
		log.Infof("Building batch. Batch[%d][%d]", b.part, b.index)
		b.index++
		if b.index >= 500 {
			b.index = 0
			b.part++
		}
	}

}
