// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package registry

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
	_ = time.Tick
	_ = context.Background
)

// AttestationRegistryRecord is an auto generated low-level Go binding around an user-defined struct.
type AttestationRegistryRecord struct {
	App          string
	Version      string
	ScanId       string
	ArtifactHash [32]byte
	SbomHash     [32]byte
	ScanHash     [32]byte
	Signer       common.Address
	Timestamp    *big.Int
}

// AttestationRegistryVerificationRequest is an auto generated low-level Go binding around an user-defined struct.
type AttestationRegistryVerificationRequest struct {
	RecordId    *big.Int
	Requester   common.Address
	RequestedAt *big.Int
	Fulfilled   bool
	Result      AttestationRegistryVerificationResult
	FulfilledBy common.Address
	FulfilledAt *big.Int
}

// AttestationRegistryVerificationResult is an auto generated low-level Go binding around an user-defined struct.
type AttestationRegistryVerificationResult struct {
	Status uint8
	Stale  bool
}

// AttestationRegistryMetaData contains all meta data concerning the AttestationRegistry contract.
var AttestationRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"field\",\"type\":\"string\"}],\"name\":\"EmptyField\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"field\",\"type\":\"string\"}],\"name\":\"EmptyHash\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"InvalidPageLimit\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidVerificationResult\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"app\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"}],\"name\":\"NoRecordsForAppVersion\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oracle\",\"type\":\"address\"}],\"name\":\"OracleAlreadyAuthorized\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oracle\",\"type\":\"address\"}],\"name\":\"OracleNotAuthorized\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"}],\"name\":\"RecordNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"scanId\",\"type\":\"string\"}],\"name\":\"ScanIdAlreadyUsed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"field\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maximum\",\"type\":\"uint256\"}],\"name\":\"StringTooLong\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"UnauthorizedAdmin\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"UnauthorizedOracle\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"}],\"name\":\"VerificationAlreadyFulfilled\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"}],\"name\":\"VerificationRequestNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oracle\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"authorized\",\"type\":\"bool\"}],\"name\":\"OracleAuthorizationChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"appVersionKey\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"app\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"scanId\",\"type\":\"string\"}],\"name\":\"RecordAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oracle\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"enumAttestationRegistry.IntegrityStatus\",\"name\":\"status\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"stale\",\"type\":\"bool\"}],\"name\":\"VerificationFulfilled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"}],\"name\":\"VerificationRequested\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_APP_BYTES\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_PAGE_SIZE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_SCAN_ID_BYTES\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_VERSION_BYTES\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"app\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"scanId\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"artifactHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"sbomHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"scanHash\",\"type\":\"bytes32\"}],\"name\":\"addRecord\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oracle\",\"type\":\"address\"}],\"name\":\"authorizeOracle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"enumAttestationRegistry.IntegrityStatus\",\"name\":\"status\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"stale\",\"type\":\"bool\"}],\"internalType\":\"structAttestationRegistry.VerificationResult\",\"name\":\"result\",\"type\":\"tuple\"}],\"name\":\"fulfillVerification\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"app\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"}],\"name\":\"getLatestRecord\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"app\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"scanId\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"artifactHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"sbomHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"scanHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"internalType\":\"structAttestationRegistry.Record\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"}],\"name\":\"getRecord\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"app\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"scanId\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"artifactHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"sbomHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"scanHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"internalType\":\"structAttestationRegistry.Record\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"app\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"getRecordIds\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"page\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"}],\"name\":\"getVerification\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"requester\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"requestedAt\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"fulfilled\",\"type\":\"bool\"},{\"components\":[{\"internalType\":\"enumAttestationRegistry.IntegrityStatus\",\"name\":\"status\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"stale\",\"type\":\"bool\"}],\"internalType\":\"structAttestationRegistry.VerificationResult\",\"name\":\"result\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"fulfilledBy\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"fulfilledAt\",\"type\":\"uint256\"}],\"internalType\":\"structAttestationRegistry.VerificationRequest\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oracle\",\"type\":\"address\"}],\"name\":\"isAuthorizedOracle\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"scanId\",\"type\":\"string\"}],\"name\":\"isScanIdUsed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"recordCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"}],\"name\":\"requestVerification\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oracle\",\"type\":\"address\"}],\"name\":\"revokeOracle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"verificationRequestCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60a0604052348015600e575f5ffd5b5033608052608051611b736100395f395f818161027c0152818161055c0152610cc80152611b735ff3fe608060405234801561000f575f5ffd5b5060043610610111575f3560e01c806358937cc81161009e578063900407bc1161006e578063900407bc1461023b578063bfe609da14610243578063cd7fc46414610233578063ce088ccc1461026e578063f851a44014610277575f5ffd5b806358937cc8146101ed5780635983e6b01461020d5780635b988886146102205780638813834814610233575f5ffd5b8063229df36a116100e4578063229df36a1461017c5780632d8833e11461018f57806345f857db146101a257806348f4da20146101c55780634f7a7174146101cd575f5ffd5b806303e9e6091461011557806304d05d4b1461013e5780630c020f08146101545780630f13b76314610167575b5f5ffd5b610128610123366004611352565b6102b6565b6040516101359190611397565b60405180910390f35b61014660d681565b604051908152602001610135565b610146610162366004611352565b6104d1565b61017a610175366004611440565b610551565b005b61017a61018a36600461146d565b610669565b61012861019d3660046114e7565b6107f0565b6101b56101b0366004611553565b610a7f565b6040519015158152602001610135565b610146606481565b6101e06101db366004611592565b610ab8565b604051610135919061160f565b6102006101fb366004611352565b610be8565b6040516101359190611685565b61017a61021b366004611440565b610cbd565b61014661022e366004611704565b610dc0565b610146608081565b6101465f5481565b6101b5610251366004611440565b6001600160a01b03165f9081526005602052604090205460ff1690565b61014660015481565b61029e7f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b039091168152602001610135565b6102be6112af565b6102c7826111c5565b5f8281526002602052604090819020815161010081019092528054829082906102ef906117b9565b80601f016020809104026020016040519081016040528092919081815260200182805461031b906117b9565b80156103665780601f1061033d57610100808354040283529160200191610366565b820191905f5260205f20905b81548152906001019060200180831161034957829003601f168201915b5050505050815260200160018201805461037f906117b9565b80601f01602080910402602001604051908101604052809291908181526020018280546103ab906117b9565b80156103f65780601f106103cd576101008083540402835291602001916103f6565b820191905f5260205f20905b8154815290600101906020018083116103d957829003601f168201915b5050505050815260200160028201805461040f906117b9565b80601f016020809104026020016040519081016040528092919081815260200182805461043b906117b9565b80156104865780601f1061045d57610100808354040283529160200191610486565b820191905f5260205f20905b81548152906001019060200180831161046957829003601f168201915b505050918352505060038201546020820152600482015460408201526005820154606082015260068201546001600160a01b0316608082015260079091015460a09091015292915050565b5f6104db826111c5565b60015f81546104e990611805565b91829055505f818152600660205260408082208581556001810180546001600160a01b031916339081179091554260028301559151939450929091859185917f764966e44b6e46a4f574434e516067d609f891215d196749582e3bc43eb4d08c91a450919050565b336001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016146105a15760405163089210c760e11b81523360048201526024015b60405180910390fd5b6001600160a01b0381166105c85760405163d92e233d60e01b815260040160405180910390fd5b6001600160a01b0381165f9081526005602052604090205460ff161561060c5760405163e972640560e01b81526001600160a01b0382166004820152602401610598565b6001600160a01b0381165f81815260056020908152604091829020805460ff1916600190811790915591519182527f0aec3411330a268a33a84b9995eb60fe466ecb481eb84f18497e12f5a2da1c5c91015b60405180910390a250565b335f9081526005602052604090205460ff1661069a576040516391e4784360e01b8152336004820152602401610598565b6106a3826111f6565b5f828152600660205260409020600381015460ff16156106d95760405163a4e52c7b60e01b815260048101849052602401610598565b5f6106e76020840184611829565b60078111156106f8576106f8611651565b1480610739575061070f6040830160208401611851565b8015610739575060016107256020840184611829565b600781111561073657610736611651565b14155b1561075757604051635413bdb960e11b815260040160405180910390fd5b60038101805460ff191660011790558160048201610775828261186c565b50506005810180546001600160a01b031916339081179091554260068301558154847f97331a6b900e788f1ad34c7be03e9af54f43dcb9c49a359e4ac5723bea6a5fb86107c56020870187611829565b6107d56040880160208901611851565b6040516107e39291906118cd565b60405180910390a4505050565b6107f86112af565b5f60035f61080888888888611225565b81526020019081526020015f20905080805490505f036108435785858585604051633cf44cfb60e21b81526004016105989493929190611912565b60025f82600184805490506108589190611943565b8154811061086857610868611956565b905f5260205f20015481526020019081526020015f20604051806101000160405290815f82018054610899906117b9565b80601f01602080910402602001604051908101604052809291908181526020018280546108c5906117b9565b80156109105780601f106108e757610100808354040283529160200191610910565b820191905f5260205f20905b8154815290600101906020018083116108f357829003601f168201915b50505050508152602001600182018054610929906117b9565b80601f0160208091040260200160405190810160405280929190818152602001828054610955906117b9565b80156109a05780601f10610977576101008083540402835291602001916109a0565b820191905f5260205f20905b81548152906001019060200180831161098357829003601f168201915b505050505081526020016002820180546109b9906117b9565b80601f01602080910402602001604051908101604052809291908181526020018280546109e5906117b9565b8015610a305780601f10610a0757610100808354040283529160200191610a30565b820191905f5260205f20905b815481529060010190602001808311610a1357829003601f168201915b505050918352505060038201546020820152600482015460408201526005820154606082015260068201546001600160a01b0316608082015260079091015460a0909101529695505050505050565b5f60045f8484604051610a9392919061196a565b604080519182900390912082526020820192909252015f205460ff1690505b92915050565b6060811580610ac75750606482115b15610ae75760405162aec5cf60e61b815260048101839052602401610598565b5f60035f610af78a8a8a8a611225565b81526020019081526020015f20905080805490508410610b26575050604080515f815260208101909152610bde565b80545f90610b35908690611943565b905083811115610b425750825b8067ffffffffffffffff811115610b5b57610b5b611979565b604051908082528060200260200182016040528015610b84578160200160208202803683370190505b5092505f5b81811015610bda5782610b9c828861198d565b81548110610bac57610bac611956565b905f5260205f200154848281518110610bc757610bc7611956565b6020908102919091010152600101610b89565b5050505b9695505050505050565b610bf06112f8565b610bf9826111f6565b5f82815260066020908152604091829020825160e0810184528154815260018201546001600160a01b031692810192909252600281015482840152600381015460ff908116151560608401528351808501909452600482018054939492936080860193928391166007811115610c7157610c71611651565b6007811115610c8257610c82611651565b81529054610100900460ff16151560209182015290825260058301546001600160a01b03169082015260069091015460409091015292915050565b336001600160a01b037f00000000000000000000000000000000000000000000000000000000000000001614610d085760405163089210c760e11b8152336004820152602401610598565b6001600160a01b038116610d2f5760405163d92e233d60e01b815260040160405180910390fd5b6001600160a01b0381165f9081526005602052604090205460ff16610d725760405163d801a0a960e01b81526001600160a01b0382166004820152602401610598565b6001600160a01b0381165f818152600560209081526040808320805460ff19169055519182527f0aec3411330a268a33a84b9995eb60fe466ecb481eb84f18497e12f5a2da1c5c910161065e565b335f9081526005602052604081205460ff16610df1576040516391e4784360e01b8152336004820152602401610598565b610e198a8a6040518060400160405280600381526020016206170760ec1b81525060d661125d565b610e458888604051806040016040528060078152602001663b32b939b4b7b760c91b815250608061125d565b610e708686604051806040016040528060068152602001651cd8d85b925960d21b815250608061125d565b83610ead57604051630274a90b60e21b815260206004820152600c60248201526b0c2e4e8d2ccc2c6e890c2e6d60a31b6044820152606401610598565b82610ee657604051630274a90b60e21b81526020600482015260086024820152670e6c4deda90c2e6d60c31b6044820152606401610598565b81610f1f57604051630274a90b60e21b81526020600482015260086024820152670e6c6c2dc90c2e6d60c31b6044820152606401610598565b5f8686604051610f3092919061196a565b60408051918290039091205f8181526004602052919091205490915060ff1615610f71578686604051637564601760e01b81526004016105989291906119a0565b5f610f7e8c8c8c8c611225565b90505f5f8154610f8d90611805565b91905081905592506040518061010001604052808d8d8080601f0160208091040260200160405190810160405280939291908181526020018383808284375f92019190915250505090825250604080516020601f8e018190048102820181019092528c815291810191908d908d90819084018382808284375f92019190915250505090825250604080516020601f8c018190048102820181019092528a815291810191908b908b90819084018382808284375f92018290525093855250505060208083018a905260408084018a9052606084018990523360808501524260a0909401939093528682526002905220815181906110899082611a04565b506020820151600182019061109e9082611a04565b50604082015160028201906110b39082611a04565b50606082015181600301556080820151816004015560a0820151816005015560c0820151816006015f6101000a8154816001600160a01b0302191690836001600160a01b0316021790555060e08201518160070155905050600160045f8481526020019081526020015f205f6101000a81548160ff02191690831515021790555060035f8281526020019081526020015f2083908060018154018082558091505060019003905f5260205f20015f9091909190915055336001600160a01b031681847f6d7d428869e2f71e1ac781d7735767a1232bcf781a8f24810fae780567d8a1d38f8f8f8f8f8f6040516111ae96959493929190611abf565b60405180910390a450509998505050505050505050565b8015806111d257505f5481115b156111f357604051635dab5c9b60e01b815260048101829052602401610598565b50565b801580611204575060015481115b156111f35760405163367f52b360e21b815260048101829052602401610598565b5f8484848460405160200161123d9493929190611912565b604051602081830303815290604052805190602001209050949350505050565b825f819003611281578260405163a69290e760e01b81526004016105989190611b07565b818111156112a85782818360405163d7d74b9d60e01b815260040161059893929190611b19565b5050505050565b6040518061010001604052806060815260200160608152602001606081526020015f81526020015f81526020015f81526020015f6001600160a01b031681526020015f81525090565b6040518060e001604052805f81526020015f6001600160a01b031681526020015f81526020015f1515815260200161133f604080518082019091525f808252602082015290565b81525f6020820181905260409091015290565b5f60208284031215611362575f5ffd5b5035919050565b5f81518084528060208401602086015e5f602082860101526020601f19601f83011685010191505092915050565b602081525f825161010060208401526113b4610120840182611369565b90506020840151601f198483030160408501526113d18282611369565b9150506040840151601f198483030160608501526113ef8282611369565b91505060608401516080840152608084015160a084015260a084015160c084015260c084015161142a60e08501826001600160a01b03169052565b5060e08401516101008401528091505092915050565b5f60208284031215611450575f5ffd5b81356001600160a01b0381168114611466575f5ffd5b9392505050565b5f5f828403606081121561147f575f5ffd5b833592506040601f1982011215611494575f5ffd5b506020830190509250929050565b5f5f83601f8401126114b2575f5ffd5b50813567ffffffffffffffff8111156114c9575f5ffd5b6020830191508360208285010111156114e0575f5ffd5b9250929050565b5f5f5f5f604085870312156114fa575f5ffd5b843567ffffffffffffffff811115611510575f5ffd5b61151c878288016114a2565b909550935050602085013567ffffffffffffffff81111561153b575f5ffd5b611547878288016114a2565b95989497509550505050565b5f5f60208385031215611564575f5ffd5b823567ffffffffffffffff81111561157a575f5ffd5b611586858286016114a2565b90969095509350505050565b5f5f5f5f5f5f608087890312156115a7575f5ffd5b863567ffffffffffffffff8111156115bd575f5ffd5b6115c989828a016114a2565b909750955050602087013567ffffffffffffffff8111156115e8575f5ffd5b6115f489828a016114a2565b979a9699509760408101359660609091013595509350505050565b602080825282518282018190525f918401906040840190835b81811015611646578351835260209384019390920191600101611628565b509095945050505050565b634e487b7160e01b5f52602160045260245ffd5b6008811061168157634e487b7160e01b5f52602160045260245ffd5b9052565b5f610100820190508251825260018060a01b0360208401511660208301526040830151604083015260608301511515606083015260808301516116cc608084018251611665565b6020810151151560a08401525060a08301516116f360c08401826001600160a01b03169052565b5060c083015160e083015292915050565b5f5f5f5f5f5f5f5f5f60c08a8c03121561171c575f5ffd5b893567ffffffffffffffff811115611732575f5ffd5b61173e8c828d016114a2565b909a5098505060208a013567ffffffffffffffff81111561175d575f5ffd5b6117698c828d016114a2565b90985096505060408a013567ffffffffffffffff811115611788575f5ffd5b6117948c828d016114a2565b9a9d999c50979a969997986060880135976080810135975060a0013595509350505050565b600181811c908216806117cd57607f821691505b6020821081036117eb57634e487b7160e01b5f52602260045260245ffd5b50919050565b634e487b7160e01b5f52601160045260245ffd5b5f60018201611816576118166117f1565b5060010190565b600881106111f3575f5ffd5b5f60208284031215611839575f5ffd5b81356114668161181d565b80151581146111f3575f5ffd5b5f60208284031215611861575f5ffd5b813561146681611844565b81356118778161181d565b6008811061189357634e487b7160e01b5f52602160045260245ffd5b815460ff821691508160ff19821617835560208401356118b281611844565b61ffff199190911690911790151560081b61ff001617905550565b604081016118db8285611665565b82151560208301529392505050565b81835281816020850137505f828201602090810191909152601f909101601f19169091010190565b604081525f6119256040830186886118ea565b82810360208401526119388185876118ea565b979650505050505050565b81810381811115610ab257610ab26117f1565b634e487b7160e01b5f52603260045260245ffd5b818382375f9101908152919050565b634e487b7160e01b5f52604160045260245ffd5b80820180821115610ab257610ab26117f1565b602081525f6119b36020830184866118ea565b949350505050565b601f8211156119ff57805f5260205f20601f840160051c810160208510156119e05750805b601f840160051c820191505b818110156112a8575f81556001016119ec565b505050565b815167ffffffffffffffff811115611a1e57611a1e611979565b611a3281611a2c84546117b9565b846119bb565b6020601f821160018114611a64575f8315611a4d5750848201515b5f19600385901b1c1916600184901b1784556112a8565b5f84815260208120601f198516915b82811015611a935787850151825560209485019460019092019101611a73565b5084821015611ab057868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b606081525f611ad260608301888a6118ea565b8281036020840152611ae58187896118ea565b90508281036040840152611afa8185876118ea565b9998505050505050505050565b602081525f6114666020830184611369565b606081525f611b2b6060830186611369565b6020830194909452506040015291905056fea26469706673582212201643b12fd326f31f94cef6df05a53dd2cec54d457af42384d6e72c07c7c4c0e764736f6c634300081c0033",
}

// AttestationRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use AttestationRegistryMetaData.ABI instead.
var AttestationRegistryABI = AttestationRegistryMetaData.ABI

// AttestationRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AttestationRegistryMetaData.Bin instead.
var AttestationRegistryBin = AttestationRegistryMetaData.Bin

// DeployAttestationRegistry deploys a new Ethereum contract, binding an instance of AttestationRegistry to it.
func DeployAttestationRegistry(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *AttestationRegistry, error) {
	parsed, err := AttestationRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AttestationRegistryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AttestationRegistry{AttestationRegistryCaller: AttestationRegistryCaller{contract: contract}, AttestationRegistryTransactor: AttestationRegistryTransactor{contract: contract}, AttestationRegistryFilterer: AttestationRegistryFilterer{contract: contract}}, nil
}

// AttestationRegistry is an auto generated Go binding around an Ethereum contract.
type AttestationRegistry struct {
	AttestationRegistryCaller     // Read-only binding to the contract
	AttestationRegistryTransactor // Write-only binding to the contract
	AttestationRegistryFilterer   // Log filterer for contract events
}

// AttestationRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type AttestationRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AttestationRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AttestationRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AttestationRegistrySession struct {
	Contract     *AttestationRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// AttestationRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AttestationRegistryCallerSession struct {
	Contract *AttestationRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// AttestationRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AttestationRegistryTransactorSession struct {
	Contract     *AttestationRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// AttestationRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type AttestationRegistryRaw struct {
	Contract *AttestationRegistry // Generic contract binding to access the raw methods on
}

// AttestationRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AttestationRegistryCallerRaw struct {
	Contract *AttestationRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// AttestationRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AttestationRegistryTransactorRaw struct {
	Contract *AttestationRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAttestationRegistry creates a new instance of AttestationRegistry, bound to a specific deployed contract.
func NewAttestationRegistry(address common.Address, backend bind.ContractBackend) (*AttestationRegistry, error) {
	contract, err := bindAttestationRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistry{AttestationRegistryCaller: AttestationRegistryCaller{contract: contract}, AttestationRegistryTransactor: AttestationRegistryTransactor{contract: contract}, AttestationRegistryFilterer: AttestationRegistryFilterer{contract: contract}}, nil
}

// NewAttestationRegistryCaller creates a new read-only instance of AttestationRegistry, bound to a specific deployed contract.
func NewAttestationRegistryCaller(address common.Address, caller bind.ContractCaller) (*AttestationRegistryCaller, error) {
	contract, err := bindAttestationRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryCaller{contract: contract}, nil
}

// NewAttestationRegistryTransactor creates a new write-only instance of AttestationRegistry, bound to a specific deployed contract.
func NewAttestationRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*AttestationRegistryTransactor, error) {
	contract, err := bindAttestationRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryTransactor{contract: contract}, nil
}

// NewAttestationRegistryFilterer creates a new log filterer instance of AttestationRegistry, bound to a specific deployed contract.
func NewAttestationRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*AttestationRegistryFilterer, error) {
	contract, err := bindAttestationRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryFilterer{contract: contract}, nil
}

