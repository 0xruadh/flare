// (c) 2021, Flare Networks Limited. All rights reserved.
// Please see the file LICENSE for licensing terms.

// SPDX-License-Identifier: MIT
pragma solidity 0.7.6;

contract TestFTSOBlockHeight {

    address public testerAddress;

    mapping(uint64 => uint64) public chainHeight;

    modifier isTesterAddress() {
        require(msg.sender == testerAddress, "msg.sender != testerAddress");
        _;
    }

    constructor() {
        testerAddress = msg.sender;
    }

    function setBlockHeight(uint64 chainId, uint64 ledgerNum) external isTesterAddress {
        require(ledgerNum > chainHeight[chainId], "ledgerNum <= chainHeight[chainId]");
        chainHeight[chainId] = ledgerNum;
    }
    
    function getBlockHeight(uint64 chainId) external view returns (uint64 ledgerNum) {
        return chainHeight[chainId];
    }
    
}
