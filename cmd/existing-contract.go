package cmd

import (
	"depositter/deposit"
	"depositter/manager"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func existingContractCmd(ctx *cli.Context) error {
	dc := manager.NewDepositContract(ctx,
		ctx.String(URLFlag.Name),
		ctx.Int64(ChainIDFlag.Name),
		ctx.String(PrivateFlag.Name),
		ctx.String(PublicFlag.Name),
	)

	ctrAddr := common.Hex2Bytes(ctx.String(contractAddressFlag.Name))

	dc.Bind(ctx, common.Address(ctrAddr))

	parser := deposit.NewParser()
	parser.Parse(ctx, ctx.String(DepositFileFlag.Name))
	for _, d := range parser.Deposits {
		dc.Transactor.Value, _ = new(big.Int).SetString("8192000000000000000000", 10)
		dc.Transactor.GasLimit = 2_000_000
		tx, err := dc.Contract.Deposit(dc.Transactor, d.PubKey, d.WithdrawalCredential, d.ContractAddress, d.Signature, d.DepositDataRoot)
		if err != nil {
			log.Errorf("Error sending deposit, err: %s", err)
		}
		log.Printf("Tx hash: %s; Pubkey: %s", tx.Hash(), common.Bytes2Hex(d.PubKey))
	}

	//batch := parser.BuildBatch(ctx, dc)
	// log.Info("Sending batch...")
	// for _, b := range batch {
	// 	if err := dc.Client.Client().BatchCallContext(ctx.Context, b); err != nil {
	// 		log.Error("Error sending txs")
	// 	}
	// }

	return nil
}