// bindAttestationRegistry binds a generic wrapper to an already deployed contract.
func bindAttestationRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AttestationRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AttestationRegistry *AttestationRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AttestationRegistry.Contract.AttestationRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AttestationRegistry *AttestationRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.AttestationRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AttestationRegistry *AttestationRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.AttestationRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AttestationRegistry *AttestationRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AttestationRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AttestationRegistry *AttestationRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AttestationRegistry *AttestationRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.contract.Transact(opts, method, params...)
}

// MAXAPPBYTES is a free data retrieval call binding the contract method 0x04d05d4b.
//
// Solidity: function MAX_APP_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCaller) MAXAPPBYTES(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "MAX_APP_BYTES")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXAPPBYTES is a free data retrieval call binding the contract method 0x04d05d4b.
//
// Solidity: function MAX_APP_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistrySession) MAXAPPBYTES() (*big.Int, error) {
	return _AttestationRegistry.Contract.MAXAPPBYTES(&_AttestationRegistry.CallOpts)
}

// MAXAPPBYTES is a free data retrieval call binding the contract method 0x04d05d4b.
//
// Solidity: function MAX_APP_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCallerSession) MAXAPPBYTES() (*big.Int, error) {
	return _AttestationRegistry.Contract.MAXAPPBYTES(&_AttestationRegistry.CallOpts)
}

// MAXPAGESIZE is a free data retrieval call binding the contract method 0x48f4da20.
//
// Solidity: function MAX_PAGE_SIZE() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCaller) MAXPAGESIZE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "MAX_PAGE_SIZE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXPAGESIZE is a free data retrieval call binding the contract method 0x48f4da20.
//
// Solidity: function MAX_PAGE_SIZE() view returns(uint256)
func (_AttestationRegistry *AttestationRegistrySession) MAXPAGESIZE() (*big.Int, error) {
	return _AttestationRegistry.Contract.MAXPAGESIZE(&_AttestationRegistry.CallOpts)
}

// MAXPAGESIZE is a free data retrieval call binding the contract method 0x48f4da20.
//
// Solidity: function MAX_PAGE_SIZE() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCallerSession) MAXPAGESIZE() (*big.Int, error) {
	return _AttestationRegistry.Contract.MAXPAGESIZE(&_AttestationRegistry.CallOpts)
}

// MAXSCANIDBYTES is a free data retrieval call binding the contract method 0x88138348.
//
// Solidity: function MAX_SCAN_ID_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCaller) MAXSCANIDBYTES(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "MAX_SCAN_ID_BYTES")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXSCANIDBYTES is a free data retrieval call binding the contract method 0x88138348.
//
// Solidity: function MAX_SCAN_ID_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistrySession) MAXSCANIDBYTES() (*big.Int, error) {
	return _AttestationRegistry.Contract.MAXSCANIDBYTES(&_AttestationRegistry.CallOpts)
}

// MAXSCANIDBYTES is a free data retrieval call binding the contract method 0x88138348.
//
// Solidity: function MAX_SCAN_ID_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCallerSession) MAXSCANIDBYTES() (*big.Int, error) {
	return _AttestationRegistry.Contract.MAXSCANIDBYTES(&_AttestationRegistry.CallOpts)
}

// MAXVERSIONBYTES is a free data retrieval call binding the contract method 0xcd7fc464.
//
// Solidity: function MAX_VERSION_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCaller) MAXVERSIONBYTES(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "MAX_VERSION_BYTES")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXVERSIONBYTES is a free data retrieval call binding the contract method 0xcd7fc464.
//
// Solidity: function MAX_VERSION_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistrySession) MAXVERSIONBYTES() (*big.Int, error) {
	return _AttestationRegistry.Contract.MAXVERSIONBYTES(&_AttestationRegistry.CallOpts)
}

// MAXVERSIONBYTES is a free data retrieval call binding the contract method 0xcd7fc464.
//
// Solidity: function MAX_VERSION_BYTES() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCallerSession) MAXVERSIONBYTES() (*big.Int, error) {
	return _AttestationRegistry.Contract.MAXVERSIONBYTES(&_AttestationRegistry.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_AttestationRegistry *AttestationRegistryCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "admin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_AttestationRegistry *AttestationRegistrySession) Admin() (common.Address, error) {
	return _AttestationRegistry.Contract.Admin(&_AttestationRegistry.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() view returns(address)
func (_AttestationRegistry *AttestationRegistryCallerSession) Admin() (common.Address, error) {
	return _AttestationRegistry.Contract.Admin(&_AttestationRegistry.CallOpts)
}

// GetLatestRecord is a free data retrieval call binding the contract method 0x2d8833e1.
//
// Solidity: function getLatestRecord(string app, string version) view returns((string,string,string,bytes32,bytes32,bytes32,address,uint256))
func (_AttestationRegistry *AttestationRegistryCaller) GetLatestRecord(opts *bind.CallOpts, app string, version string) (AttestationRegistryRecord, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "getLatestRecord", app, version)

	if err != nil {
		return *new(AttestationRegistryRecord), err
	}

	out0 := *abi.ConvertType(out[0], new(AttestationRegistryRecord)).(*AttestationRegistryRecord)

	return out0, err

}

// GetLatestRecord is a free data retrieval call binding the contract method 0x2d8833e1.
//
// Solidity: function getLatestRecord(string app, string version) view returns((string,string,string,bytes32,bytes32,bytes32,address,uint256))
func (_AttestationRegistry *AttestationRegistrySession) GetLatestRecord(app string, version string) (AttestationRegistryRecord, error) {
	return _AttestationRegistry.Contract.GetLatestRecord(&_AttestationRegistry.CallOpts, app, version)
}

// GetLatestRecord is a free data retrieval call binding the contract method 0x2d8833e1.
//
// Solidity: function getLatestRecord(string app, string version) view returns((string,string,string,bytes32,bytes32,bytes32,address,uint256))
func (_AttestationRegistry *AttestationRegistryCallerSession) GetLatestRecord(app string, version string) (AttestationRegistryRecord, error) {
	return _AttestationRegistry.Contract.GetLatestRecord(&_AttestationRegistry.CallOpts, app, version)
}

// GetRecord is a free data retrieval call binding the contract method 0x03e9e609.
//
// Solidity: function getRecord(uint256 recordId) view returns((string,string,string,bytes32,bytes32,bytes32,address,uint256))
func (_AttestationRegistry *AttestationRegistryCaller) GetRecord(opts *bind.CallOpts, recordId *big.Int) (AttestationRegistryRecord, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "getRecord", recordId)

	if err != nil {
		return *new(AttestationRegistryRecord), err
	}

	out0 := *abi.ConvertType(out[0], new(AttestationRegistryRecord)).(*AttestationRegistryRecord)

	return out0, err

}

