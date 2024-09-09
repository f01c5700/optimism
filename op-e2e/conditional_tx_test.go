package op_e2e

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/stretchr/testify/require"
)

const (
	sendTxCondMethodName = "eth_sendRawTransactionConditional"
)

func TestSendRawTransactionConditionalDisabled(t *testing.T) {
	InitParallel(t)
	cfg := DefaultSystemConfig(t)
	cfg.GethOptions[RoleSeq] = []geth.GethOption{func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
		ethCfg.RollupSequencerEnableTxConditional = false
		return nil
	}}

	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")

	err = sys.NodeClient(RoleSeq).Client().Call(nil, sendTxCondMethodName)
	require.Error(t, err)

	// method not found json error
	require.Equal(t, -32601, err.(*rpc.JsonError).Code)
}

func TestSendRawTransactionConditionalDisabledWhenSequencerHTTPSet(t *testing.T) {
	InitParallel(t)
	cfg := DefaultSystemConfig(t)
	cfg.GethOptions[RoleSeq] = []geth.GethOption{func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
		ethCfg.RollupSequencerHTTP = "http://localhost:8545"
		ethCfg.RollupSequencerEnableTxConditional = true
		return nil
	}}

	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")

	err = sys.NodeClient(RoleSeq).Client().Call(nil, sendTxCondMethodName)
	require.Error(t, err)

	// method not found json error
	require.Equal(t, -32601, err.(*rpc.JsonError).Code)
}

func TestSendRawTransactionConditionalEnabled(t *testing.T) {
	InitParallel(t)
	cfg := DefaultSystemConfig(t)
	cfg.GethOptions[RoleSeq] = []geth.GethOption{func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
		ethCfg.RollupSequencerEnableTxConditional = true
		return nil
	}}

	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")

	// wait for a couple l2 blocks to be created as conditionals are checked against older state
	l2Client := sys.NodeClient(RoleSeq)
	require.NoError(t, wait.ForBlock(context.Background(), l2Client, 5))

	gasLimit := uint64(21000) // Gas limit for a standard ETH transfer
	gasPrice, err := l2Client.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	from, to := cfg.Secrets.Addresses().Alice, cfg.Secrets.Addresses().Bob
	nonce, err := l2Client.PendingNonceAt(context.Background(), from)
	require.NoError(t, err)

	tx := types.NewTransaction(nonce, to, big.NewInt(params.Ether), gasLimit, gasPrice, nil)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(cfg.L2ChainIDBig()), cfg.Secrets.Alice)
	require.NoError(t, err)

	// send a sample tx with a conditional that will pass
	txBytes, err := rlp.EncodeToBytes(signedTx)
	require.NoError(t, err)
	require.NoError(t, l2Client.Client().Call(nil, sendTxCondMethodName, hexutil.Encode(txBytes), &types.TransactionConditional{BlockNumberMin: big.NewInt(0)}))
	_, err = wait.ForReceiptOK(context.Background(), l2Client, signedTx.Hash())
	require.NoError(t, err)
}

func TestSendRawTransactionConditionalE2E(t *testing.T) {
	InitParallel(t)
	cfg := DefaultSystemConfig(t)
	cfg.GethOptions[RoleSeq] = []geth.GethOption{func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
		ethCfg.RollupSequencerEnableTxConditional = true
		ethCfg.Miner.RollupTransactionConditionalBurstRate = 5000 // not parsed from default CLI values so explicily set
		return nil
	}}

	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")

	// wait for a couple l2 blocks to be created as conditionals are checked against older state
	l2Client := sys.NodeClient(RoleSeq)
	require.NoError(t, wait.ForBlock(context.Background(), l2Client, 5))

	mkTx := func() *types.Transaction {
		gasLimit := uint64(21000) // Gas limit for a standard ETH transfer
		gasPrice, err := l2Client.SuggestGasPrice(context.Background())
		require.NoError(t, err)

		from, to := cfg.Secrets.Addresses().Alice, cfg.Secrets.Addresses().Bob
		nonce, err := l2Client.PendingNonceAt(context.Background(), from)
		require.NoError(t, err)

		tx := types.NewTransaction(nonce, to, big.NewInt(params.Ether), gasLimit, gasPrice, nil)
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(cfg.L2ChainIDBig()), cfg.Secrets.Alice)
		require.NoError(t, err)

		return signedTx
	}

	uint64Ptr := func(num uint64) *uint64 { return &num }
	rpcClient := l2Client.Client()

	// send a tx with a conditional that will fail
	tx := mkTx()
	txBytes, err := rlp.EncodeToBytes(tx)
	require.NoError(t, err)
	err = rpcClient.Call(nil, sendTxCondMethodName, hexutil.Encode(txBytes), &types.TransactionConditional{TimestampMax: uint64Ptr(0)})
	require.Error(t, err)
	require.Equal(t, params.TransactionConditionalRejectedErrCode, err.(*rpc.JsonError).Code)

	// but works as a regular transaction
	require.NoError(t, l2Client.SendTransaction(context.Background(), tx))
	_, err = wait.ForReceiptOK(context.Background(), l2Client, tx.Hash())
	require.NoError(t, err)

	// send tx with conditional that will pass
	tx = mkTx()
	txBytes, err = rlp.EncodeToBytes(tx)
	require.NoError(t, err)

	var hash common.Hash
	require.NoError(t, rpcClient.Call(&hash, sendTxCondMethodName, hexutil.Encode(txBytes), &types.TransactionConditional{TimestampMin: uint64Ptr(1)}))
	require.Equal(t, tx.Hash(), hash)
	_, err = wait.ForReceiptOK(context.Background(), l2Client, tx.Hash())
	require.NoError(t, err)
}
