package interop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-chain-ops/interopgen"
	op_service "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
)

var EnvPrefix = "OP_INTEROP"

var (
	l1ChainIDFlag = &cli.Uint64Flag{
		Name:    "l1.chainid",
		Value:   900100,
		EnvVars: op_service.PrefixEnvVar(EnvPrefix, "L1_CHAINID"),
	}
	l2ChainIDsFlag = &cli.Uint64SliceFlag{
		Name:    "l2.chainids",
		Value:   cli.NewUint64Slice(900200, 900201),
		EnvVars: op_service.PrefixEnvVar(EnvPrefix, "L2_CHAINIDS"),
	}
	timestampFlag = &cli.Uint64Flag{
		Name:    "timestamp",
		Value:   0,
		EnvVars: op_service.PrefixEnvVar(EnvPrefix, "TIMESTAMP"),
		Usage:   "Will use current timestamp, plus 5 seconds, if not set",
	}
	mnemonicFlag = &cli.StringFlag{
		Name:    "mnemonic",
		Value:   devkeys.TestMnemonic,
		EnvVars: op_service.PrefixEnvVar(EnvPrefix, "MNEMONIC"),
	}
	artifactsDirFlag = &cli.StringFlag{
		Name:    "artifacts-dir",
		Value:   "packages/contracts-bedrock/forge-artifacts",
		EnvVars: op_service.PrefixEnvVar(EnvPrefix, "ARTIFACTS_DIR"),
	}
	foundryDirFlag = &cli.StringFlag{
		Name:    "foundry-dir",
		Value:   "packages/contracts-bedrock",
		EnvVars: op_service.PrefixEnvVar(EnvPrefix, "FOUNDRY_DIR"),
		Usage:   "Optional, for source-map info during genesis generation",
	}
	outDirFlag = &cli.StringFlag{
		Name:    "out-dir",
		Value:   ".interop-devnet",
		EnvVars: op_service.PrefixEnvVar(EnvPrefix, "OUT_DIR"),
	}
)

var InteropDevSetup = &cli.Command{
	Name:  "dev-setup",
	Usage: "Generate devnet genesis configs with one L1 and multiple L2s",
	Flags: cliapp.ProtectFlags(append([]cli.Flag{
		l1ChainIDFlag,
		l2ChainIDsFlag,
		timestampFlag,
		mnemonicFlag,
		artifactsDirFlag,
		foundryDirFlag,
		outDirFlag,
	}, oplog.CLIFlags(EnvPrefix)...)),
	Action: func(cliCtx *cli.Context) error {
		logCfg := oplog.ReadCLIConfig(cliCtx)
		logger := oplog.NewLogger(cliCtx.App.Writer, logCfg)

		recipe := &interopgen.InteropDevRecipe{
			L1ChainID:        cliCtx.Uint64(l1ChainIDFlag.Name),
			L2ChainIDs:       cliCtx.Uint64Slice(l2ChainIDsFlag.Name),
			GenesisTimestamp: cliCtx.Uint64(timestampFlag.Name),
		}
		if recipe.GenesisTimestamp == 0 {
			recipe.GenesisTimestamp = uint64(time.Now().Unix() + 5)
		}
		mnemonic := strings.TrimSpace(cliCtx.String(mnemonicFlag.Name))
		if mnemonic == devkeys.TestMnemonic {
			logger.Warn("Using default test mnemonic!")
		}
		keys, err := devkeys.NewMnemonicDevKeys(mnemonic)
		if err != nil {
			return fmt.Errorf("failed to setup dev keys from mnemonic: %w", err)
		}
		worldCfg, err := recipe.Build(keys)
		if err != nil {
			return fmt.Errorf("failed to build deploy configs from interop recipe: %w", err)
		}
		if err := worldCfg.Check(logger); err != nil {
			return fmt.Errorf("invalid deploy configs: %w", err)
		}
		artifactsDir := cliCtx.String(artifactsDirFlag.Name)
		af := foundry.OpenArtifactsDir(artifactsDir)
		var srcFs *foundry.SourceMapFS
		if cliCtx.IsSet(foundryDirFlag.Name) {
			srcDir := cliCtx.String(foundryDirFlag.Name)
			srcFs = foundry.NewSourceMapFS(os.DirFS(srcDir))
		}
		worldDeployment, worldOutput, err := interopgen.Deploy(logger, af, srcFs, worldCfg)
		if err != nil {
			return fmt.Errorf("failed to deploy interop dev setup: %w", err)
		}
		outDir := cliCtx.String(outDirFlag.Name)
		// Write deployments
		{
			deploymentsDir := filepath.Join(outDir, "deployments")
			l1Dir := filepath.Join(deploymentsDir, "l1")
			if err := writeJson(filepath.Join(l1Dir, "common.json"), worldDeployment.L1); err != nil {
				return fmt.Errorf("failed to write L1 deployment data: %w", err)
			}
			if err := writeJson(filepath.Join(l1Dir, "superchain.json"), worldDeployment.Superchain); err != nil {
				return fmt.Errorf("failed to write Superchain deployment data: %w", err)
			}
			l2sDir := filepath.Join(deploymentsDir, "l2")
			for id, dep := range worldDeployment.L2s {
				l2Dir := filepath.Join(l2sDir, id)
				if err := writeJson(filepath.Join(l2Dir, "addresses.json"), dep); err != nil {
					return fmt.Errorf("failed to write L2 %s deployment data: %w", id, err)
				}
			}
		}
		// write genesis
		{
			genesisDir := filepath.Join(outDir, "genesis")
			l1Dir := filepath.Join(genesisDir, "l1")
			if err := writeJson(filepath.Join(l1Dir, "genesis.json"), worldOutput.L1.Genesis); err != nil {
				return fmt.Errorf("failed to write L1 genesis data: %w", err)
			}
			l2sDir := filepath.Join(genesisDir, "l2")
			for id, dep := range worldOutput.L2s {
				l2Dir := filepath.Join(l2sDir, id)
				if err := writeJson(filepath.Join(l2Dir, "genesis.json"), dep.Genesis); err != nil {
					return fmt.Errorf("failed to write L2 %s genesis config: %w", id, err)
				}
				if err := writeJson(filepath.Join(l2Dir, "rollup.json"), dep.RollupCfg); err != nil {
					return fmt.Errorf("failed to write L2 %s rollup config: %w", id, err)
				}
			}
		}
		return nil
	},
}

func writeJson(path string, content any) error {
	outDir := filepath.Dir(path)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create dir %q: %w", outDir, err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(content); err != nil {
		return fmt.Errorf("failed to write JSON content: %w", err)
	}
	return nil
}

var InteropCmd = &cli.Command{
	Name:  "interop",
	Usage: "Experimental tools for OP-Stack interop networks.",
	Subcommands: cli.Commands{
		InteropDevSetup,
	},
}
