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

// AttestationBadgeMetaData contains all meta data concerning the AttestationBadge contract.
var AttestationBadgeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"minter_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"}],\"name\":\"BadgeAlreadyMinted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BadgeNonTransferable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC721EnumerableForbiddenBatchMint\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ERC721IncorrectOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ERC721InsufficientApproval\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"approver\",\"type\":\"address\"}],\"name\":\"ERC721InvalidApprover\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"ERC721InvalidOperator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"ERC721InvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"}],\"name\":\"ERC721InvalidReceiver\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"ERC721InvalidSender\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ERC721NonexistentToken\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"ERC721OutOfBoundsIndex\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidRecord\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MinterRequired\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"UnauthorizedMinter\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"badgeOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"recordId\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minter\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"recordOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"tokenByIndex\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"tokenOfOwnerByIndex\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60a060405234801561000f575f5ffd5b5060405161174e38038061174e83398101604081905261002e916100cd565b604051806040016040528060118152602001704174746573746174696f6e20426164676560781b815250604051806040016040528060068152602001651055151154d560d21b815250815f90816100859190610192565b5060016100928282610192565b5050506001600160a01b0381166100bc576040516334f09c8560e01b815260040160405180910390fd5b6001600160a01b031660805261024c565b5f602082840312156100dd575f5ffd5b81516001600160a01b03811681146100f3575f5ffd5b9392505050565b634e487b7160e01b5f52604160045260245ffd5b600181811c9082168061012257607f821691505b60208210810361014057634e487b7160e01b5f52602260045260245ffd5b50919050565b601f82111561018d57805f5260205f20601f840160051c8101602085101561016b5750805b601f840160051c820191505b8181101561018a575f8155600101610177565b50505b505050565b81516001600160401b038111156101ab576101ab6100fa565b6101bf816101b9845461010e565b84610146565b6020601f8211600181146101f1575f83156101da5750848201515b5f19600385901b1c1916600184901b17845561018a565b5f84815260208120601f198516915b828110156102205787850151825560209485019460019092019101610200565b508482101561023d57868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b6080516114e361026b5f395f818161016d01526104ef01526114e35ff3fe608060405234801561000f575f5ffd5b5060043610610127575f3560e01c806342842e0e116100a957806395d89b411161006e57806395d89b41146102af578063a22cb465146102b7578063b88d4fde146102ca578063c87b56dd146102dd578063e985e9c5146102f0575f5ffd5b806342842e0e146102445780634f6ccce71461025757806355d742e41461026a5780636352211e1461028957806370a082311461029c575f5ffd5b806318160ddd116100ef57806318160ddd146101cf57806323b872dd146101e15780632f745c59146101f45780633ca8fb631461020757806340c10f1914610231575f5ffd5b806301ffc9a71461012b57806306fdde03146101535780630754617214610168578063081812fc146101a7578063095ea7b3146101ba575b5f5ffd5b61013e610139366004611129565b610303565b60405190151581526020015b60405180910390f35b61015b61032d565b60405161014a9190611172565b61018f7f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b03909116815260200161014a565b61018f6101b5366004611184565b6103bc565b6101cd6101c83660046111b6565b6103e3565b005b6008545b60405190815260200161014a565b6101cd6101ef3660046111de565b6103f2565b6101d36102023660046111b6565b610480565b6101d36102153660046111b6565b600b60209081525f928352604080842090915290825290205481565b6101d361023f3660046111b6565b6104e3565b6101cd6102523660046111de565b6105f3565b6101d3610265366004611184565b610612565b6101d3610278366004611184565b600c6020525f908152604090205481565b61018f610297366004611184565b610667565b6101d36102aa366004611218565b610671565b61015b6106b6565b6101cd6102c5366004611231565b6106c5565b6101cd6102d836600461127e565b6106d0565b61015b6102eb366004611184565b6106e8565b61013e6102fe36600461135b565b610759565b5f6001600160e01b0319821663780e9d6360e01b1480610327575061032782610786565b92915050565b60605f805461033b9061138c565b80601f01602080910402602001604051908101604052809291908181526020018280546103679061138c565b80156103b25780601f10610389576101008083540402835291602001916103b2565b820191905f5260205f20905b81548152906001019060200180831161039557829003601f168201915b5050505050905090565b5f6103c6826107d5565b505f828152600460205260409020546001600160a01b0316610327565b6103ee82823361080d565b5050565b6001600160a01b03821661042057604051633250574960e11b81525f60048201526024015b60405180910390fd5b5f61042c83833361081a565b9050836001600160a01b0316816001600160a01b03161461047a576040516364283d7b60e01b81526001600160a01b0380861660048301526024820184905282166044820152606401610417565b50505050565b5f61048a83610671565b82106104bb5760405163295f44f760e21b81526001600160a01b038416600482015260248101839052604401610417565b506001600160a01b03919091165f908152600660209081526040808320938352929052205490565b5f336001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000161461052f5760405163a4f3966560e01b8152336004820152602401610417565b815f0361054f5760405163f2a3e8db60e01b815260040160405180910390fd5b6001600160a01b0383165f908152600b60209081526040808320858452909152902054156105a257604051634a5f41a760e11b81526001600160a01b038416600482015260248101839052604401610417565b600a5f81546105b0906113d8565b91829055506001600160a01b0384165f908152600b602090815260408083208684528252808320849055838352600c909152902083905590506103278382610862565b61060d83838360405180602001604052805f8152506106d0565b505050565b5f61061c60085490565b82106106445760405163295f44f760e21b81525f600482015260248101839052604401610417565b60088281548110610657576106576113f0565b905f5260205f2001549050919050565b5f610327826107d5565b5f6001600160a01b03821661069b576040516322718ad960e21b81525f6004820152602401610417565b506001600160a01b03165f9081526003602052604090205490565b60606001805461033b9061138c565b6103ee3383836108c3565b6106db8484846103f2565b61047a338585858561098a565b60606106f3826107d5565b505f61070960408051602081019091525f815290565b90505f8151116107275760405180602001604052805f815250610752565b8061073184610ab2565b60405160200161074292919061141b565b6040516020818303038152906040525b9392505050565b6001600160a01b039182165f90815260056020908152604080832093909416825291909152205460ff1690565b5f6001600160e01b031982166380ac58cd60e01b14806107b657506001600160e01b03198216635b5e139f60e01b145b8061032757506301ffc9a760e01b6001600160e01b0319831614610327565b5f818152600260205260408120546001600160a01b03168061032757604051637e27328960e01b815260048101849052602401610417565b61060d8383836001610b42565b5f828152600260205260408120546001600160a01b03161561084f57604051635672518360e11b815260040160405180910390fd5b61085a848484610c46565b949350505050565b6001600160a01b03821661088b57604051633250574960e11b81525f6004820152602401610417565b5f61089783835f61081a565b90506001600160a01b0381161561060d576040516339e3563760e11b81525f6004820152602401610417565b6001600160a01b0383166108ec5760405163a9fbf51f60e01b81525f6004820152602401610417565b6001600160a01b03821661091e57604051630b61174360e31b81526001600160a01b0383166004820152602401610417565b6001600160a01b038381165f81815260056020908152604080832094871680845294825291829020805460ff191686151590811790915591519182527f17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31910160405180910390a3505050565b6001600160a01b0383163b15610aab57604051630a85bd0160e11b81526001600160a01b0384169063150b7a02906109cc90889088908790879060040161142f565b6020604051808303815f875af1925050508015610a06575060408051601f3d908101601f19168201909252610a039181019061146b565b60015b610a6d573d808015610a33576040519150601f19603f3d011682016040523d82523d5f602084013e610a38565b606091505b5080515f03610a6557604051633250574960e11b81526001600160a01b0385166004820152602401610417565b805160208201fd5b6001600160e01b03198116630a85bd0160e11b14610aa957604051633250574960e11b81526001600160a01b0385166004820152602401610417565b505b5050505050565b60605f610abe83610d11565b60010190505f8167ffffffffffffffff811115610add57610add61126a565b6040519080825280601f01601f191660200182016040528015610b07576020820181803683370190505b5090508181016020015b5f19016f181899199a1a9b1b9c1cb0b131b232b360811b600a86061a8153600a8504945084610b1157509392505050565b8080610b5657506001600160a01b03821615155b15610c17575f610b65846107d5565b90506001600160a01b03831615801590610b915750826001600160a01b0316816001600160a01b031614155b8015610ba45750610ba28184610759565b155b15610bcd5760405163a9fbf51f60e01b81526001600160a01b0384166004820152602401610417565b8115610c155783856001600160a01b0316826001600160a01b03167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92560405160405180910390a45b505b50505f90815260046020526040902080546001600160a01b0319166001600160a01b0392909216919091179055565b5f5f610c53858585610de8565b90506001600160a01b038116610caf57610caa84600880545f838152600960205260408120829055600182018355919091527ff3f7a9fe364faab93b216da50a3214154f22a0a2b415b23a84c8169e8b636ee30155565b610cd2565b846001600160a01b0316816001600160a01b031614610cd257610cd28185610eda565b6001600160a01b038516610cee57610ce984610f57565b61085a565b846001600160a01b0316816001600160a01b03161461085a5761085a8585610ffe565b5f8072184f03e93ff9f4daa797ed6e38ed64bf6a1f0160401b8310610d4f5772184f03e93ff9f4daa797ed6e38ed64bf6a1f0160401b830492506040015b6d04ee2d6d415b85acef81000000008310610d7b576d04ee2d6d415b85acef8100000000830492506020015b662386f26fc100008310610d9957662386f26fc10000830492506010015b6305f5e1008310610db1576305f5e100830492506008015b6127108310610dc557612710830492506004015b60648310610dd7576064830492506002015b600a83106103275760010192915050565b5f828152600260205260408120546001600160a01b0390811690831615610e1457610e1481848661104c565b6001600160a01b03811615610e4e57610e2f5f855f5f610b42565b6001600160a01b0381165f90815260036020526040902080545f190190555b6001600160a01b03851615610e7c576001600160a01b0385165f908152600360205260409020805460010190555b5f8481526002602052604080822080546001600160a01b0319166001600160a01b0389811691821790925591518793918516917fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef91a4949350505050565b5f610ee483610671565b5f838152600760209081526040808320546001600160a01b0388168452600690925290912091925090818314610f39575f83815260208281526040808320548584528184208190558352600790915290208290555b5f938452600760209081526040808620869055938552525081205550565b6008545f90610f6890600190611486565b5f8381526009602052604081205460088054939450909284908110610f8f57610f8f6113f0565b905f5260205f20015490508060088381548110610fae57610fae6113f0565b5f918252602080832090910192909255828152600990915260408082208490558582528120556008805480610fe557610fe5611499565b600190038181905f5260205f20015f9055905550505050565b5f600161100a84610671565b6110149190611486565b6001600160a01b039093165f908152600660209081526040808320868452825280832085905593825260079052919091209190915550565b6110578383836110b0565b61060d576001600160a01b03831661108557604051637e27328960e01b815260048101829052602401610417565b60405163177e802f60e01b81526001600160a01b038316600482015260248101829052604401610417565b5f6001600160a01b0383161580159061085a5750826001600160a01b0316846001600160a01b031614806110e957506110e98484610759565b8061085a5750505f908152600460205260409020546001600160a01b03908116911614919050565b6001600160e01b031981168114611126575f5ffd5b50565b5f60208284031215611139575f5ffd5b813561075281611111565b5f81518084528060208401602086015e5f602082860101526020601f19601f83011685010191505092915050565b602081525f6107526020830184611144565b5f60208284031215611194575f5ffd5b5035919050565b80356001600160a01b03811681146111b1575f5ffd5b919050565b5f5f604083850312156111c7575f5ffd5b6111d08361119b565b946020939093013593505050565b5f5f5f606084860312156111f0575f5ffd5b6111f98461119b565b92506112076020850161119b565b929592945050506040919091013590565b5f60208284031215611228575f5ffd5b6107528261119b565b5f5f60408385031215611242575f5ffd5b61124b8361119b565b91506020830135801515811461125f575f5ffd5b809150509250929050565b634e487b7160e01b5f52604160045260245ffd5b5f5f5f5f60808587031215611291575f5ffd5b61129a8561119b565b93506112a86020860161119b565b925060408501359150606085013567ffffffffffffffff8111156112ca575f5ffd5b8501601f810187136112da575f5ffd5b803567ffffffffffffffff8111156112f4576112f461126a565b604051601f8201601f19908116603f0116810167ffffffffffffffff811182821017156113235761132361126a565b60405281815282820160200189101561133a575f5ffd5b816020840160208301375f6020838301015280935050505092959194509250565b5f5f6040838503121561136c575f5ffd5b6113758361119b565b91506113836020840161119b565b90509250929050565b600181811c908216806113a057607f821691505b6020821081036113be57634e487b7160e01b5f52602260045260245ffd5b50919050565b634e487b7160e01b5f52601160045260245ffd5b5f600182016113e9576113e96113c4565b5060010190565b634e487b7160e01b5f52603260045260245ffd5b5f81518060208401855e5f93019283525090919050565b5f61085a6114298386611404565b84611404565b6001600160a01b03858116825284166020820152604081018390526080606082018190525f9061146190830184611144565b9695505050505050565b5f6020828403121561147b575f5ffd5b815161075281611111565b81810381811115610327576103276113c4565b634e487b7160e01b5f52603160045260245ffdfea26469706673582212204b1827898690666826ebe42c64708d4386bf7f105b9a79ac65596ec4b4dd99b564736f6c634300081c0033",
}

