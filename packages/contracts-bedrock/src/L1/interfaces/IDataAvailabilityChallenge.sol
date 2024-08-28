// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { IOwnableUpgradeable } from "src/universal/interfaces/IOwnableUpgradeable.sol";
import { ISemver } from "src/universal/ISemver.sol";

enum ChallengeStatus {
    Uninitialized,
    Active,
    Resolved,
    Expired
}

enum CommitmentType {
    Keccak256
}

struct Challenge {
    address challenger;
    uint256 lockedBond;
    uint256 startBlock;
    uint256 resolvedBlock;
}

/// @title IDataAvailabilityChallenge
/// @notice Interface for the DataAvailabilityChallenge contract.
interface IDataAvailabilityChallenge is IOwnableUpgradeable, ISemver {
    event BalanceChanged(address account, uint256 balance);
    event ChallengeStatusChanged(
        uint256 indexed challengedBlockNumber, bytes challengedCommitment, ChallengeStatus status
    );
    event RequiredBondSizeChanged(uint256 challengeWindow);
    event ResolverRefundPercentageChanged(uint256 resolverRefundPercentage);

    receive() external payable;

    function balances(address _account) external view returns (uint256);
    function bondSize() external view returns (uint256);
    function challenge(uint256 _challengedBlockNumber, bytes memory _challengedCommitment) external payable;
    function challengeWindow() external view returns (uint256);
    function deposit() external payable;
    function fixedResolutionCost() external view returns (uint256);
    function getChallenge(
        uint256 _challengedBlockNumber,
        bytes memory _challengedCommitment
    )
        external
        view
        returns (Challenge memory);
    function getChallengeStatus(
        uint256 _challengedBlockNumber,
        bytes memory _challengedCommitment
    )
        external
        view
        returns (ChallengeStatus);
    function initialize(
        address _owner,
        uint256 _challengeWindow,
        uint256 _resolveWindow,
        uint256 _bondSize,
        uint256 _resolverRefundPercentage
    )
        external;
    function resolve(
        uint256 _challengedBlockNumber,
        bytes memory _challengedCommitment,
        bytes memory _resolveData
    )
        external;
    function resolveWindow() external view returns (uint256);
    function resolverRefundPercentage() external view returns (uint256);
    function setBondSize(uint256 _bondSize) external;
    function setResolverRefundPercentage(uint256 _resolverRefundPercentage) external;
    function unlockBond(uint256 _challengedBlockNumber, bytes memory _challengedCommitment) external;
    function validateCommitment(bytes memory _commitment) external pure;
    function variableResolutionCost() external view returns (uint256);
    function variableResolutionCostPrecision() external view returns (uint256);
    function withdraw() external;
}
