// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { IInitializable } from "src/universal/interfaces/IInitializable.sol";

/// @title IERC721Bridge
/// @notice Interface for the ERC721Bridge contract.
interface IERC721Bridge is IInitializable {
    event ERC721BridgeFinalized(
        address indexed localToken,
        address indexed remoteToken,
        address indexed from,
        address to,
        uint256 tokenId,
        bytes extraData
    );
    event ERC721BridgeInitiated(
        address indexed localToken,
        address indexed remoteToken,
        address indexed from,
        address to,
        uint256 tokenId,
        bytes extraData
    );

    function MESSENGER() external view returns (address);
    function OTHER_BRIDGE() external view returns (address);
    function bridgeERC721(
        address _localToken,
        address _remoteToken,
        uint256 _tokenId,
        uint32 _minGasLimit,
        bytes memory _extraData
    )
        external;
    function bridgeERC721To(
        address _localToken,
        address _remoteToken,
        address _to,
        uint256 _tokenId,
        uint32 _minGasLimit,
        bytes memory _extraData
    )
        external;
    function messenger() external view returns (address);
    function otherBridge() external view returns (address);
    function paused() external view returns (bool);
}
