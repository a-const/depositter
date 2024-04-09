package cmd

import (
	"github.com/urfave/cli/v2"
)

type Flags struct {
	AppFlags              []cli.Flag
	ExistingContractFlags []cli.Flag
}

var (
	URLFlag             *cli.StringFlag
	ChainIDFlag         *cli.Int64Flag
	PrivateFlag         *cli.StringFlag
	PublicFlag          *cli.StringFlag
	DepositFileFlag     *cli.StringFlag
	contractAddressFlag *cli.StringFlag
)

func (f *Flags) SetAppFlags() {
	URLFlag = &cli.StringFlag{
		Required: true,
		Name:     "url",
		Usage:    "URL of execution Client RPC",
	}
	ChainIDFlag = &cli.Int64Flag{
		Required: true,
		Name:     "chainid",
		Usage:    "Chain id",
	}
	PrivateFlag = &cli.StringFlag{
		Required: true,
		Name:     "private",
		Usage:    "Private key of account, responsible for deploy and deposits",
	}
	PublicFlag = &cli.StringFlag{
		Required: true,
		Name:     "public",
		Usage:    "Public key of account, responsible for deploy and deposits",
	}
	DepositFileFlag = &cli.StringFlag{
		Required: true,
		Name:     "deposit-file",
		Usage:    "Name of file, which contains deposit data",
	}

	f.AppFlags = []cli.Flag{
		URLFlag,
		ChainIDFlag,
		PrivateFlag,
		PublicFlag,
		DepositFileFlag,
	}
}

func (f *Flags) SetExisitingContractFlags() {
	contractAddressFlag = &cli.StringFlag{
		Name:     "contract-address",
		Required: true,
		Usage:    "Address of deposit contract (without 0x)",
	}

	f.ExistingContractFlags = []cli.Flag{
		contractAddressFlag,
	}
}