// GetRecord is a free data retrieval call binding the contract method 0x03e9e609.
//
// Solidity: function getRecord(uint256 recordId) view returns((string,string,string,bytes32,bytes32,bytes32,address,uint256))
func (_AttestationRegistry *AttestationRegistrySession) GetRecord(recordId *big.Int) (AttestationRegistryRecord, error) {
	return _AttestationRegistry.Contract.GetRecord(&_AttestationRegistry.CallOpts, recordId)
}

// GetRecord is a free data retrieval call binding the contract method 0x03e9e609.
//
// Solidity: function getRecord(uint256 recordId) view returns((string,string,string,bytes32,bytes32,bytes32,address,uint256))
func (_AttestationRegistry *AttestationRegistryCallerSession) GetRecord(recordId *big.Int) (AttestationRegistryRecord, error) {
	return _AttestationRegistry.Contract.GetRecord(&_AttestationRegistry.CallOpts, recordId)
}

// GetRecordIds is a free data retrieval call binding the contract method 0x4f7a7174.
//
// Solidity: function getRecordIds(string app, string version, uint256 offset, uint256 limit) view returns(uint256[] page)
func (_AttestationRegistry *AttestationRegistryCaller) GetRecordIds(opts *bind.CallOpts, app string, version string, offset *big.Int, limit *big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "getRecordIds", app, version, offset, limit)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetRecordIds is a free data retrieval call binding the contract method 0x4f7a7174.
//
// Solidity: function getRecordIds(string app, string version, uint256 offset, uint256 limit) view returns(uint256[] page)
func (_AttestationRegistry *AttestationRegistrySession) GetRecordIds(app string, version string, offset *big.Int, limit *big.Int) ([]*big.Int, error) {
	return _AttestationRegistry.Contract.GetRecordIds(&_AttestationRegistry.CallOpts, app, version, offset, limit)
}

// GetRecordIds is a free data retrieval call binding the contract method 0x4f7a7174.
//
// Solidity: function getRecordIds(string app, string version, uint256 offset, uint256 limit) view returns(uint256[] page)
func (_AttestationRegistry *AttestationRegistryCallerSession) GetRecordIds(app string, version string, offset *big.Int, limit *big.Int) ([]*big.Int, error) {
	return _AttestationRegistry.Contract.GetRecordIds(&_AttestationRegistry.CallOpts, app, version, offset, limit)
}

// GetVerification is a free data retrieval call binding the contract method 0x58937cc8.
//
// Solidity: function getVerification(uint256 requestId) view returns((uint256,address,uint256,bool,(uint8,bool),address,uint256))
func (_AttestationRegistry *AttestationRegistryCaller) GetVerification(opts *bind.CallOpts, requestId *big.Int) (AttestationRegistryVerificationRequest, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "getVerification", requestId)

	if err != nil {
		return *new(AttestationRegistryVerificationRequest), err
	}

	out0 := *abi.ConvertType(out[0], new(AttestationRegistryVerificationRequest)).(*AttestationRegistryVerificationRequest)

	return out0, err

}

// GetVerification is a free data retrieval call binding the contract method 0x58937cc8.
//
// Solidity: function getVerification(uint256 requestId) view returns((uint256,address,uint256,bool,(uint8,bool),address,uint256))
func (_AttestationRegistry *AttestationRegistrySession) GetVerification(requestId *big.Int) (AttestationRegistryVerificationRequest, error) {
	return _AttestationRegistry.Contract.GetVerification(&_AttestationRegistry.CallOpts, requestId)
}

// GetVerification is a free data retrieval call binding the contract method 0x58937cc8.
//
// Solidity: function getVerification(uint256 requestId) view returns((uint256,address,uint256,bool,(uint8,bool),address,uint256))
func (_AttestationRegistry *AttestationRegistryCallerSession) GetVerification(requestId *big.Int) (AttestationRegistryVerificationRequest, error) {
	return _AttestationRegistry.Contract.GetVerification(&_AttestationRegistry.CallOpts, requestId)
}

// IsAuthorizedOracle is a free data retrieval call binding the contract method 0xbfe609da.
//
// Solidity: function isAuthorizedOracle(address oracle) view returns(bool)
func (_AttestationRegistry *AttestationRegistryCaller) IsAuthorizedOracle(opts *bind.CallOpts, oracle common.Address) (bool, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "isAuthorizedOracle", oracle)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAuthorizedOracle is a free data retrieval call binding the contract method 0xbfe609da.
//
// Solidity: function isAuthorizedOracle(address oracle) view returns(bool)
func (_AttestationRegistry *AttestationRegistrySession) IsAuthorizedOracle(oracle common.Address) (bool, error) {
	return _AttestationRegistry.Contract.IsAuthorizedOracle(&_AttestationRegistry.CallOpts, oracle)
}

// IsAuthorizedOracle is a free data retrieval call binding the contract method 0xbfe609da.
//
// Solidity: function isAuthorizedOracle(address oracle) view returns(bool)
func (_AttestationRegistry *AttestationRegistryCallerSession) IsAuthorizedOracle(oracle common.Address) (bool, error) {
	return _AttestationRegistry.Contract.IsAuthorizedOracle(&_AttestationRegistry.CallOpts, oracle)
}

// IsScanIdUsed is a free data retrieval call binding the contract method 0x45f857db.
//
// Solidity: function isScanIdUsed(string scanId) view returns(bool)
func (_AttestationRegistry *AttestationRegistryCaller) IsScanIdUsed(opts *bind.CallOpts, scanId string) (bool, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "isScanIdUsed", scanId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsScanIdUsed is a free data retrieval call binding the contract method 0x45f857db.
//
// Solidity: function isScanIdUsed(string scanId) view returns(bool)
func (_AttestationRegistry *AttestationRegistrySession) IsScanIdUsed(scanId string) (bool, error) {
	return _AttestationRegistry.Contract.IsScanIdUsed(&_AttestationRegistry.CallOpts, scanId)
}

// IsScanIdUsed is a free data retrieval call binding the contract method 0x45f857db.
//
// Solidity: function isScanIdUsed(string scanId) view returns(bool)
func (_AttestationRegistry *AttestationRegistryCallerSession) IsScanIdUsed(scanId string) (bool, error) {
	return _AttestationRegistry.Contract.IsScanIdUsed(&_AttestationRegistry.CallOpts, scanId)
}

// RecordCount is a free data retrieval call binding the contract method 0x900407bc.
//
// Solidity: function recordCount() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCaller) RecordCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "recordCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RecordCount is a free data retrieval call binding the contract method 0x900407bc.
//
// Solidity: function recordCount() view returns(uint256)
func (_AttestationRegistry *AttestationRegistrySession) RecordCount() (*big.Int, error) {
	return _AttestationRegistry.Contract.RecordCount(&_AttestationRegistry.CallOpts)
}

// RecordCount is a free data retrieval call binding the contract method 0x900407bc.
//
// Solidity: function recordCount() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCallerSession) RecordCount() (*big.Int, error) {
	return _AttestationRegistry.Contract.RecordCount(&_AttestationRegistry.CallOpts)
}

