#!/usr/bin/env bash
set -euo pipefail

echoerr() {
  echo "$@" 1>&2
}

if [ -z "${DEPLOY_ETH_RPC_URL:-}" ]; then
  echoerr "Error: DEPLOY_ETH_RPC_URL is not set"
  exit 1
fi

if [ -z "${DEPLOY_PRIVATE_KEY:-}" ]; then
  echoerr "Error: DEPLOY_PRIVATE_KEY is not set"
  exit 1
fi

if [ -z "${DEPLOY_STATE_PATH:-}" ]; then
  echoerr "Error: DEPLOY_STATE_PATH is not set"
  exit 1
fi

cd "/workspace/optimism/packages/contracts-bedrock"

export ETH_RPC_URL="$DEPLOY_ETH_RPC_URL"

# if impl salt isn't set assign default
if [ -z "${DEPLOY_IMPL_SALT:-}" ]; then
  DEPLOY_IMPL_SALT=$(openssl rand -hex 32)
fi

echoerr "> Using impl salt: $DEPLOY_IMPL_SALT"

echoerr "> Checking CREATE2 deployer..."

codesize=$(cast codesize 0x4e59b44847b379578588920cA78FbF26c0B4956C)
if [ "$codesize" == "0" ]; then
  cast publish 0xf8a58085174876e800830186a08080b853604580600e600039806000f350fe7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf31ba02222222222222222222222222222222222222222222222222222222222222222a02222222222222222222222222222222222222222222222222222222222222222
  codesize=$(cast codesize 0x4e59b44847b379578588920cA78FbF26c0B4956C)
  if [ "$codesize" == "0" ]; then
    echoerr "> CREATE2 deployment failed."
    exit 1
  elif [ "$codesize" == "69" ]; then
    echoerr "> CREATE2 deployer successfully deployed."
  else
    echoerr "> CREATE2 deployer failed with unexpected exit code $?."
    exit 1
  fi
elif [ "$codesize" == "69" ]; then
  echoerr "> CREATE2 deployer is already deployed."
else
  echoerr "> CREATE2 deployer failed with unexpected exit code $?."
  exit 1
fi

verify_flag=""
if [ -n "${DEPLOY_VERIFY:-}" ]; then
  verify_flag="--verify"
fi

echoerr "> Extracting deploy config..."
cat "$DEPLOY_STATE_PATH" | jq -r '.deployConfig' > ./deploy-config/deploy-config.json
echoerr "> Done."

echoerr "> Deploying..."
DEPLOY_CONFIG_PATH=./deploy-config/deploy-config.json \
IMPL_SALT="$DEPLOY_IMPL_SALT" \
DEPLOYMENT_OUTFILE=./deployments/deployment.json \
DEPLOYMENT_CONTEXT="docker-deployer" \
  forge script scripts/deploy/Deploy.s.sol:Deploy \
    --private-key "$DEPLOY_PRIVATE_KEY" \
    --rpc-url "$DEPLOY_ETH_RPC_URL" \
    --broadcast \
    --gas-estimate-multiplier 200 \
    "$verify_flag"
echoerr "> Done."