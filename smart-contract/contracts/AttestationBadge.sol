pragma solidity ^0.8.28;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {ERC721Enumerable} from "@openzeppelin/contracts/token/ERC721/extensions/ERC721Enumerable.sol";

contract AttestationBadge is ERC721Enumerable {
    error MinterRequired();
    error UnauthorizedMinter(address caller);
    error InvalidRecord();
    error BadgeAlreadyMinted(address recipient, uint256 recordId);
    error BadgeNonTransferable();

    address public immutable minter;
    uint256 private lastTokenId;
    mapping(address => mapping(uint256 => uint256)) public badgeOf;
    mapping(uint256 => uint256) public recordOf;

    constructor(address minter_) ERC721("Attestation Badge", "ATTEST") {
        if (minter_ == address(0)) revert MinterRequired();
        minter = minter_;
    }

    function mint(address to, uint256 recordId) external returns (uint256 tokenId) {
        if (msg.sender != minter) revert UnauthorizedMinter(msg.sender);
        if (recordId == 0) revert InvalidRecord();
        if (badgeOf[to][recordId] != 0) revert BadgeAlreadyMinted(to, recordId);

        tokenId = ++lastTokenId;
        badgeOf[to][recordId] = tokenId;
        recordOf[tokenId] = recordId;
        _mint(to, tokenId);
    }

    function _update(address to, uint256 tokenId, address auth) internal override returns (address) {
        if (_ownerOf(tokenId) != address(0)) revert BadgeNonTransferable();
        return super._update(to, tokenId, auth);
    }
}