// AttestationBadgeABI is the input ABI used to generate the binding from.
// Deprecated: Use AttestationBadgeMetaData.ABI instead.
var AttestationBadgeABI = AttestationBadgeMetaData.ABI

// AttestationBadgeBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AttestationBadgeMetaData.Bin instead.
var AttestationBadgeBin = AttestationBadgeMetaData.Bin

// DeployAttestationBadge deploys a new Ethereum contract, binding an instance of AttestationBadge to it.
func DeployAttestationBadge(auth *bind.TransactOpts, backend bind.ContractBackend, minter_ common.Address) (common.Address, *types.Transaction, *AttestationBadge, error) {
	parsed, err := AttestationBadgeMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AttestationBadgeBin), backend, minter_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AttestationBadge{AttestationBadgeCaller: AttestationBadgeCaller{contract: contract}, AttestationBadgeTransactor: AttestationBadgeTransactor{contract: contract}, AttestationBadgeFilterer: AttestationBadgeFilterer{contract: contract}}, nil
}

// AttestationBadge is an auto generated Go binding around an Ethereum contract.
type AttestationBadge struct {
	AttestationBadgeCaller     // Read-only binding to the contract
	AttestationBadgeTransactor // Write-only binding to the contract
	AttestationBadgeFilterer   // Log filterer for contract events
}

// AttestationBadgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type AttestationBadgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationBadgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AttestationBadgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationBadgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AttestationBadgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationBadgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AttestationBadgeSession struct {
	Contract     *AttestationBadge // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AttestationBadgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AttestationBadgeCallerSession struct {
	Contract *AttestationBadgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// AttestationBadgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AttestationBadgeTransactorSession struct {
	Contract     *AttestationBadgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// AttestationBadgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type AttestationBadgeRaw struct {
	Contract *AttestationBadge // Generic contract binding to access the raw methods on
}

// AttestationBadgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AttestationBadgeCallerRaw struct {
	Contract *AttestationBadgeCaller // Generic read-only contract binding to access the raw methods on
}

// AttestationBadgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AttestationBadgeTransactorRaw struct {
	Contract *AttestationBadgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAttestationBadge creates a new instance of AttestationBadge, bound to a specific deployed contract.
func NewAttestationBadge(address common.Address, backend bind.ContractBackend) (*AttestationBadge, error) {
	contract, err := bindAttestationBadge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AttestationBadge{AttestationBadgeCaller: AttestationBadgeCaller{contract: contract}, AttestationBadgeTransactor: AttestationBadgeTransactor{contract: contract}, AttestationBadgeFilterer: AttestationBadgeFilterer{contract: contract}}, nil
}

