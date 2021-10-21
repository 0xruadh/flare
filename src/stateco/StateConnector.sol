// (c) 2021, Flare Networks Limited. All rights reserved.
// Please see the file LICENSE for licensing terms.

// SPDX-License-Identifier: MIT
pragma solidity 0.8.9;

contract StateConnector {

//====================================================================
// Data Structures
//====================================================================

    struct HashExists {
        bool        exists;
        bool        proven;
        uint256     revealTime;
        uint64      indexValue;
        bytes32     hashValue;
    }

    uint256 private constant TWO_PHASE_COMMIT_LOWER_BOUND = 30;
    uint256 private constant TWO_PHASE_COMMIT_UPPER_BOUND = 1 days;
    address private constant GENESIS_COINBASE = address(0x0100000000000000000000000000000000000000);

    // Finalised payment hashes
    mapping(bytes32 => HashExists) private payments;

//====================================================================
// Events
//====================================================================

    event PaymentSet(uint64 chainId, uint64 ledger, bytes32 txId, uint16 utxo, bytes32 paymentHash);

//====================================================================
// Constructor
//====================================================================

    constructor() {
    }

//====================================================================
// Functions
//====================================================================  

    function setPaymentFinality(
        uint64 chainId,
        uint64 ledger,
        bytes32 txId,
        uint16 utxo,
        bytes32 paymentHash
    ) external returns (
        uint256 _instructions,
        bytes32 _txId,
        bytes32 _paymentHash
    ) {
        require(ledger > 0, "ledger == 0");
        require(txId > 0x0, "txId == 0x0");
        require(paymentHash > 0x0, "paymentHash == 0x0");
        require(block.coinbase == msg.sender || block.coinbase == GENESIS_COINBASE, "invalid block.coinbase value");

        bytes32 location = keccak256(abi.encodePacked(
            keccak256(abi.encodePacked("FlareStateConnector_LOCATION")),
            keccak256(abi.encodePacked(chainId)),
            keccak256(abi.encodePacked(ledger)),
            keccak256(abi.encodePacked(txId)),
            keccak256(abi.encodePacked(utxo))
        ));

        bytes32 finalisedPaymentLocation = keccak256(abi.encodePacked(
            keccak256(abi.encodePacked("FlareStateConnector_FINALISED")),
            location
        ));
        require(!payments[finalisedPaymentLocation].proven, "payment already proven");

        bool initialCommit;
        bytes32 proposedPaymentLocation = keccak256(abi.encodePacked(
            keccak256(abi.encodePacked("FlareStateConnector_PROPOSED")),
            keccak256(abi.encodePacked(msg.sender)),
            location
        ));
        if (payments[proposedPaymentLocation].exists) {
            require(block.timestamp >= payments[proposedPaymentLocation].revealTime, 
                "block.timestamp < payments[proposedPaymentLocation].revealTime");
            require(payments[proposedPaymentLocation].revealTime + TWO_PHASE_COMMIT_UPPER_BOUND > block.timestamp,
                "reveal is too late");
            require(payments[proposedPaymentLocation].indexValue == ledger, 
                "invalid ledger");
            require(payments[proposedPaymentLocation].hashValue == paymentHash, 
                "invalid paymentHash");
        } else if (block.coinbase != msg.sender && block.coinbase == GENESIS_COINBASE) {
            initialCommit = true;
        }

        if (block.coinbase == msg.sender && block.coinbase != GENESIS_COINBASE) {
            if (!payments[proposedPaymentLocation].exists) {
                payments[proposedPaymentLocation] = HashExists(
                    true,
                    false,
                    block.timestamp + TWO_PHASE_COMMIT_LOWER_BOUND,
                    ledger,
                    paymentHash
                );
            } else {
                payments[finalisedPaymentLocation] = HashExists(
                    true,
                    true,
                    block.timestamp,
                    ledger,
                    paymentHash
                );
                emit PaymentSet(chainId, ledger, txId, utxo, paymentHash);
            }
        }

        return (
            uint256(initialCommit?1:0)<<192 | uint256(chainId)<<128 | uint256(ledger)<<64 | uint256(utxo),
            txId,
            paymentHash
        );
    }

    function getPaymentFinality(
        uint64 chainId,
        uint64 ledger,
        bytes32 txId,
        uint16 utxo,
        bytes32 destinationHash,
        bytes32 dataHash,
        uint256 amount
    ) external view returns (
        bool _proven
    ) {
        bytes32 location = keccak256(abi.encodePacked(
            keccak256(abi.encodePacked("FlareStateConnector_LOCATION")),
            keccak256(abi.encodePacked(chainId)),
            keccak256(abi.encodePacked(ledger)),
            keccak256(abi.encodePacked(txId)),
            keccak256(abi.encodePacked(utxo))
        ));
        bytes32 finalisedPaymentLocation = keccak256(abi.encodePacked(
            keccak256(abi.encodePacked("FlareStateConnector_FINALISED")),
            location
        ));
        bytes32 paymentHash = keccak256(abi.encodePacked(
            keccak256(abi.encodePacked("FlareStateConnector_PAYMENTHASH")),
            destinationHash,
            dataHash,
            keccak256(abi.encodePacked(amount))
        ));
        require(payments[finalisedPaymentLocation].exists, "payment does not exist");
        require(payments[finalisedPaymentLocation].proven, "payment is not yet proven");
        require(payments[finalisedPaymentLocation].indexValue == ledger, "invalid ledger value");
        require(payments[finalisedPaymentLocation].hashValue == paymentHash, "invalid paymentHash");

        return (payments[finalisedPaymentLocation].proven);
    }

}
