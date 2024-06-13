package cmd

import (
	"depositter/deposit"
	"depositter/manager"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func existingContractCmd(ctx *cli.Context) error {
	dc := manager.NewDepositContract(ctx,
		ctx.String(URLFlag.Name),
		ctx.Int64(ChainIDFlag.Name),
		ctx.String(PrivateFlag.Name),
		ctx.String(AddressFlag.Name),
	)

	ctrAddr := common.Hex2Bytes(ctx.String(contractAddressFlag.Name))

	dc.Bind(ctx, common.Address(ctrAddr))

	parser := deposit.NewParser()
	parser.Parse(ctx, ctx.String(DepositFileFlag.Name))
	batch := parser.BuildBatch(ctx, dc)
	log.Info("Sending batch...")
	for _, b := range batch {
		if err := dc.Client.Client().BatchCallContext(ctx.Context, b); err != nil {
			log.Error("Error sending txs")
		}
	}
	log.Info("Done!")
	return nil
}
