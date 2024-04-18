package deposit

import (
	"encoding/json"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"depositter/manager"
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

func (p *Parser) BuildBatch(ctx *cli.Context, dc *manager.DepositContract) [][]rpc.BatchElem {
	return NewBuilder(ctx.Context, dc, p, len(p.Deposits)).BuildBatch()
}
