package state

import (
	"github.com/ethereum/go-ethereum/common"
)

// State contains the data needed to recreate the deployment
// as it progresses and once it is fully applied.
type State struct {
	// Version versions the state so we can update it later.
	Version int `json:"version"`

	// AppliedIntent contains the chain intent that was last
	// successfully applied. It is diffed against new intent
	// in order to determine what deployment steps to take.
	// This field is nil for new deployments.
	AppliedIntent *Intent `json:"appliedIntent"`

	// SuperchainDeployment contains the addresses of the Superchain
	// deployment. It only contains the proxies because the implementations
	// can be looked up on chain.
	SuperchainDeployment *SuperchainDeployment `json:"superchainDeployment"`
}

type SuperchainDeployment struct {
	ProxyAdminAddress            common.Address `json:"proxyAdminAddress"`
	SuperchainConfigProxyAddress common.Address `json:"superchainConfigAddress"`
	ProtocolVersionsProxyAddress common.Address `json:"protocolVersionsProxyAddress"`
}
