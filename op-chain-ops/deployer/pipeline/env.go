package pipeline

import (
	"context"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-chain-ops/deployer/state"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"path"
)

type Env struct {
	Workdir    string
	L1Client   *ethclient.Client
	L1RPCUrl   string
	PrivateKey string
	Logger     log.Logger
}

func (e *Env) ReadIntent() (*state.Intent, error) {
	intentPath := path.Join(e.Workdir, "intent.toml")
	var intent state.Intent
	if err := state.ReadTOMLFile(intentPath, &intent); err != nil {
		return nil, fmt.Errorf("failed to read intent file: %w", err)
	}
	return &intent, nil
}

func (e *Env) ReadState() (*state.State, error) {
	statePath := path.Join(e.Workdir, "state.json")
	var st state.State
	if err := state.ReadJSONFile(statePath, &st); err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}
	return &st, nil
}

func (e *Env) WriteState(st *state.State) error {
	statePath := path.Join(e.Workdir, "state.json")
	if err := state.WriteJSONFile(statePath, st); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	return nil
}

type Stage func(ctx context.Context, env *Env, intent *state.Intent, state2 *state.State) error
