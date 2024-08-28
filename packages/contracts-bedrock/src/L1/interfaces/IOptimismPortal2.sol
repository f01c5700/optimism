// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { IInitializable } from "src/universal/interfaces/IInitializable.sol";
import { ISemver } from "src/universal/ISemver.sol";
import { IResourceMetering } from "src/L1/interfaces/IResourceMetering.sol";
import { GameType, Timestamp } from "src/dispute/lib/Types.sol";
import { Types } from "src/libraries/Types.sol";

/// @title IOptimismPortal2
/// @notice Interface for the OptimismPortal2 contract.
interface IOptimismPortal2 is IInitializable, IResourceMetering, ISemver {
    event DisputeGameBlacklisted(address indexed disputeGame);
    event RespectedGameTypeSet(GameType indexed newGameType, Timestamp indexed updatedAt);
    event TransactionDeposited(address indexed from, address indexed to, uint256 indexed version, bytes opaqueData);
    event WithdrawalFinalized(bytes32 indexed withdrawalHash, bool success);
    event WithdrawalProven(bytes32 indexed withdrawalHash, address indexed from, address indexed to);
    event WithdrawalProvenExtension1(bytes32 indexed withdrawalHash, address indexed proofSubmitter);

    receive() external payable;

    function balance() external view returns (uint256);
    function blacklistDisputeGame(address _disputeGame) external;
    function checkWithdrawal(bytes32 _withdrawalHash, address _proofSubmitter) external view;
    function depositERC20Transaction(
        address _to,
        uint256 _mint,
        uint256 _value,
        uint64 _gasLimit,
        bool _isCreation,
        bytes memory _data
    )
        external;
    function depositTransaction(
        address _to,
        uint256 _value,
        uint64 _gasLimit,
        bool _isCreation,
        bytes memory _data
    )
        external
        payable;
    function disputeGameBlacklist(address) external view returns (bool);
    function disputeGameFactory() external view returns (address);
    function disputeGameFinalityDelaySeconds() external view returns (uint256);
    function donateETH() external payable;
    function finalizeWithdrawalTransaction(Types.WithdrawalTransaction memory _tx) external;
    function finalizeWithdrawalTransactionExternalProof(
        Types.WithdrawalTransaction memory _tx,
        address _proofSubmitter
    )
        external;
    function finalizedWithdrawals(bytes32) external view returns (bool);
    function guardian() external view returns (address);
    function initialize(
        address _disputeGameFactory,
        address _systemConfig,
        address _superchainConfig,
        GameType _initialRespectedGameType
    )
        external;
    function l2Sender() external view returns (address);
    function minimumGasLimit(uint64 _byteCount) external pure returns (uint64);
    function numProofSubmitters(bytes32 _withdrawalHash) external view returns (uint256);
    function paused() external view returns (bool);
    function proofMaturityDelaySeconds() external view returns (uint256);
    function proofSubmitters(bytes32, uint256) external view returns (address);
    function proveWithdrawalTransaction(
        Types.WithdrawalTransaction memory _tx,
        uint256 _disputeGameIndex,
        Types.OutputRootProof memory _outputRootProof,
        bytes[] memory _withdrawalProof
    )
        external;
    function provenWithdrawals(bytes32, address) external view returns (address disputeGameProxy_, uint64 timestamp_);
    function respectedGameType() external view returns (GameType);
    function respectedGameTypeUpdatedAt() external view returns (uint64);
    function setGasPayingToken(address _token, uint8 _decimals, bytes32 _name, bytes32 _symbol) external;
    function setRespectedGameType(GameType _gameType) external;
    function superchainConfig() external view returns (address);
    function systemConfig() external view returns (address);
    function version() external pure returns (string memory);
}
