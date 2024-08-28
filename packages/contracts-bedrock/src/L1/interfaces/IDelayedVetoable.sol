// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { ISemver } from "src/universal/ISemver.sol";

/// @title IDelayedVetoable
/// @notice Interface for the DelayedVetoable contract.
interface IDelayedVetoable is ISemver {
    event DelayActivated(uint256 delay);
    event Forwarded(bytes32 indexed callHash, bytes data);
    event Initiated(bytes32 indexed callHash, bytes data);
    event Vetoed(bytes32 indexed callHash, bytes data);

    fallback() external;

    function delay() external returns (uint256 delay_);
    function initiator() external returns (address initiator_);
    function queuedAt(bytes32 _callHash) external returns (uint256 queuedAt_);
    function target() external returns (address target_);
    function vetoer() external returns (address vetoer_);
}
