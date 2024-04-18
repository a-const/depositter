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
	pool       *sync.Pool
	wg         *sync.WaitGroup
	dc         *manager.DepositContract
	p          *Parser
	startNonce uint64
	nonce      *atomic.Int64

	batch [][]rpc.BatchElem
}

type BatchElement struct {
	elem  rpc.BatchElem
	nonce uint64
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
	b.startNonce = n
	b.nonce = &atomic.Int64{}
	b.nonce.Add(int64(n))
	b.batch = make([][]rpc.BatchElem, length/500+1)
	for i := 0; i < len(b.batch); i++ {
		b.batch[i] = make([]rpc.BatchElem, 500)
	}

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
	input := make(chan *Deposit)
	go func() {
		defer close(input)
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

func (b *Builder) MakeElem(tx *types.Transaction) *BatchElement {
	bin, err := tx.MarshalBinary()

	if err != nil {
		log.Error("Error marshaling tx to binary")
	}

	elem := rpc.BatchElem{
		Method: "eth_sendRawTransaction",
		//Method: "eth_estimateGas",
		Args: []any{hexutil.Encode(bin)},
	}

	return &BatchElement{
		elem:  elem,
		nonce: tx.Nonce(),
	}
}

func (b *Builder) AppendElem(elem *BatchElement) {
	part := (elem.nonce - b.startNonce) / 500
	index := (elem.nonce - b.startNonce) % 500
	b.batch[part][index] = elem.elem
	log.Infof("Building batch. Batch[%d][%d]", part, index)
}
