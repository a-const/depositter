package main

import (
	"encoding/json"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type DepositJSON struct {
	PubKey               string `json:"pubkey"`
	WithdrawalCredential string `json:"withdrawal_credentials"`
	ContractAddress      string `json:"contract_address"`
	Signature            string `json:"signature"`
	DepositDataRoot      string `json:"deposit_data_root"`
	Amount               int64  `json:"amount"`
}

type Deposit struct {
	PubKey               []byte
	WithdrawalCredential []byte
	ContractAddress      []byte
	Signature            []byte
	DepositDataRoot      [32]byte
	Amount               int64
}

type Parser struct {
	Deposits []*Deposit
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(ctx *cli.Context, filename string) {
	dep := []DepositJSON{}
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Deposit file open error")
	}
	if err := json.NewDecoder(file).Decode(&dep); err != nil {
		log.Errorf("Error unmarshalling deposit data, err: %s", err)
	}
	p.Deposits = make([]*Deposit, len(dep))
	for i := 0; i < len(dep); i++ {
		p.Deposits[i] = &Deposit{}
	}
	for i, d := range dep {
		p.Deposits[i].PubKey = common.Hex2Bytes(d.PubKey)
		p.Deposits[i].WithdrawalCredential = common.Hex2Bytes(d.WithdrawalCredential)
		p.Deposits[i].ContractAddress = common.Hex2Bytes(d.ContractAddress)
		p.Deposits[i].Signature = common.Hex2Bytes(d.Signature)
		ddr := common.Hex2Bytes(d.DepositDataRoot)
		p.Deposits[i].DepositDataRoot = [32]byte(ddr)
		p.Deposits[i].Amount = d.Amount
	}
}

func (p *Parser) BuildBatch(ctx *cli.Context, dc *DepositContract) [][]rpc.BatchElem {
	nonce, err := dc.client.PendingNonceAt(ctx.Context, dc.PublicCommon)
	batch := make([][]rpc.BatchElem, len(p.Deposits)/500+1)
	for i := 0; i < len(batch); i++ {
		batch[i] = make([]rpc.BatchElem, 500)
	}
	if err != nil {
		log.Fatalf("Error retrieving pending nonce, err: %s", err)
		ctx.Err()
	}
	var (
		part  int = 0
		index int = 0
	)

	dc.transactor.NoSend = true
	dc.transactor.Value, _ = new(big.Int).SetString("8192000000000000000000", 10)
	dc.transactor.GasLimit = 2_000_000
	for i, d := range p.Deposits {
		dc.transactor.Nonce = big.NewInt(int64(nonce))
		nonce++
		tx, err := dc.Contract.Deposit(
			dc.transactor,
			d.PubKey,
			d.WithdrawalCredential,
			d.ContractAddress,
			d.Signature,
			d.DepositDataRoot,
		)
		if err != nil {
			log.Errorf("Error building batch element with index: %d. Error: %s", i, err)
		}
		bin, err := tx.MarshalBinary()
		if err != nil {
			log.Error("Error marshaling tx to binary")
		}

		elem := rpc.BatchElem{
			Method: "eth_sendRawTransaction",
			//Method: "eth_estimateGas",
			Args: []any{hexutil.Encode(bin)},
		}
		batch[part][index] = elem
		log.Infof("Building batch. Index: %d Batch[%d][%d]", i, part, index)
		index++
		if index >= 500 {
			index = 0
			part++
		}
	}
	log.Info("Bulding done!")
	return batch
}
