package deposit

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type (
	TxFn         func(*Deposit) *types.Transaction
	TxPipe       func(chan *Deposit) chan *types.Transaction
	MakeElemFn   func(*types.Transaction) rpc.BatchElem
	MakeElemPipe func(chan *types.Transaction) chan rpc.BatchElem
	AppendFn     func(rpc.BatchElem)
	AppendPipe   func(chan rpc.BatchElem)
)

func TxAddToPipe(tf TxFn) TxPipe {
	return func(input chan *Deposit) chan *types.Transaction {
		output := make(chan *types.Transaction)
		go func() {
			defer close(output)
			for in := range input {
				output <- tf(in)
			}
		}()
		return output
	}
}

func BathcElemAddToPipe(tf MakeElemFn) MakeElemPipe {
	return func(input chan *types.Transaction) chan rpc.BatchElem {
		output := make(chan rpc.BatchElem)
		go func() {
			defer close(output)
			for in := range input {
				output <- tf(in)
			}
		}()
		return output
	}
}

func AppendAddToPipe(tf AppendFn) AppendPipe {
	return func(input chan rpc.BatchElem) {
		go func() {
			for in := range input {
				tf(in)
			}
		}()
	}
}

func MakePipe(input chan *Deposit, tx TxFn, me MakeElemFn, a AppendFn) {
	output := input
	txOutput := TxAddToPipe(tx)(output)
	batchOutput := BathcElemAddToPipe(me)(txOutput)
	AppendAddToPipe(a)(batchOutput)
}