// NewAttestationBadgeCaller creates a new read-only instance of AttestationBadge, bound to a specific deployed contract.
func NewAttestationBadgeCaller(address common.Address, caller bind.ContractCaller) (*AttestationBadgeCaller, error) {
	contract, err := bindAttestationBadge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AttestationBadgeCaller{contract: contract}, nil
}

// NewAttestationBadgeTransactor creates a new write-only instance of AttestationBadge, bound to a specific deployed contract.
func NewAttestationBadgeTransactor(address common.Address, transactor bind.ContractTransactor) (*AttestationBadgeTransactor, error) {
	contract, err := bindAttestationBadge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AttestationBadgeTransactor{contract: contract}, nil
}

// NewAttestationBadgeFilterer creates a new log filterer instance of AttestationBadge, bound to a specific deployed contract.
func NewAttestationBadgeFilterer(address common.Address, filterer bind.ContractFilterer) (*AttestationBadgeFilterer, error) {
	contract, err := bindAttestationBadge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AttestationBadgeFilterer{contract: contract}, nil
}

// bindAttestationBadge binds a generic wrapper to an already deployed contract.
func bindAttestationBadge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AttestationBadgeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AttestationBadge *AttestationBadgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AttestationBadge.Contract.AttestationBadgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AttestationBadge *AttestationBadgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationBadge.Contract.AttestationBadgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AttestationBadge *AttestationBadgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AttestationBadge.Contract.AttestationBadgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AttestationBadge *AttestationBadgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AttestationBadge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AttestationBadge *AttestationBadgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationBadge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AttestationBadge *AttestationBadgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AttestationBadge.Contract.contract.Transact(opts, method, params...)
}

