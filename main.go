package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"runtime"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"

	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/logger"
)

const (
	Version = "0.0.2"
)

var (
	clilogger = logger.NewLogger("Parser")
	app       = utils.NewApp(Version, "Ethereum chain parser")
)

func init() {
	app.Action = run
	app.Name = "BlockParser"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "mongo-url",
			Value: "mongodb://localhost",
			Usage: "MongoDB connection url",
		},
		cli.StringFlag{
			Name:  "mongo-database",
			Value: "chain_explorer",
			Usage: "MongoDB database",
		},
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.DataDirFlag,
		utils.JSpathFlag,
		utils.ListenPortFlag,
		utils.LogFileFlag,
		utils.LogJSONFlag,
		utils.VerbosityFlag,
		utils.MaxPeersFlag,
		utils.EtherbaseFlag,
		utils.BlockchainVersionFlag,
		utils.MinerThreadsFlag,
		utils.MiningEnabledFlag,
		utils.NATFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.VMDebugFlag,
		utils.ProtocolVersionFlag,
		utils.NetworkIdFlag,
		utils.RPCCORSDomainFlag,
		utils.BacktraceAtFlag,
		utils.LogToStdErrFlag,
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer logger.Flush()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(ctx *cli.Context) {
	importer := NewImporter(ctx)
	utils.HandleInterrupt()

	cfg := utils.MakeEthConfig("EthChainParser", Version, ctx)

	ethereum, err := eth.New(cfg)
	if err != nil {
		utils.Fatalf("%v", err)
	}
	utils.StartEthereum(ethereum)

	if ctx.GlobalBool(utils.RPCEnabledFlag.Name) {
		utils.StartRPC(ethereum, ctx)
	}

	events := ethereum.EventMux().Subscribe(
		core.ChainEvent{},
		core.TxPreEvent{},
	)

	defer events.Unsubscribe()
	for {
		select {
		case ev, isopen := <-events.Chan():
			if !isopen {
				return
			}
			switch ev := ev.(type) {
			case core.ChainEvent:
				importer.importBlock(ev.Block)
			case core.TxPreEvent:
				// Not dealing with incoming txes for now
				//importer.importTx(ev.Tx)
			}
		}
	}

	ethereum.WaitForShutdown()
	logger.Flush()
	fmt.Printf("Shutting down\n")
}
