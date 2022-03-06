pragma solidity ^0.8.0;

contract Origin {
    address owner;

    function test() public{
        owner = msg.sender;
    }
}