// BadgeOf is a free data retrieval call binding the contract method 0x3ca8fb63.
//
// Solidity: function badgeOf(address , uint256 ) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCaller) BadgeOf(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "badgeOf", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BadgeOf is a free data retrieval call binding the contract method 0x3ca8fb63.
//
// Solidity: function badgeOf(address , uint256 ) view returns(uint256)
func (_AttestationBadge *AttestationBadgeSession) BadgeOf(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _AttestationBadge.Contract.BadgeOf(&_AttestationBadge.CallOpts, arg0, arg1)
}

// BadgeOf is a free data retrieval call binding the contract method 0x3ca8fb63.
//
// Solidity: function badgeOf(address , uint256 ) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCallerSession) BadgeOf(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _AttestationBadge.Contract.BadgeOf(&_AttestationBadge.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AttestationBadge *AttestationBadgeSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _AttestationBadge.Contract.BalanceOf(&_AttestationBadge.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _AttestationBadge.Contract.BalanceOf(&_AttestationBadge.CallOpts, owner)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AttestationBadge *AttestationBadgeCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AttestationBadge *AttestationBadgeSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _AttestationBadge.Contract.GetApproved(&_AttestationBadge.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_AttestationBadge *AttestationBadgeCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _AttestationBadge.Contract.GetApproved(&_AttestationBadge.CallOpts, tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AttestationBadge *AttestationBadgeCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AttestationBadge *AttestationBadgeSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _AttestationBadge.Contract.IsApprovedForAll(&_AttestationBadge.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_AttestationBadge *AttestationBadgeCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _AttestationBadge.Contract.IsApprovedForAll(&_AttestationBadge.CallOpts, owner, operator)
}

// Minter is a free data retrieval call binding the contract method 0x07546172.
//
// Solidity: function minter() view returns(address)
func (_AttestationBadge *AttestationBadgeCaller) Minter(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "minter")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Minter is a free data retrieval call binding the contract method 0x07546172.
//
// Solidity: function minter() view returns(address)
func (_AttestationBadge *AttestationBadgeSession) Minter() (common.Address, error) {
	return _AttestationBadge.Contract.Minter(&_AttestationBadge.CallOpts)
}

