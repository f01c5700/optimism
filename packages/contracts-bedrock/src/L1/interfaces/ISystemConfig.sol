// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { IGasToken } from "src/libraries/GasPayingToken.sol";
import { IOwnableUpgradeable } from "src/universal/interfaces/IOwnableUpgradeable.sol";
import { ISemver } from "src/universal/ISemver.sol";

/// @title ISystemConfig
/// @notice Interface for the SystemConfig contract.
interface ISystemConfig is IOwnableUpgradeable, IGasToken, ISemver {
    enum UpdateType {
        BATCHER,
        GAS_CONFIG,
        GAS_LIMIT,
        UNSAFE_BLOCK_SIGNER
    }

    struct Addresses {
        address l1CrossDomainMessenger;
        address l1ERC721Bridge;
        address l1StandardBridge;
        address disputeGameFactory;
        address optimismPortal;
        address optimismMintableERC20Factory;
        address gasPayingToken;
    }

    struct ResourceConfig {
        uint32 maxResourceLimit;
        uint8 elasticityMultiplier;
        uint8 baseFeeMaxChangeDenominator;
        uint32 minimumBaseFee;
        uint32 systemTxMaxGas;
        uint128 maximumBaseFee;
    }

    event ConfigUpdate(uint256 indexed version, UpdateType indexed updateType, bytes data);

    function BATCH_INBOX_SLOT() external view returns (bytes32);
    function DISPUTE_GAME_FACTORY_SLOT() external view returns (bytes32);
    function L1_CROSS_DOMAIN_MESSENGER_SLOT() external view returns (bytes32);
    function L1_ERC_721_BRIDGE_SLOT() external view returns (bytes32);
    function L1_STANDARD_BRIDGE_SLOT() external view returns (bytes32);
    function OPTIMISM_MINTABLE_ERC20_FACTORY_SLOT() external view returns (bytes32);
    function OPTIMISM_PORTAL_SLOT() external view returns (bytes32);
    function START_BLOCK_SLOT() external view returns (bytes32);
    function UNSAFE_BLOCK_SIGNER_SLOT() external view returns (bytes32);
    function VERSION() external view returns (uint256);
    function basefeeScalar() external view returns (uint32);
    function batchInbox() external view returns (address addr_);
    function batcherHash() external view returns (bytes32);
    function blobbasefeeScalar() external view returns (uint32);
    function disputeGameFactory() external view returns (address addr_);
    function gasLimit() external view returns (uint64);
    function initialize(
        address _owner,
        uint32 _basefeeScalar,
        uint32 _blobbasefeeScalar,
        bytes32 _batcherHash,
        uint64 _gasLimit,
        address _unsafeBlockSigner,
        ResourceConfig memory _config,
        address _batchInbox,
        Addresses memory _addresses
    )
        external;
    function l1CrossDomainMessenger() external view returns (address addr_);
    function l1ERC721Bridge() external view returns (address addr_);
    function l1StandardBridge() external view returns (address addr_);
    function maximumGasLimit() external pure returns (uint64);
    function minimumGasLimit() external view returns (uint64);
    function optimismMintableERC20Factory() external view returns (address addr_);
    function optimismPortal() external view returns (address addr_);
    function overhead() external view returns (uint256);
    function resourceConfig() external view returns (ResourceConfig memory);
    function scalar() external view returns (uint256);
    function setBatcherHash(bytes32 _batcherHash) external;
    function setGasConfig(uint256 _overhead, uint256 _scalar) external;
    function setGasConfigEcotone(uint32 _basefeeScalar, uint32 _blobbasefeeScalar) external;
    function setGasLimit(uint64 _gasLimit) external;
    function setUnsafeBlockSigner(address _unsafeBlockSigner) external;
    function startBlock() external view returns (uint256 startBlock_);
    function unsafeBlockSigner() external view returns (address addr_);
}