// VerificationRequestCount is a free data retrieval call binding the contract method 0xce088ccc.
//
// Solidity: function verificationRequestCount() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCaller) VerificationRequestCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationRegistry.contract.Call(opts, &out, "verificationRequestCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VerificationRequestCount is a free data retrieval call binding the contract method 0xce088ccc.
//
// Solidity: function verificationRequestCount() view returns(uint256)
func (_AttestationRegistry *AttestationRegistrySession) VerificationRequestCount() (*big.Int, error) {
	return _AttestationRegistry.Contract.VerificationRequestCount(&_AttestationRegistry.CallOpts)
}

// VerificationRequestCount is a free data retrieval call binding the contract method 0xce088ccc.
//
// Solidity: function verificationRequestCount() view returns(uint256)
func (_AttestationRegistry *AttestationRegistryCallerSession) VerificationRequestCount() (*big.Int, error) {
	return _AttestationRegistry.Contract.VerificationRequestCount(&_AttestationRegistry.CallOpts)
}

// AddRecord is a paid mutator transaction binding the contract method 0x5b988886.
//
// Solidity: function addRecord(string app, string version, string scanId, bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) returns(uint256 recordId)
func (_AttestationRegistry *AttestationRegistryTransactor) AddRecord(opts *bind.TransactOpts, app string, version string, scanId string, artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (*types.Transaction, error) {
	return _AttestationRegistry.contract.Transact(opts, "addRecord", app, version, scanId, artifactHash, sbomHash, scanHash)
}

// AddRecord is a paid mutator transaction binding the contract method 0x5b988886.
//
// Solidity: function addRecord(string app, string version, string scanId, bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) returns(uint256 recordId)
func (_AttestationRegistry *AttestationRegistrySession) AddRecord(app string, version string, scanId string, artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.AddRecord(&_AttestationRegistry.TransactOpts, app, version, scanId, artifactHash, sbomHash, scanHash)
}

// AddRecord is a paid mutator transaction binding the contract method 0x5b988886.
//
// Solidity: function addRecord(string app, string version, string scanId, bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) returns(uint256 recordId)
func (_AttestationRegistry *AttestationRegistryTransactorSession) AddRecord(app string, version string, scanId string, artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.AddRecord(&_AttestationRegistry.TransactOpts, app, version, scanId, artifactHash, sbomHash, scanHash)
}

// AuthorizeOracle is a paid mutator transaction binding the contract method 0x0f13b763.
//
// Solidity: function authorizeOracle(address oracle) returns()
func (_AttestationRegistry *AttestationRegistryTransactor) AuthorizeOracle(opts *bind.TransactOpts, oracle common.Address) (*types.Transaction, error) {
	return _AttestationRegistry.contract.Transact(opts, "authorizeOracle", oracle)
}

// AuthorizeOracle is a paid mutator transaction binding the contract method 0x0f13b763.
//
// Solidity: function authorizeOracle(address oracle) returns()
func (_AttestationRegistry *AttestationRegistrySession) AuthorizeOracle(oracle common.Address) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.AuthorizeOracle(&_AttestationRegistry.TransactOpts, oracle)
}

// AuthorizeOracle is a paid mutator transaction binding the contract method 0x0f13b763.
//
// Solidity: function authorizeOracle(address oracle) returns()
func (_AttestationRegistry *AttestationRegistryTransactorSession) AuthorizeOracle(oracle common.Address) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.AuthorizeOracle(&_AttestationRegistry.TransactOpts, oracle)
}

// FulfillVerification is a paid mutator transaction binding the contract method 0x229df36a.
//
// Solidity: function fulfillVerification(uint256 requestId, (uint8,bool) result) returns()
func (_AttestationRegistry *AttestationRegistryTransactor) FulfillVerification(opts *bind.TransactOpts, requestId *big.Int, result AttestationRegistryVerificationResult) (*types.Transaction, error) {
	return _AttestationRegistry.contract.Transact(opts, "fulfillVerification", requestId, result)
}

// FulfillVerification is a paid mutator transaction binding the contract method 0x229df36a.
//
// Solidity: function fulfillVerification(uint256 requestId, (uint8,bool) result) returns()
func (_AttestationRegistry *AttestationRegistrySession) FulfillVerification(requestId *big.Int, result AttestationRegistryVerificationResult) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.FulfillVerification(&_AttestationRegistry.TransactOpts, requestId, result)
}

// FulfillVerification is a paid mutator transaction binding the contract method 0x229df36a.
//
// Solidity: function fulfillVerification(uint256 requestId, (uint8,bool) result) returns()
func (_AttestationRegistry *AttestationRegistryTransactorSession) FulfillVerification(requestId *big.Int, result AttestationRegistryVerificationResult) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.FulfillVerification(&_AttestationRegistry.TransactOpts, requestId, result)
}

// RequestVerification is a paid mutator transaction binding the contract method 0x0c020f08.
//
// Solidity: function requestVerification(uint256 recordId) returns(uint256 requestId)
func (_AttestationRegistry *AttestationRegistryTransactor) RequestVerification(opts *bind.TransactOpts, recordId *big.Int) (*types.Transaction, error) {
	return _AttestationRegistry.contract.Transact(opts, "requestVerification", recordId)
}

// RequestVerification is a paid mutator transaction binding the contract method 0x0c020f08.
//
// Solidity: function requestVerification(uint256 recordId) returns(uint256 requestId)
func (_AttestationRegistry *AttestationRegistrySession) RequestVerification(recordId *big.Int) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.RequestVerification(&_AttestationRegistry.TransactOpts, recordId)
}

// RequestVerification is a paid mutator transaction binding the contract method 0x0c020f08.
//
// Solidity: function requestVerification(uint256 recordId) returns(uint256 requestId)
func (_AttestationRegistry *AttestationRegistryTransactorSession) RequestVerification(recordId *big.Int) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.RequestVerification(&_AttestationRegistry.TransactOpts, recordId)
}

// RevokeOracle is a paid mutator transaction binding the contract method 0x5983e6b0.
//
// Solidity: function revokeOracle(address oracle) returns()
func (_AttestationRegistry *AttestationRegistryTransactor) RevokeOracle(opts *bind.TransactOpts, oracle common.Address) (*types.Transaction, error) {
	return _AttestationRegistry.contract.Transact(opts, "revokeOracle", oracle)
}

// RevokeOracle is a paid mutator transaction binding the contract method 0x5983e6b0.
//
// Solidity: function revokeOracle(address oracle) returns()
func (_AttestationRegistry *AttestationRegistrySession) RevokeOracle(oracle common.Address) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.RevokeOracle(&_AttestationRegistry.TransactOpts, oracle)
}

// RevokeOracle is a paid mutator transaction binding the contract method 0x5983e6b0.
//
// Solidity: function revokeOracle(address oracle) returns()
func (_AttestationRegistry *AttestationRegistryTransactorSession) RevokeOracle(oracle common.Address) (*types.Transaction, error) {
	return _AttestationRegistry.Contract.RevokeOracle(&_AttestationRegistry.TransactOpts, oracle)
}