// Minter is a free data retrieval call binding the contract method 0x07546172.
//
// Solidity: function minter() view returns(address)
func (_AttestationBadge *AttestationBadgeCallerSession) Minter() (common.Address, error) {
	return _AttestationBadge.Contract.Minter(&_AttestationBadge.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AttestationBadge *AttestationBadgeCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AttestationBadge *AttestationBadgeSession) Name() (string, error) {
	return _AttestationBadge.Contract.Name(&_AttestationBadge.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AttestationBadge *AttestationBadgeCallerSession) Name() (string, error) {
	return _AttestationBadge.Contract.Name(&_AttestationBadge.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AttestationBadge *AttestationBadgeCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AttestationBadge *AttestationBadgeSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _AttestationBadge.Contract.OwnerOf(&_AttestationBadge.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_AttestationBadge *AttestationBadgeCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _AttestationBadge.Contract.OwnerOf(&_AttestationBadge.CallOpts, tokenId)
}

// RecordOf is a free data retrieval call binding the contract method 0x55d742e4.
//
// Solidity: function recordOf(uint256 ) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCaller) RecordOf(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "recordOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RecordOf is a free data retrieval call binding the contract method 0x55d742e4.
//
// Solidity: function recordOf(uint256 ) view returns(uint256)
func (_AttestationBadge *AttestationBadgeSession) RecordOf(arg0 *big.Int) (*big.Int, error) {
	return _AttestationBadge.Contract.RecordOf(&_AttestationBadge.CallOpts, arg0)
}

// RecordOf is a free data retrieval call binding the contract method 0x55d742e4.
//
// Solidity: function recordOf(uint256 ) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCallerSession) RecordOf(arg0 *big.Int) (*big.Int, error) {
	return _AttestationBadge.Contract.RecordOf(&_AttestationBadge.CallOpts, arg0)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AttestationBadge *AttestationBadgeCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AttestationBadge *AttestationBadgeSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AttestationBadge.Contract.SupportsInterface(&_AttestationBadge.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AttestationBadge *AttestationBadgeCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AttestationBadge.Contract.SupportsInterface(&_AttestationBadge.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AttestationBadge *AttestationBadgeCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AttestationBadge *AttestationBadgeSession) Symbol() (string, error) {
	return _AttestationBadge.Contract.Symbol(&_AttestationBadge.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AttestationBadge *AttestationBadgeCallerSession) Symbol() (string, error) {
	return _AttestationBadge.Contract.Symbol(&_AttestationBadge.CallOpts)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(uint256 index) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCaller) TokenByIndex(opts *bind.CallOpts, index *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "tokenByIndex", index)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(uint256 index) view returns(uint256)
func (_AttestationBadge *AttestationBadgeSession) TokenByIndex(index *big.Int) (*big.Int, error) {
	return _AttestationBadge.Contract.TokenByIndex(&_AttestationBadge.CallOpts, index)
}

// TokenByIndex is a free data retrieval call binding the contract method 0x4f6ccce7.
//
// Solidity: function tokenByIndex(uint256 index) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCallerSession) TokenByIndex(index *big.Int) (*big.Int, error) {
	return _AttestationBadge.Contract.TokenByIndex(&_AttestationBadge.CallOpts, index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(address owner, uint256 index) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCaller) TokenOfOwnerByIndex(opts *bind.CallOpts, owner common.Address, index *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "tokenOfOwnerByIndex", owner, index)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(address owner, uint256 index) view returns(uint256)
func (_AttestationBadge *AttestationBadgeSession) TokenOfOwnerByIndex(owner common.Address, index *big.Int) (*big.Int, error) {
	return _AttestationBadge.Contract.TokenOfOwnerByIndex(&_AttestationBadge.CallOpts, owner, index)
}

// TokenOfOwnerByIndex is a free data retrieval call binding the contract method 0x2f745c59.
//
// Solidity: function tokenOfOwnerByIndex(address owner, uint256 index) view returns(uint256)
func (_AttestationBadge *AttestationBadgeCallerSession) TokenOfOwnerByIndex(owner common.Address, index *big.Int) (*big.Int, error) {
	return _AttestationBadge.Contract.TokenOfOwnerByIndex(&_AttestationBadge.CallOpts, owner, index)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AttestationBadge *AttestationBadgeCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AttestationBadge *AttestationBadgeSession) TokenURI(tokenId *big.Int) (string, error) {
	return _AttestationBadge.Contract.TokenURI(&_AttestationBadge.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_AttestationBadge *AttestationBadgeCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _AttestationBadge.Contract.TokenURI(&_AttestationBadge.CallOpts, tokenId)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_AttestationBadge *AttestationBadgeCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationBadge.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_AttestationBadge *AttestationBadgeSession) TotalSupply() (*big.Int, error) {
	return _AttestationBadge.Contract.TotalSupply(&_AttestationBadge.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_AttestationBadge *AttestationBadgeCallerSession) TotalSupply() (*big.Int, error) {
	return _AttestationBadge.Contract.TotalSupply(&_AttestationBadge.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.Contract.Approve(&_AttestationBadge.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.Contract.Approve(&_AttestationBadge.TransactOpts, to, tokenId)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 recordId) returns(uint256 tokenId)
func (_AttestationBadge *AttestationBadgeTransactor) Mint(opts *bind.TransactOpts, to common.Address, recordId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.contract.Transact(opts, "mint", to, recordId)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 recordId) returns(uint256 tokenId)
func (_AttestationBadge *AttestationBadgeSession) Mint(to common.Address, recordId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.Contract.Mint(&_AttestationBadge.TransactOpts, to, recordId)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 recordId) returns(uint256 tokenId)
func (_AttestationBadge *AttestationBadgeTransactorSession) Mint(to common.Address, recordId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.Contract.Mint(&_AttestationBadge.TransactOpts, to, recordId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.Contract.SafeTransferFrom(&_AttestationBadge.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.Contract.SafeTransferFrom(&_AttestationBadge.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AttestationBadge *AttestationBadgeTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AttestationBadge.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AttestationBadge *AttestationBadgeSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AttestationBadge.Contract.SafeTransferFrom0(&_AttestationBadge.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_AttestationBadge *AttestationBadgeTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _AttestationBadge.Contract.SafeTransferFrom0(&_AttestationBadge.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AttestationBadge *AttestationBadgeTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _AttestationBadge.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AttestationBadge *AttestationBadgeSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AttestationBadge.Contract.SetApprovalForAll(&_AttestationBadge.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_AttestationBadge *AttestationBadgeTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _AttestationBadge.Contract.SetApprovalForAll(&_AttestationBadge.TransactOpts, operator, approved)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.Contract.TransferFrom(&_AttestationBadge.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_AttestationBadge *AttestationBadgeTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _AttestationBadge.Contract.TransferFrom(&_AttestationBadge.TransactOpts, from, to, tokenId)
}

// AttestationBadgeApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the AttestationBadge contract.
type AttestationBadgeApprovalIterator struct {
	Event *AttestationBadgeApproval // Event containing the contract specifics and raw log

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
func (it *AttestationBadgeApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationBadgeApproval)
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
		it.Event = new(AttestationBadgeApproval)
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
func (it *AttestationBadgeApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationBadgeApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationBadgeApproval represents a Approval event raised by the AttestationBadge contract.
type AttestationBadgeApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AttestationBadge *AttestationBadgeFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*AttestationBadgeApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AttestationBadge.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AttestationBadgeApprovalIterator{contract: _AttestationBadge.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AttestationBadge *AttestationBadgeFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *AttestationBadgeApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AttestationBadge.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationBadgeApproval)
				if err := _AttestationBadge.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_AttestationBadge *AttestationBadgeFilterer) ParseApproval(log types.Log) (*AttestationBadgeApproval, error) {
	event := new(AttestationBadgeApproval)
	if err := _AttestationBadge.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationBadgeApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the AttestationBadge contract.
type AttestationBadgeApprovalForAllIterator struct {
	Event *AttestationBadgeApprovalForAll // Event containing the contract specifics and raw log

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
func (it *AttestationBadgeApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationBadgeApprovalForAll)
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
		it.Event = new(AttestationBadgeApprovalForAll)
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
func (it *AttestationBadgeApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationBadgeApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationBadgeApprovalForAll represents a ApprovalForAll event raised by the AttestationBadge contract.
type AttestationBadgeApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AttestationBadge *AttestationBadgeFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*AttestationBadgeApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AttestationBadge.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &AttestationBadgeApprovalForAllIterator{contract: _AttestationBadge.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AttestationBadge *AttestationBadgeFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *AttestationBadgeApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AttestationBadge.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationBadgeApprovalForAll)
				if err := _AttestationBadge.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_AttestationBadge *AttestationBadgeFilterer) ParseApprovalForAll(log types.Log) (*AttestationBadgeApprovalForAll, error) {
	event := new(AttestationBadgeApprovalForAll)
	if err := _AttestationBadge.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationBadgeTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the AttestationBadge contract.
type AttestationBadgeTransferIterator struct {
	Event *AttestationBadgeTransfer // Event containing the contract specifics and raw log

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
func (it *AttestationBadgeTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationBadgeTransfer)
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
		it.Event = new(AttestationBadgeTransfer)
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
func (it *AttestationBadgeTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationBadgeTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationBadgeTransfer represents a Transfer event raised by the AttestationBadge contract.
type AttestationBadgeTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AttestationBadge *AttestationBadgeFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*AttestationBadgeTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AttestationBadge.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AttestationBadgeTransferIterator{contract: _AttestationBadge.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AttestationBadge *AttestationBadgeFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *AttestationBadgeTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _AttestationBadge.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationBadgeTransfer)
				if err := _AttestationBadge.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_AttestationBadge *AttestationBadgeFilterer) ParseTransfer(log types.Log) (*AttestationBadgeTransfer, error) {
	event := new(AttestationBadgeTransfer)
	if err := _AttestationBadge.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
