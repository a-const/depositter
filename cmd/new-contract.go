package cmd

import (
	"depositter/deposit"
	"depositter/manager"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func newContractCmd(ctx *cli.Context) error {
	dc := manager.NewDepositContract(ctx,
		ctx.String(URLFlag.Name),
		ctx.Int64(ChainIDFlag.Name),
		ctx.String(PrivateFlag.Name),
		ctx.String(AddressFlag.Name),
	)
	dc.Deploy(ctx)
	log.Infof("Contract created with address: %s", dc.Address.String())

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
