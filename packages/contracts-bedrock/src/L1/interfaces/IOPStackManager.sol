// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { ISemver } from "src/universal/ISemver.sol";

/// @title IOPStackManager
/// @notice Interface for the OPStackManager contract.
interface IOPStackManager is ISemver {
    struct Roles {
        address proxyAdminOwner;
        address systemConfigOwner;
        address batcher;
        address unsafeBlockSigner;
        address proposer;
        address challenger;
    }

    event Deployed(uint256 indexed l2ChainId, address indexed systemConfig);

    function deploy(
        uint256 _l2ChainId,
        uint32 _basefeeScalar,
        uint32 _blobBasefeeScalar,
        Roles memory _roles
    )
        external
        view
        returns (address systemConfig_);
}
