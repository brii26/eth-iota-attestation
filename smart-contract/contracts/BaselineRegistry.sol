pragma solidity ^0.8.28;

contract BaselineRegistry {
    error WriterRequired();
    error UnauthorizedWriter(address caller);

    address public immutable writer;
    mapping(bytes32 => bool) private fingerprints;

    event HashesRecorded(bytes32 indexed key);

    constructor(address writer_) {
        if (writer_ == address(0)) revert WriterRequired();
        writer = writer_;
    }

    function addHashes(bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) external {
        if (msg.sender != writer) revert UnauthorizedWriter(msg.sender);
        bytes32 key = keccak256(abi.encode(artifactHash, sbomHash, scanHash));
        fingerprints[key] = true;
        emit HashesRecorded(key);
    }

    function hasHashes(bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) external view returns (bool) {
        return fingerprints[keccak256(abi.encode(artifactHash, sbomHash, scanHash))];
    }
}
