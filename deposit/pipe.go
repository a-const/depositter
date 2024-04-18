package deposit

import (
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
)

type (
	TxFn         func(*Deposit) *types.Transaction
	TxPipe       func(chan *Deposit) chan *types.Transaction
	MakeElemFn   func(*types.Transaction) *BatchElement
	MakeElemPipe func(chan *types.Transaction) chan *BatchElement
	AppendFn     func(*BatchElement)
	AppendPipe   func(chan *BatchElement)
)

func TxAddToPipe(tf TxFn) TxPipe {
	return func(input chan *Deposit) chan *types.Transaction {
		output := make(chan *types.Transaction)
		var wg sync.WaitGroup
		wg.Add(100)
		go func() {
			wg.Wait()
			close(output)
		}()

		for i := 0; i < 100; i++ {
			go func() {
				defer wg.Done()
				for in := range input {
					output <- tf(in)
				}
			}()
		}

		return output
	}
}

func BathcElemAddToPipe(tf MakeElemFn) MakeElemPipe {
	return func(input chan *types.Transaction) chan *BatchElement {
		output := make(chan *BatchElement)
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
	return func(input chan *BatchElement) {
		for in := range input {
			tf(in)
		}
	}
}

func MakePipe(input chan *Deposit, tx TxFn, me MakeElemFn, a AppendFn) {
	output := input
	txOutput := TxAddToPipe(tx)(output)
	batchOutput := BathcElemAddToPipe(me)(txOutput)
	AppendAddToPipe(a)(batchOutput)
}
