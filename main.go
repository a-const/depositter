package main

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	// Basic app settings
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	app := cli.App{}

	app.Name = "FTN Depositter"
	app.Usage = "Tool for FTN deposits"
	app.Version = "1.0"

	appFlags := make([]cli.Flag, 0, 4)
	URLFlag := &cli.StringFlag{
		Required: true,
		Name:     "url",
		Usage:    "URL of execution client RPC",
	}
	ChainIDFlag := &cli.Int64Flag{
		Required: true,
		Name:     "chainid",
		Usage:    "Chain id",
	}
	PrivateFlag := &cli.StringFlag{
		Required: true,
		Name:     "private",
		Usage:    "Private key of account, responsible for deploy and deposits",
	}
	PublicFlag := &cli.StringFlag{
		Required: true,
		Name:     "public",
		Usage:    "Public key of account, responsible for deploy and deposits",
	}
	DepositFileFlag := &cli.StringFlag{
		Required: true,
		Name:     "deposit-file",
		Usage:    "Name of file, which contains deposit data",
	}

	appFlags = append(appFlags, URLFlag, ChainIDFlag, PrivateFlag, PublicFlag, DepositFileFlag)
	app.Flags = appFlags

	//New contract scenario
	newContractCommand := &cli.Command{
		Name: "new-contract",
		Action: func(ctx *cli.Context) error {
			dc := NewDepositContract(ctx,
				ctx.String(URLFlag.Name),
				ctx.Int64(ChainIDFlag.Name),
				ctx.String(PrivateFlag.Name),
				ctx.String(PublicFlag.Name),
			)
			dc.Deploy(ctx)
			log.Infof("Contract created with address: %s", dc.Address.String())

			parser := NewParser()
			parser.Parse(ctx, ctx.String(DepositFileFlag.Name))
			batch := parser.BuildBatch(ctx, dc)
			log.Info("Sending batch...")
			for _, b := range batch {
				if err := dc.client.Client().BatchCallContext(ctx.Context, b); err != nil {
					log.Error("Error sending txs")
				}

			}
			return nil
		},
	}
	//Existing contract scenario
	existingContractFlags := make([]cli.Flag, 0, 1)
	contractAddressFlag := &cli.StringFlag{
		Name:     "contract-address",
		Required: true,
		Usage:    "Address of deposit contract (without 0x)",
	}
	existingContractFlags = append(existingContractFlags, contractAddressFlag)
	existingContractCommand := &cli.Command{
		Name:  "existing-contract",
		Flags: existingContractFlags,
		Action: func(ctx *cli.Context) error {
			dc := NewDepositContract(ctx,
				ctx.String(URLFlag.Name),
				ctx.Int64(ChainIDFlag.Name),
				ctx.String(PrivateFlag.Name),
				ctx.String(PublicFlag.Name),
			)

			ctrAddr := common.Hex2Bytes(ctx.String(contractAddressFlag.Name))

			dc.Bind(ctx, common.Address(ctrAddr))

			parser := NewParser()
			parser.Parse(ctx, ctx.String(DepositFileFlag.Name))
			batch := parser.BuildBatch(ctx, dc)
			log.Info("Sending batch...")
			for _, b := range batch {
				if err := dc.client.Client().BatchCallContext(ctx.Context, b); err != nil {
					log.Error("Error sending txs")
				}
			}

			return nil
		},
	}

	app.Commands = []*cli.Command{
		newContractCommand,
		existingContractCommand,
	}

	err := app.RunContext(context.TODO(), os.Args)
	if err != nil {
		log.Fatalf("can't start app! err: %s", err)
	}

}
