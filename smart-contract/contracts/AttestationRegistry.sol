// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @notice Append-only attestations and oracle-fulfilled verification requests.
/// @dev Hashes describe original file bytes. The registry does not verify file contents.
contract AttestationRegistry {
    uint256 public constant MAX_APP_BYTES = 214;
    uint256 public constant MAX_VERSION_BYTES = 128;
    uint256 public constant MAX_SCAN_ID_BYTES = 128;
    uint256 public constant MAX_PAGE_SIZE = 100;

    struct Record {
        string app;
        string version;
        string scanId;
        bytes32 artifactHash;
        bytes32 sbomHash;
        bytes32 scanHash;
        address signer;
        uint256 timestamp;
    }

    enum IntegrityStatus {
        Pending,
        Valid,
        ArtifactMismatch,
        SbomMismatch,
        ScanMismatch,
        UntrustedSigner,
        RecordNotFound,
        FileUnavailable
    }

    struct VerificationResult {
        IntegrityStatus status;
        bool stale;
    }

    struct VerificationRequest {
        uint256 recordId;
        address requester;
        uint256 requestedAt;
        bool fulfilled;
        VerificationResult result;
        address fulfilledBy;
        uint256 fulfilledAt;
    }

    address public immutable admin;
    uint256 public recordCount;
    uint256 public verificationRequestCount;

    mapping(uint256 => Record) private records;
    mapping(bytes32 => uint256[]) private recordIdsByAppVersion;
    mapping(bytes32 => bool) private usedScanIds;
    mapping(address => bool) private authorizedOracles;
    mapping(uint256 => VerificationRequest) private verificationRequests;

    error UnauthorizedAdmin(address caller);
    error UnauthorizedOracle(address caller);
    error ZeroAddress();
    error OracleAlreadyAuthorized(address oracle);
    error OracleNotAuthorized(address oracle);
    error EmptyField(string field);
    error StringTooLong(string field, uint256 length, uint256 maximum);
    error EmptyHash(string field);
    error ScanIdAlreadyUsed(string scanId);
    error RecordNotFound(uint256 recordId);
    error NoRecordsForAppVersion(string app, string version);
    error InvalidPageLimit(uint256 limit);
    error VerificationRequestNotFound(uint256 requestId);
    error VerificationAlreadyFulfilled(uint256 requestId);
    error InvalidVerificationResult();

    event OracleAuthorizationChanged(address indexed oracle, bool authorized);
    event RecordAdded(
        uint256 indexed recordId,
        bytes32 indexed appVersionKey,
        address indexed signer,
        string app,
        string version,
        string scanId
    );
    event VerificationRequested(
        uint256 indexed requestId,
        uint256 indexed recordId,
        address indexed requester
    );
    event VerificationFulfilled(
        uint256 indexed requestId,
        uint256 indexed recordId,
        address indexed oracle,
        IntegrityStatus status,
        bool stale
    );

    constructor() {
        admin = msg.sender;
    }

    modifier onlyAdmin() {
        if (msg.sender != admin) revert UnauthorizedAdmin(msg.sender);
        _;
    }

    modifier onlyOracle() {
        if (!authorizedOracles[msg.sender]) revert UnauthorizedOracle(msg.sender);
        _;
    }

    function authorizeOracle(address oracle) external onlyAdmin {
        if (oracle == address(0)) revert ZeroAddress();
        if (authorizedOracles[oracle]) revert OracleAlreadyAuthorized(oracle);
        authorizedOracles[oracle] = true;
        emit OracleAuthorizationChanged(oracle, true);
    }

    function revokeOracle(address oracle) external onlyAdmin {
        if (oracle == address(0)) revert ZeroAddress();
        if (!authorizedOracles[oracle]) revert OracleNotAuthorized(oracle);
        authorizedOracles[oracle] = false;
        emit OracleAuthorizationChanged(oracle, false);
    }

    function isAuthorizedOracle(address oracle) external view returns (bool) {
        return authorizedOracles[oracle];
    }

    /// @notice Stores one scan. IDs start at one and follow recording order.
    function addRecord(
        string calldata app,
        string calldata version,
        string calldata scanId,
        bytes32 artifactHash,
        bytes32 sbomHash,
        bytes32 scanHash
    ) external onlyOracle returns (uint256 recordId) {
        _validateString(app, "app", MAX_APP_BYTES);
        _validateString(version, "version", MAX_VERSION_BYTES);
        _validateString(scanId, "scanId", MAX_SCAN_ID_BYTES);
        if (artifactHash == bytes32(0)) revert EmptyHash("artifactHash");
        if (sbomHash == bytes32(0)) revert EmptyHash("sbomHash");
        if (scanHash == bytes32(0)) revert EmptyHash("scanHash");

        bytes32 scanIdKey = keccak256(bytes(scanId));
        if (usedScanIds[scanIdKey]) revert ScanIdAlreadyUsed(scanId);
        bytes32 appVersionKey = _appVersionKey(app, version);

        recordId = ++recordCount;
        records[recordId] = Record({
            app: app,
            version: version,
            scanId: scanId,
            artifactHash: artifactHash,
            sbomHash: sbomHash,
            scanHash: scanHash,
            signer: msg.sender,
            timestamp: block.timestamp
        });
        usedScanIds[scanIdKey] = true;
        recordIdsByAppVersion[appVersionKey].push(recordId);

        emit RecordAdded(recordId, appVersionKey, msg.sender, app, version, scanId);
    }

    function getRecord(uint256 recordId) external view returns (Record memory) {
        _requireRecord(recordId);
        return records[recordId];
    }

    function getLatestRecord(
        string calldata app,
        string calldata version
    ) external view returns (Record memory) {
        uint256[] storage ids = recordIdsByAppVersion[_appVersionKey(app, version)];
        if (ids.length == 0) revert NoRecordsForAppVersion(app, version);
        return records[ids[ids.length - 1]];
    }

    /// @notice Returns IDs oldest first. An offset beyond the history returns an empty page.
    function getRecordIds(
        string calldata app,
        string calldata version,
        uint256 offset,
        uint256 limit
    ) external view returns (uint256[] memory page) {
        if (limit == 0 || limit > MAX_PAGE_SIZE) revert InvalidPageLimit(limit);
        uint256[] storage ids = recordIdsByAppVersion[_appVersionKey(app, version)];
        if (offset >= ids.length) return new uint256[](0);

        uint256 count = ids.length - offset;
        if (count > limit) count = limit;
        page = new uint256[](count);
        for (uint256 i = 0; i < count; ++i) {
            page[i] = ids[offset + i];
        }
    }

    function isScanIdUsed(string calldata scanId) external view returns (bool) {
        return usedScanIds[keccak256(bytes(scanId))];
    }

    function requestVerification(uint256 recordId) external returns (uint256 requestId) {
        _requireRecord(recordId);
        requestId = ++verificationRequestCount;
        VerificationRequest storage request = verificationRequests[requestId];
        request.recordId = recordId;
        request.requester = msg.sender;
        request.requestedAt = block.timestamp;
        emit VerificationRequested(requestId, recordId, msg.sender);
    }

    /// @notice Accepts one result from a currently authorized oracle.
    /// @dev Freshness is separate from integrity. Only a valid result can be marked stale.
    function fulfillVerification(
        uint256 requestId,
        VerificationResult calldata result
    ) external onlyOracle {
        _requireVerification(requestId);
        VerificationRequest storage request = verificationRequests[requestId];
        if (request.fulfilled) revert VerificationAlreadyFulfilled(requestId);
        if (
            result.status == IntegrityStatus.Pending ||
            (result.stale && result.status != IntegrityStatus.Valid)
        ) revert InvalidVerificationResult();

        request.fulfilled = true;
        request.result = result;
        request.fulfilledBy = msg.sender;
        request.fulfilledAt = block.timestamp;
        emit VerificationFulfilled(requestId, request.recordId, msg.sender, result.status, result.stale);
    }

    function getVerification(uint256 requestId) external view returns (VerificationRequest memory) {
        _requireVerification(requestId);
        return verificationRequests[requestId];
    }

    function _appVersionKey(string calldata app, string calldata version) private pure returns (bytes32) {
        return keccak256(abi.encode(app, version));
    }

    function _validateString(string calldata value, string memory field, uint256 maximum) private pure {
        uint256 length = bytes(value).length;
        if (length == 0) revert EmptyField(field);
        if (length > maximum) revert StringTooLong(field, length, maximum);
    }

    function _requireRecord(uint256 recordId) private view {
        if (recordId == 0 || recordId > recordCount) revert RecordNotFound(recordId);
    }

    function _requireVerification(uint256 requestId) private view {
        if (requestId == 0 || requestId > verificationRequestCount) {
            revert VerificationRequestNotFound(requestId);
        }
    }
}
