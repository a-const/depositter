package manager

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"depositter/contract"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type DepositContract struct {
	URL     string
	ChainID int64
	Private string
	Public  string

	Client     *ethclient.Client
	Transactor *bind.TransactOpts

	Ctx           context.Context
	Address       common.Address
	Tx            *types.Transaction
	Contract      *contract.Contract
	PrivateCommon common.Address
	PublicCommon  common.Address
}

func NewDepositContract(ctx *cli.Context, url string, chi int64, private string, public string) *DepositContract {
	return &DepositContract{
		URL:           url,
		ChainID:       chi,
		Private:       private,
		Public:        public,
		PrivateCommon: common.Address(common.Hex2Bytes(private)),
		PublicCommon:  common.Address(common.Hex2Bytes(public)),
	}
}

func (d *DepositContract) Init(ctx *cli.Context) {
	var (
		err     error
		PKBytes []byte
	)
	d.Client, err = ethclient.Dial(d.URL)
	//d.client.Client().
	if err != nil {
		log.Errorf("Failed to dial %s: %v\n", d.URL, err)
		ctx.Err()
	}
	log.Info("Client has been created")

	PKBytes, err = hex.DecodeString(d.Private)
	if err != nil {
		log.Errorf("Failed to decode string to hex: %v\n", err)
		ctx.Err()
	}
	log.Info("PK read")

	privateKey := crypto.ToECDSAUnsafe(PKBytes)
	d.Transactor, err = bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(d.ChainID))
	if err != nil {
		log.Errorf("Failed to build Transactor: %v\n", err)
		ctx.Err()
	}

}

func (d *DepositContract) Deploy(ctx *cli.Context) error {
	d.Init(ctx)
	addr, tx, ctr, err := contract.DeployContract(d.Transactor, d.Client)
	if err != nil {
		return err
	}

	d.Address = addr
	d.Tx = tx
	d.Contract = ctr
	return nil
}

func (d *DepositContract) Bind(ctx *cli.Context, ctrAddr common.Address) {
	d.Init(ctx)

	ctr, err := contract.NewContract(ctrAddr, d.Client)
	if err != nil {
		log.Fatalf("Failed to bind contract, err: %s", err)
		ctx.Err()
	}
	d.Contract = ctr
	log.Info("Contract bind")
}
