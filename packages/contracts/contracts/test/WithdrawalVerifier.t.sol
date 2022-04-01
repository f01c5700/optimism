//SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

/* Testing utilities */
import { DSTest } from "../../lib/ds-test/src/test.sol";
import { Vm } from "../../lib/forge-std/src/Vm.sol";
import { L2OutputOracle_Initializer } from "./L2OutputOracle.t.sol";

/* Target contract dependencies */
import { L2OutputOracle } from "../L1/L2OutputOracle.sol";

/* Target contract */
import { WithdrawalVerifier } from "../L1/WithdrawalVerifier.sol";

contract WithdrawalVerifierTest is DSTest {
    // Utilities
    Vm vm = Vm(HEVM_ADDRESS);
    bytes32 nonZeroHash = keccak256(abi.encode("NON_ZERO"));

    // Dependencies
    L2OutputOracle oracle;

    // Oracle constructor arguments
    address sequencer = 0x000000000000000000000000000000000000AbBa;
    uint256 submissionInterval = 1800;
    uint256 l2BlockTime = 2;
    bytes32 genesisL2Output = keccak256(abi.encode(0));
    uint256 historicalTotalBlocks = 100;

    // Test target
    WithdrawalVerifier wv;

    // Target constructor arguments
    address withdrawalsPredeploy = 0x4200000000000000000000000000000000000015;

    // Cache of timestamps
    uint256 startingBlockTimestamp;
    uint256 appendedTimestamp;

    // By default the first block has timestamp zero, which will cause underflows in the tests,
    // so we jump ahead to the exact time that I wrote this line.
    uint256 initTime = 1648757197;

    // Generate an output that we can work with. We can use whatever values we want
    // except for the withdrawerRoot. This one was generated by running the withdrawor.spec.ts
    // test script against Geth.
    bytes32 version = bytes32(hex"00");
    bytes32 stateRoot = keccak256(abi.encode(1));
    bytes32 withdrawerRoot = 0xb8576230d94535779ec872748df80a094fcad002a8fc2b37c5b8fe250b384be6;
    bytes32 latestBlockhash = keccak256(abi.encode(2));
    WithdrawalVerifier.OutputRootProof outputRootProof;

    constructor() {
        // Move time forward so we have a non-zero starting timestamp
        vm.warp(initTime);

        // Deploy the L2OutputOracle and transfer owernship to the sequencer
        oracle = new L2OutputOracle(
            submissionInterval,
            l2BlockTime,
            genesisL2Output,
            historicalTotalBlocks,
            sequencer
        );
        startingBlockTimestamp = block.timestamp;

        wv = new WithdrawalVerifier(oracle, withdrawalsPredeploy);
    }

    function setUp() external {
        vm.warp(initTime);
        bytes32 outputRoot = keccak256(
            abi.encode(version, stateRoot, withdrawerRoot, latestBlockhash)
        );

        uint256 nextTimestamp = oracle.nextTimestamp();
        // Warp to 1 second after the timestamp we'll append
        vm.warp(nextTimestamp + 1);
        vm.prank(sequencer);
        oracle.appendL2Output(outputRoot, nextTimestamp, 0, 0);

        // cache the appendedTimestamp
        appendedTimestamp = nextTimestamp;
        outputRootProof = WithdrawalVerifier.OutputRootProof({
            timestamp: appendedTimestamp,
            version: version,
            stateRoot: stateRoot,
            withdrawerRoot: withdrawerRoot,
            latestBlockhash: latestBlockhash
        });
    }

    function test_verifyWithdrawal() external {
        // Warp to after the finality window
        vm.warp(appendedTimestamp + 7 days);
        wv.verifyWithdrawal(
            0, // nonce
            0xDe3829A23DF1479438622a08a116E8Eb3f620BB5, // sender
            0x1111111111111111111111111111111111111111, // target
            0, // value
            50_000, // gasLimit
            hex"111111111111111111111111111111111111111111111111111111111111111111111111111111111111", //data
            outputRootProof
        );
    }

    function test_cannotVerifyRecentWithdrawal() external {
        // This call should fail because the output root we're using was appended 1 second ago.
        vm.expectRevert("Too soon");
        wv.verifyWithdrawal(
            0, // nonce
            0xDe3829A23DF1479438622a08a116E8Eb3f620BB5, // sender
            0x1111111111111111111111111111111111111111, // target
            0, // value
            50_000, // gasLimit
            hex"111111111111111111111111111111111111111111111111111111111111111111111111111111111111", //data
            outputRootProof
        );
    }

    function test_cannotVerifyInvalidProof() external {
        // This call should fail because the output root we're using was appended 1 second ago.
        vm.warp(appendedTimestamp + 7 days);
        vm.expectRevert("Calculated output root does not match expected value");
        WithdrawalVerifier.OutputRootProof memory invalidOutpuRootProof = outputRootProof;
        invalidOutpuRootProof.latestBlockhash = 0;
        wv.verifyWithdrawal(
            0, // nonce
            0xDe3829A23DF1479438622a08a116E8Eb3f620BB5, // sender
            0x1111111111111111111111111111111111111111, // target
            0, // value
            50_000, // gasLimit
            hex"111111111111111111111111111111111111111111111111111111111111111111111111111111111111", //data
            invalidOutpuRootProof
        );
    }
}
