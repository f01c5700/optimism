package deployer

import (
	"context"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-chain-ops/deployer/pipeline"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
)

type ApplyConfig struct {
	L1RPCUrl   string
	Workdir    string
	PrivateKey string
	Logger     log.Logger
}

func (a *ApplyConfig) Check() error {
	if a.L1RPCUrl == "" {
		return fmt.Errorf("l1RPCUrl must be specified")
	}

	if a.Workdir == "" {
		return fmt.Errorf("workdir must be specified")
	}

	if a.PrivateKey == "" {
		return fmt.Errorf("private key must be specified")
	}

	if a.Logger == nil {
		return fmt.Errorf("logger must be specified")
	}

	return nil
}

func ApplyCLI() func(cliCtx *cli.Context) error {
	return func(cliCtx *cli.Context) error {
		logCfg := oplog.ReadCLIConfig(cliCtx)
		l := oplog.NewLogger(oplog.AppOut(cliCtx), logCfg)
		oplog.SetGlobalLogHandler(l.Handler())

		l1RPCUrl := cliCtx.String(L1RPCURLFlagName)
		workdir := cliCtx.String(WorkdirFlagName)
		privateKey := cliCtx.String(PrivateKeyFlagName)

		ctx, cancel := context.WithCancel(cliCtx.Context)
		defer cancel()

		errCh := make(chan error, 1)
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)

		go func() {
			err := Apply(ctx, ApplyConfig{
				L1RPCUrl:   l1RPCUrl,
				Workdir:    workdir,
				PrivateKey: privateKey,
				Logger:     l,
			})
			errCh <- err
		}()

		select {
		case err := <-errCh:
			cancel()
			return err
		case <-sigs:
			cancel()
			<-errCh
			return nil
		}
	}
}

func Apply(ctx context.Context, cfg ApplyConfig) error {
	if err := cfg.Check(); err != nil {
		return fmt.Errorf("invalid config for apply: %w", err)
	}

	l1Client, err := ethclient.Dial(cfg.L1RPCUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to L1 RPC: %w", err)
	}

	env := &pipeline.Env{
		Workdir:    cfg.Workdir,
		L1RPCUrl:   cfg.L1RPCUrl,
		L1Client:   l1Client,
		PrivateKey: cfg.PrivateKey,
		Logger:     cfg.Logger,
	}

	intent, err := env.ReadIntent()
	if err != nil {
		return err
	}

	if err := intent.Check(); err != nil {
		return fmt.Errorf("invalid intent: %w", err)
	}

	st, err := env.ReadState()
	if err != nil {
		return err
	}

	pline := []struct {
		name  string
		stage pipeline.Stage
	}{
		{"init", pipeline.Init},
		{"deploy-superchain", pipeline.DeploySuperchain},
	}
	for _, stage := range pline {
		if err := stage.stage(ctx, env, intent, st); err != nil {
			return fmt.Errorf("error in pipeline stage: %w", err)
		}
	}

	st.AppliedIntent = intent
	if err := env.WriteState(st); err != nil {
		return err
	}

	return nil
}
