package pipeline

import (
	"context"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-chain-ops/deployer/opsm"
	"github.com/ethereum-optimism/optimism/op-chain-ops/deployer/state"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
)

const DefaultContractsBedrockRepo = "us-docker.pkg.dev/oplabs-tools-artifacts/images/contracts-bedrock"

func DeploySuperchain(ctx context.Context, env *Env, intent *state.Intent, st *state.State) error {
	lgr := env.Logger.New("stage", "deploy-superchain")

	if !shouldDeploySuperchain(intent, st) {
		lgr.Info("superchain deployment not needed")
		return nil
	}

	lgr.Info("deploying superchain")

	contractsRepo := intent.ContractsRepo
	if contractsRepo == "" {
		contractsRepo = DefaultContractsBedrockRepo
	}

	dso, err := opsm.DeploySuperchainDocker(
		ctx,
		opsm.DeploySuperchainOpts{
			ContractsImage: fmt.Sprintf("%s:%s", intent.ContractsRepo, intent.ContractsVersion),
			PrivateKey:     env.PrivateKey,
			Input: opsm.DeploySuperchainInput{
				Roles: opsm.DeploySuperchainRoles{
					ProxyAdminOwner:       intent.SuperchainRoles.ProxyAdminOwner,
					ProtocolVersionsOwner: intent.SuperchainRoles.ProtocolVersionsOwner,
					Guardian:              intent.SuperchainRoles.Guardian,
				},
				Paused:                     false,
				RequiredProtocolVersion:    rollup.OPStackSupport,
				RecommendedProtocolVersion: rollup.OPStackSupport,
			},
			L1RPCUrl: env.L1RPCUrl,
			Logger:   lgr,
		},
	)
	if err != nil {
		return fmt.Errorf("error deploying superchain: %w", err)
	}

	st.SuperchainDeployment = &state.SuperchainDeployment{
		ProxyAdminAddress:            dso.SuperchainProxyAdmin,
		SuperchainConfigProxyAddress: dso.SuperchainConfigProxy,
		ProtocolVersionsProxyAddress: dso.ProtocolVersionsProxy,
	}

	if err := env.WriteState(st); err != nil {
		return err
	}

	return nil
}

func shouldDeploySuperchain(intent *state.Intent, st *state.State) bool {
	if st.AppliedIntent == nil {
		return true
	}

	if st.SuperchainDeployment == nil {
		return true
	}

	return false
}
