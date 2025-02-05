package main

import (
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/venus-miner/node/modules/dtypes"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"

	miner0 "github.com/filecoin-project/specs-actors/actors/builtin/miner"

	lcli "github.com/filecoin-project/venus-miner/cli"
)

func isSupportedSectorSize(ssize abi.SectorSize) bool { // nolint
	for spf := range miner0.SupportedProofTypes {
		switch spf {
		case abi.RegisteredSealProof_StackedDrg2KiBV1:
			if ssize == 2048 {
				return true
			}
		case abi.RegisteredSealProof_StackedDrg8MiBV1:
			if ssize == 8<<20 {
				return true
			}
		case abi.RegisteredSealProof_StackedDrg512MiBV1:
			if ssize == 512<<20 {
				return true
			}
		case abi.RegisteredSealProof_StackedDrg32GiBV1:
			if ssize == 32<<30 {
				return true
			}
		case abi.RegisteredSealProof_StackedDrg64GiBV1:
			if ssize == 64<<30 {
				return true
			}
		default:

		}
	}

	return false
}

var addressCmd = &cli.Command{
	Name:  "address",
	Usage: "manage the miner address",
	Subcommands: []*cli.Command{
		updateCmd,
		listCmd,
		stateCmd,
		startMiningCmd,
		stopMiningCmd,
		addCmd,
	},
}

var updateCmd = &cli.Command{
	Name:  "update",
	Usage: "reacquire address from venus-auth",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:     "skip",
			Required: false,
		},
		&cli.Int64Flag{
			Name:     "limit",
			Required: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		postApi, closer, err := lcli.GetMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		skip := cctx.Int64("skip")
		limit := cctx.Int64("limit")

		miners, err := postApi.UpdateAddress(cctx.Context, skip, limit)
		if err != nil {
			return err
		}

		formatJson, err := json.MarshalIndent(miners, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(formatJson))

		return nil
	},
}

var addCmd = &cli.Command{
	Name:  "add",
	Usage: "add a miner",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "miner",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "id",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "name",
			Required: false,
		},
	},
	Action: func(cctx *cli.Context) error {
		mi := dtypes.MinerInfo{Id: cctx.String("id"), Name: cctx.String("name")}

		addr, err := address.NewFromString(cctx.String("miner"))
		if err != nil {
			return nil
		}
		mi.Addr = addr

		postApi, closer, err := lcli.GetMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		err = postApi.AddAddress(cctx.Context, mi)
		if err != nil {
			return err
		}

		fmt.Println("add miner success.")
		return nil

	},
}

var listCmd = &cli.Command{
	Name:  "list",
	Usage: "print miners",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		postApi, closer, err := lcli.GetMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		miners, err := postApi.ListAddress(cctx.Context)
		if err != nil {
			return err
		}

		formatJson, err := json.MarshalIndent(miners, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(formatJson))
		return nil

	},
}

var stateCmd = &cli.Command{
	Name:      "state",
	Usage:     "print state of mining",
	Flags:     []cli.Flag{},
	ArgsUsage: "[address ...]",
	Action: func(cctx *cli.Context) error {
		postApi, closer, err := lcli.GetMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		var addrs []address.Address
		for i, s := range cctx.Args().Slice() {
			minerAddr, err := address.NewFromString(s)
			if err != nil {
				return xerrors.Errorf("parsing %d-th miner: %w", i, err)
			}

			addrs = append(addrs, minerAddr)
		}

		states, err := postApi.StatesForMining(cctx.Context, addrs)
		if err != nil {
			return err
		}

		formatJson, err := json.MarshalIndent(states, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(formatJson))
		return nil

	},
}

var startMiningCmd = &cli.Command{
	Name:      "start",
	Usage:     "start mining for specified miner",
	Flags:     []cli.Flag{},
	ArgsUsage: "[address ...]",
	Action: func(cctx *cli.Context) error {
		postApi, closer, err := lcli.GetMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		var addrs []address.Address
		for i, s := range cctx.Args().Slice() {
			minerAddr, err := address.NewFromString(s)
			if err != nil {
				return xerrors.Errorf("parsing %d-th miner: %w", i, err)
			}

			addrs = append(addrs, minerAddr)
		}

		err = postApi.Start(ctx, addrs)
		if err != nil {
			return err
		}

		fmt.Println("start mining success.")
		return nil
	},
}

var stopMiningCmd = &cli.Command{
	Name:      "stop",
	Usage:     "stop mining for specified miner",
	Flags:     []cli.Flag{},
	ArgsUsage: "[address ...]",
	Action: func(cctx *cli.Context) error {
		postApi, closer, err := lcli.GetMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		var addrs []address.Address
		for i, s := range cctx.Args().Slice() {
			minerAddr, err := address.NewFromString(s)
			if err != nil {
				return xerrors.Errorf("parsing %d-th miner: %w", i, err)
			}

			addrs = append(addrs, minerAddr)
		}

		err = postApi.Stop(ctx, addrs)
		if err != nil {
			return err
		}

		fmt.Println("stop mining success.")
		return nil
	},
}