// AttestationRegistryOracleAuthorizationChangedIterator is returned from FilterOracleAuthorizationChanged and is used to iterate over the raw logs and unpacked data for OracleAuthorizationChanged events raised by the AttestationRegistry contract.
type AttestationRegistryOracleAuthorizationChangedIterator struct {
	Event *AttestationRegistryOracleAuthorizationChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationRegistryOracleAuthorizationChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationRegistryOracleAuthorizationChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationRegistryOracleAuthorizationChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationRegistryOracleAuthorizationChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationRegistryOracleAuthorizationChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationRegistryOracleAuthorizationChanged represents a OracleAuthorizationChanged event raised by the AttestationRegistry contract.
type AttestationRegistryOracleAuthorizationChanged struct {
	Oracle     common.Address
	Authorized bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOracleAuthorizationChanged is a free log retrieval operation binding the contract event 0x0aec3411330a268a33a84b9995eb60fe466ecb481eb84f18497e12f5a2da1c5c.
//
// Solidity: event OracleAuthorizationChanged(address indexed oracle, bool authorized)
func (_AttestationRegistry *AttestationRegistryFilterer) FilterOracleAuthorizationChanged(opts *bind.FilterOpts, oracle []common.Address) (*AttestationRegistryOracleAuthorizationChangedIterator, error) {

	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _AttestationRegistry.contract.FilterLogs(opts, "OracleAuthorizationChanged", oracleRule)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryOracleAuthorizationChangedIterator{contract: _AttestationRegistry.contract, event: "OracleAuthorizationChanged", logs: logs, sub: sub}, nil
}

// WatchOracleAuthorizationChanged is a free log subscription operation binding the contract event 0x0aec3411330a268a33a84b9995eb60fe466ecb481eb84f18497e12f5a2da1c5c.
//
// Solidity: event OracleAuthorizationChanged(address indexed oracle, bool authorized)
func (_AttestationRegistry *AttestationRegistryFilterer) WatchOracleAuthorizationChanged(opts *bind.WatchOpts, sink chan<- *AttestationRegistryOracleAuthorizationChanged, oracle []common.Address) (event.Subscription, error) {

	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _AttestationRegistry.contract.WatchLogs(opts, "OracleAuthorizationChanged", oracleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationRegistryOracleAuthorizationChanged)
				if err := _AttestationRegistry.contract.UnpackLog(event, "OracleAuthorizationChanged", log); err != nil {
					// If the signature doesn't match, skip this log.
					if errors.Is(err, bind.ErrEventSignatureMismatch) {
						continue
					}
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOracleAuthorizationChanged is a log parse operation binding the contract event 0x0aec3411330a268a33a84b9995eb60fe466ecb481eb84f18497e12f5a2da1c5c.
//
// Solidity: event OracleAuthorizationChanged(address indexed oracle, bool authorized)
func (_AttestationRegistry *AttestationRegistryFilterer) ParseOracleAuthorizationChanged(log types.Log) (*AttestationRegistryOracleAuthorizationChanged, error) {
	event := new(AttestationRegistryOracleAuthorizationChanged)
	if err := _AttestationRegistry.contract.UnpackLog(event, "OracleAuthorizationChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationRegistryRecordAddedIterator is returned from FilterRecordAdded and is used to iterate over the raw logs and unpacked data for RecordAdded events raised by the AttestationRegistry contract.
type AttestationRegistryRecordAddedIterator struct {
	Event *AttestationRegistryRecordAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationRegistryRecordAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationRegistryRecordAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationRegistryRecordAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationRegistryRecordAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationRegistryRecordAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationRegistryRecordAdded represents a RecordAdded event raised by the AttestationRegistry contract.
type AttestationRegistryRecordAdded struct {
	RecordId      *big.Int
	AppVersionKey [32]byte
	Signer        common.Address
	App           string
	Version       string
	ScanId        string
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRecordAdded is a free log retrieval operation binding the contract event 0x6d7d428869e2f71e1ac781d7735767a1232bcf781a8f24810fae780567d8a1d3.
//
// Solidity: event RecordAdded(uint256 indexed recordId, bytes32 indexed appVersionKey, address indexed signer, string app, string version, string scanId)
func (_AttestationRegistry *AttestationRegistryFilterer) FilterRecordAdded(opts *bind.FilterOpts, recordId []*big.Int, appVersionKey [][32]byte, signer []common.Address) (*AttestationRegistryRecordAddedIterator, error) {

	var recordIdRule []interface{}
	for _, recordIdItem := range recordId {
		recordIdRule = append(recordIdRule, recordIdItem)
	}
	var appVersionKeyRule []interface{}
	for _, appVersionKeyItem := range appVersionKey {
		appVersionKeyRule = append(appVersionKeyRule, appVersionKeyItem)
	}
	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _AttestationRegistry.contract.FilterLogs(opts, "RecordAdded", recordIdRule, appVersionKeyRule, signerRule)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryRecordAddedIterator{contract: _AttestationRegistry.contract, event: "RecordAdded", logs: logs, sub: sub}, nil
}

// WatchRecordAdded is a free log subscription operation binding the contract event 0x6d7d428869e2f71e1ac781d7735767a1232bcf781a8f24810fae780567d8a1d3.
//
// Solidity: event RecordAdded(uint256 indexed recordId, bytes32 indexed appVersionKey, address indexed signer, string app, string version, string scanId)
func (_AttestationRegistry *AttestationRegistryFilterer) WatchRecordAdded(opts *bind.WatchOpts, sink chan<- *AttestationRegistryRecordAdded, recordId []*big.Int, appVersionKey [][32]byte, signer []common.Address) (event.Subscription, error) {

	var recordIdRule []interface{}
	for _, recordIdItem := range recordId {
		recordIdRule = append(recordIdRule, recordIdItem)
	}
	var appVersionKeyRule []interface{}
	for _, appVersionKeyItem := range appVersionKey {
		appVersionKeyRule = append(appVersionKeyRule, appVersionKeyItem)
	}
	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _AttestationRegistry.contract.WatchLogs(opts, "RecordAdded", recordIdRule, appVersionKeyRule, signerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationRegistryRecordAdded)
				if err := _AttestationRegistry.contract.UnpackLog(event, "RecordAdded", log); err != nil {
					// If the signature doesn't match, skip this log.
					if errors.Is(err, bind.ErrEventSignatureMismatch) {
						continue
					}
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRecordAdded is a log parse operation binding the contract event 0x6d7d428869e2f71e1ac781d7735767a1232bcf781a8f24810fae780567d8a1d3.
//
// Solidity: event RecordAdded(uint256 indexed recordId, bytes32 indexed appVersionKey, address indexed signer, string app, string version, string scanId)
func (_AttestationRegistry *AttestationRegistryFilterer) ParseRecordAdded(log types.Log) (*AttestationRegistryRecordAdded, error) {
	event := new(AttestationRegistryRecordAdded)
	if err := _AttestationRegistry.contract.UnpackLog(event, "RecordAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationRegistryVerificationFulfilledIterator is returned from FilterVerificationFulfilled and is used to iterate over the raw logs and unpacked data for VerificationFulfilled events raised by the AttestationRegistry contract.
type AttestationRegistryVerificationFulfilledIterator struct {
	Event *AttestationRegistryVerificationFulfilled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationRegistryVerificationFulfilledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationRegistryVerificationFulfilled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationRegistryVerificationFulfilled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationRegistryVerificationFulfilledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationRegistryVerificationFulfilledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationRegistryVerificationFulfilled represents a VerificationFulfilled event raised by the AttestationRegistry contract.
type AttestationRegistryVerificationFulfilled struct {
	RequestId *big.Int
	RecordId  *big.Int
	Oracle    common.Address
	Status    uint8
	Stale     bool
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterVerificationFulfilled is a free log retrieval operation binding the contract event 0x97331a6b900e788f1ad34c7be03e9af54f43dcb9c49a359e4ac5723bea6a5fb8.
//
// Solidity: event VerificationFulfilled(uint256 indexed requestId, uint256 indexed recordId, address indexed oracle, uint8 status, bool stale)
func (_AttestationRegistry *AttestationRegistryFilterer) FilterVerificationFulfilled(opts *bind.FilterOpts, requestId []*big.Int, recordId []*big.Int, oracle []common.Address) (*AttestationRegistryVerificationFulfilledIterator, error) {

	var requestIdRule []interface{}
	for _, requestIdItem := range requestId {
		requestIdRule = append(requestIdRule, requestIdItem)
	}
	var recordIdRule []interface{}
	for _, recordIdItem := range recordId {
		recordIdRule = append(recordIdRule, recordIdItem)
	}
	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _AttestationRegistry.contract.FilterLogs(opts, "VerificationFulfilled", requestIdRule, recordIdRule, oracleRule)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryVerificationFulfilledIterator{contract: _AttestationRegistry.contract, event: "VerificationFulfilled", logs: logs, sub: sub}, nil
}

// WatchVerificationFulfilled is a free log subscription operation binding the contract event 0x97331a6b900e788f1ad34c7be03e9af54f43dcb9c49a359e4ac5723bea6a5fb8.
//
// Solidity: event VerificationFulfilled(uint256 indexed requestId, uint256 indexed recordId, address indexed oracle, uint8 status, bool stale)
func (_AttestationRegistry *AttestationRegistryFilterer) WatchVerificationFulfilled(opts *bind.WatchOpts, sink chan<- *AttestationRegistryVerificationFulfilled, requestId []*big.Int, recordId []*big.Int, oracle []common.Address) (event.Subscription, error) {

	var requestIdRule []interface{}
	for _, requestIdItem := range requestId {
		requestIdRule = append(requestIdRule, requestIdItem)
	}
	var recordIdRule []interface{}
	for _, recordIdItem := range recordId {
		recordIdRule = append(recordIdRule, recordIdItem)
	}
	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _AttestationRegistry.contract.WatchLogs(opts, "VerificationFulfilled", requestIdRule, recordIdRule, oracleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationRegistryVerificationFulfilled)
				if err := _AttestationRegistry.contract.UnpackLog(event, "VerificationFulfilled", log); err != nil {
					// If the signature doesn't match, skip this log.
					if errors.Is(err, bind.ErrEventSignatureMismatch) {
						continue
					}
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVerificationFulfilled is a log parse operation binding the contract event 0x97331a6b900e788f1ad34c7be03e9af54f43dcb9c49a359e4ac5723bea6a5fb8.
//
// Solidity: event VerificationFulfilled(uint256 indexed requestId, uint256 indexed recordId, address indexed oracle, uint8 status, bool stale)
func (_AttestationRegistry *AttestationRegistryFilterer) ParseVerificationFulfilled(log types.Log) (*AttestationRegistryVerificationFulfilled, error) {
	event := new(AttestationRegistryVerificationFulfilled)
	if err := _AttestationRegistry.contract.UnpackLog(event, "VerificationFulfilled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationRegistryVerificationRequestedIterator is returned from FilterVerificationRequested and is used to iterate over the raw logs and unpacked data for VerificationRequested events raised by the AttestationRegistry contract.
type AttestationRegistryVerificationRequestedIterator struct {
	Event *AttestationRegistryVerificationRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationRegistryVerificationRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationRegistryVerificationRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationRegistryVerificationRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationRegistryVerificationRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationRegistryVerificationRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationRegistryVerificationRequested represents a VerificationRequested event raised by the AttestationRegistry contract.
type AttestationRegistryVerificationRequested struct {
	RequestId *big.Int
	RecordId  *big.Int
	Requester common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterVerificationRequested is a free log retrieval operation binding the contract event 0x764966e44b6e46a4f574434e516067d609f891215d196749582e3bc43eb4d08c.
//
// Solidity: event VerificationRequested(uint256 indexed requestId, uint256 indexed recordId, address indexed requester)
func (_AttestationRegistry *AttestationRegistryFilterer) FilterVerificationRequested(opts *bind.FilterOpts, requestId []*big.Int, recordId []*big.Int, requester []common.Address) (*AttestationRegistryVerificationRequestedIterator, error) {

	var requestIdRule []interface{}
	for _, requestIdItem := range requestId {
		requestIdRule = append(requestIdRule, requestIdItem)
	}
	var recordIdRule []interface{}
	for _, recordIdItem := range recordId {
		recordIdRule = append(recordIdRule, recordIdItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _AttestationRegistry.contract.FilterLogs(opts, "VerificationRequested", requestIdRule, recordIdRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return &AttestationRegistryVerificationRequestedIterator{contract: _AttestationRegistry.contract, event: "VerificationRequested", logs: logs, sub: sub}, nil
}

// WatchVerificationRequested is a free log subscription operation binding the contract event 0x764966e44b6e46a4f574434e516067d609f891215d196749582e3bc43eb4d08c.
//
// Solidity: event VerificationRequested(uint256 indexed requestId, uint256 indexed recordId, address indexed requester)
func (_AttestationRegistry *AttestationRegistryFilterer) WatchVerificationRequested(opts *bind.WatchOpts, sink chan<- *AttestationRegistryVerificationRequested, requestId []*big.Int, recordId []*big.Int, requester []common.Address) (event.Subscription, error) {

	var requestIdRule []interface{}
	for _, requestIdItem := range requestId {
		requestIdRule = append(requestIdRule, requestIdItem)
	}
	var recordIdRule []interface{}
	for _, recordIdItem := range recordId {
		recordIdRule = append(recordIdRule, recordIdItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _AttestationRegistry.contract.WatchLogs(opts, "VerificationRequested", requestIdRule, recordIdRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationRegistryVerificationRequested)
				if err := _AttestationRegistry.contract.UnpackLog(event, "VerificationRequested", log); err != nil {
					// If the signature doesn't match, skip this log.
					if errors.Is(err, bind.ErrEventSignatureMismatch) {
						continue
					}
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVerificationRequested is a log parse operation binding the contract event 0x764966e44b6e46a4f574434e516067d609f891215d196749582e3bc43eb4d08c.
//
// Solidity: event VerificationRequested(uint256 indexed requestId, uint256 indexed recordId, address indexed requester)
func (_AttestationRegistry *AttestationRegistryFilterer) ParseVerificationRequested(log types.Log) (*AttestationRegistryVerificationRequested, error) {
	event := new(AttestationRegistryVerificationRequested)
	if err := _AttestationRegistry.contract.UnpackLog(event, "VerificationRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
