// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { IInitializable } from "src/universal/interfaces/IInitializable.sol";

/// @title IResourceMetering
/// @notice Interface for the ResourceMetering contract.
interface IResourceMetering is IInitializable {
    function params() external view returns (uint128 prevBaseFee_, uint64 prevBoughtGas_, uint64 prevBlockNum_);
}
