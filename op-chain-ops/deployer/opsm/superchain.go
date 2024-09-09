package opsm

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/optimism/op-chain-ops/deployer/docker"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"os"
)

type DeploySuperchainInput struct {
	Roles                      DeploySuperchainRoles  `toml:"roles"`
	Paused                     bool                   `toml:"paused"`
	RequiredProtocolVersion    params.ProtocolVersion `toml:"requiredProtocolVersion"`
	RecommendedProtocolVersion params.ProtocolVersion `toml:"recommendedProtocolVersion"`
}

type DeploySuperchainRoles struct {
	ProxyAdminOwner       common.Address `toml:"proxyAdminOwner"`
	ProtocolVersionsOwner common.Address `toml:"protocolVersionsOwner"`
	Guardian              common.Address `toml:"guardian"`
}

type DeploySuperchainOutput struct {
	SuperchainProxyAdmin  common.Address `toml:"superchainProxyAdmin"`
	SuperchainConfigImpl  common.Address `toml:"superchainConfigImpl"`
	SuperchainConfigProxy common.Address `toml:"superchainConfigProxy"`
	ProtocolVersionsImpl  common.Address `toml:"protocolVersionsImpl"`
	ProtocolVersionsProxy common.Address `toml:"protocolVersionsProxy"`
}

type DeploySuperchainOpts struct {
	ContractsImage string
	PrivateKey     string
	Input          DeploySuperchainInput
	L1RPCUrl       string
	Logger         log.Logger
}

func DeploySuperchainDocker(ctx context.Context, opts DeploySuperchainOpts) (DeploySuperchainOutput, error) {
	var dso DeploySuperchainOutput
	dsiInfile, err := os.CreateTemp("", "dsi-*.toml")
	if err != nil {
		return dso, fmt.Errorf("failed to create temp file for DSI: %w", err)
	}
	defer os.Remove(dsiInfile.Name())

	if err := toml.NewEncoder(dsiInfile).Encode(opts.Input); err != nil {
		return dso, fmt.Errorf("failed to encode DSI: %w", err)
	}

	containerDSIPath := "/opt/optimism/packages/contracts-bedrock/deployments/dsi.toml"
	containerDSOPath := "/opt/optimism/packages/contracts-bedrock/deployments/dso.toml"
	deployCmd, err := docker.NewCommand(
		opts.Logger,
		"contracts-bedrock:latest",
		docker.WithCmd(
			"forge",
			"script",
			"scripts/DeploySuperchain.s.sol:DeploySuperchain",
			"--private-key",
			opts.PrivateKey,
			"--rpc-url",
			opts.L1RPCUrl,
			"--broadcast",
			"--sig",
			"run(string, string)",
			containerDSIPath,
			containerDSOPath,
		),
		docker.WithMount(dsiInfile.Name(), containerDSIPath),
		docker.WithImagePlatform(docker.LinuxAMD64Platform),
	)
	if err != nil {
		return dso, fmt.Errorf("failed to create deploy command: %w", err)
	}

	if err := deployCmd.Run(ctx); err != nil {
		return dso, fmt.Errorf("failed to perform deployment: %w", err)
	}

	containerDSOData, err := deployCmd.ReadFile(ctx, containerDSOPath)
	if err != nil {
		return dso, fmt.Errorf("failed to read DSO: %w", err)
	}

	if err := toml.Unmarshal(containerDSOData, &dso); err != nil {
		return dso, fmt.Errorf("failed to unmarshal DSO: %w", err)
	}

	return dso, nil
}
