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

// BaselineRegistryMetaData contains all meta data concerning the BaselineRegistry contract.
var BaselineRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"writer_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"UnauthorizedWriter\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WriterRequired\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"}],\"name\":\"HashesRecorded\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"artifactHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"sbomHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"scanHash\",\"type\":\"bytes32\"}],\"name\":\"addHashes\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"artifactHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"sbomHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"scanHash\",\"type\":\"bytes32\"}],\"name\":\"hasHashes\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"writer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60a0604052348015600e575f5ffd5b506040516102d13803806102d1833981016040819052602b916061565b6001600160a01b0381166051576040516372196abb60e11b815260040160405180910390fd5b6001600160a01b0316608052608c565b5f602082840312156070575f5ffd5b81516001600160a01b03811681146085575f5ffd5b9392505050565b6080516102276100aa5f395f81816048015261010a01526102275ff3fe608060405234801561000f575f5ffd5b506004361061003f575f3560e01c8063453a2abc1461004357806380e33eaa14610087578063ac1b08501461009c575b5f5ffd5b61006a7f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020015b60405180910390f35b61009a6100953660046101c8565b6100ff565b005b6100ef6100aa3660046101c8565b604080516020808201959095528082019390935260608084019290925280518084039092018252608090920182528051908301205f9081529182905290205460ff1690565b604051901515815260200161007e565b336001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000161461014e576040516340f0d1e360e11b815233600482015260240160405180910390fd5b6040805160208101859052908101839052606081018290525f9060800160408051601f1981840301815282825280516020918201205f81815291829052918120805460ff1916600117905590925082917f793c72a4906951ada93a35c39b9a3dca5e411ac8578efef8393942516273d14a9190a250505050565b5f5f5f606084860312156101da575f5ffd5b50508135936020830135935060409092013591905056fea2646970667358221220b4c831df985a917a75de29e274b72ba8e8da267f7de8eee7b32d0c6e2369c3f064736f6c634300081c0033",
}

// BaselineRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use BaselineRegistryMetaData.ABI instead.
var BaselineRegistryABI = BaselineRegistryMetaData.ABI

// BaselineRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use BaselineRegistryMetaData.Bin instead.
var BaselineRegistryBin = BaselineRegistryMetaData.Bin

// DeployBaselineRegistry deploys a new Ethereum contract, binding an instance of BaselineRegistry to it.
func DeployBaselineRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, writer_ common.Address) (common.Address, *types.Transaction, *BaselineRegistry, error) {
	parsed, err := BaselineRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(BaselineRegistryBin), backend, writer_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &BaselineRegistry{BaselineRegistryCaller: BaselineRegistryCaller{contract: contract}, BaselineRegistryTransactor: BaselineRegistryTransactor{contract: contract}, BaselineRegistryFilterer: BaselineRegistryFilterer{contract: contract}}, nil
}

// BaselineRegistry is an auto generated Go binding around an Ethereum contract.
type BaselineRegistry struct {
	BaselineRegistryCaller     // Read-only binding to the contract
	BaselineRegistryTransactor // Write-only binding to the contract
	BaselineRegistryFilterer   // Log filterer for contract events
}

// BaselineRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type BaselineRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BaselineRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BaselineRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BaselineRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BaselineRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BaselineRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BaselineRegistrySession struct {
	Contract     *BaselineRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BaselineRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BaselineRegistryCallerSession struct {
	Contract *BaselineRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// BaselineRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BaselineRegistryTransactorSession struct {
	Contract     *BaselineRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// BaselineRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type BaselineRegistryRaw struct {
	Contract *BaselineRegistry // Generic contract binding to access the raw methods on
}

// BaselineRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BaselineRegistryCallerRaw struct {
	Contract *BaselineRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// BaselineRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BaselineRegistryTransactorRaw struct {
	Contract *BaselineRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBaselineRegistry creates a new instance of BaselineRegistry, bound to a specific deployed contract.
func NewBaselineRegistry(address common.Address, backend bind.ContractBackend) (*BaselineRegistry, error) {
	contract, err := bindBaselineRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BaselineRegistry{BaselineRegistryCaller: BaselineRegistryCaller{contract: contract}, BaselineRegistryTransactor: BaselineRegistryTransactor{contract: contract}, BaselineRegistryFilterer: BaselineRegistryFilterer{contract: contract}}, nil
}

// NewBaselineRegistryCaller creates a new read-only instance of BaselineRegistry, bound to a specific deployed contract.
func NewBaselineRegistryCaller(address common.Address, caller bind.ContractCaller) (*BaselineRegistryCaller, error) {
	contract, err := bindBaselineRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BaselineRegistryCaller{contract: contract}, nil
}

// NewBaselineRegistryTransactor creates a new write-only instance of BaselineRegistry, bound to a specific deployed contract.
func NewBaselineRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*BaselineRegistryTransactor, error) {
	contract, err := bindBaselineRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BaselineRegistryTransactor{contract: contract}, nil
}

// NewBaselineRegistryFilterer creates a new log filterer instance of BaselineRegistry, bound to a specific deployed contract.
func NewBaselineRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*BaselineRegistryFilterer, error) {
	contract, err := bindBaselineRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BaselineRegistryFilterer{contract: contract}, nil
}

// bindBaselineRegistry binds a generic wrapper to an already deployed contract.
func bindBaselineRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BaselineRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BaselineRegistry *BaselineRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BaselineRegistry.Contract.BaselineRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BaselineRegistry *BaselineRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BaselineRegistry.Contract.BaselineRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BaselineRegistry *BaselineRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BaselineRegistry.Contract.BaselineRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BaselineRegistry *BaselineRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BaselineRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BaselineRegistry *BaselineRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BaselineRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BaselineRegistry *BaselineRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BaselineRegistry.Contract.contract.Transact(opts, method, params...)
}

// HasHashes is a free data retrieval call binding the contract method 0xac1b0850.
//
// Solidity: function hasHashes(bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) view returns(bool)
func (_BaselineRegistry *BaselineRegistryCaller) HasHashes(opts *bind.CallOpts, artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (bool, error) {
	var out []interface{}
	err := _BaselineRegistry.contract.Call(opts, &out, "hasHashes", artifactHash, sbomHash, scanHash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasHashes is a free data retrieval call binding the contract method 0xac1b0850.
//
// Solidity: function hasHashes(bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) view returns(bool)
func (_BaselineRegistry *BaselineRegistrySession) HasHashes(artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (bool, error) {
	return _BaselineRegistry.Contract.HasHashes(&_BaselineRegistry.CallOpts, artifactHash, sbomHash, scanHash)
}

// HasHashes is a free data retrieval call binding the contract method 0xac1b0850.
//
// Solidity: function hasHashes(bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) view returns(bool)
func (_BaselineRegistry *BaselineRegistryCallerSession) HasHashes(artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (bool, error) {
	return _BaselineRegistry.Contract.HasHashes(&_BaselineRegistry.CallOpts, artifactHash, sbomHash, scanHash)
}

// Writer is a free data retrieval call binding the contract method 0x453a2abc.
//
// Solidity: function writer() view returns(address)
func (_BaselineRegistry *BaselineRegistryCaller) Writer(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BaselineRegistry.contract.Call(opts, &out, "writer")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Writer is a free data retrieval call binding the contract method 0x453a2abc.
//
// Solidity: function writer() view returns(address)
func (_BaselineRegistry *BaselineRegistrySession) Writer() (common.Address, error) {
	return _BaselineRegistry.Contract.Writer(&_BaselineRegistry.CallOpts)
}

// Writer is a free data retrieval call binding the contract method 0x453a2abc.
//
// Solidity: function writer() view returns(address)
func (_BaselineRegistry *BaselineRegistryCallerSession) Writer() (common.Address, error) {
	return _BaselineRegistry.Contract.Writer(&_BaselineRegistry.CallOpts)
}

// AddHashes is a paid mutator transaction binding the contract method 0x80e33eaa.
//
// Solidity: function addHashes(bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) returns()
func (_BaselineRegistry *BaselineRegistryTransactor) AddHashes(opts *bind.TransactOpts, artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (*types.Transaction, error) {
	return _BaselineRegistry.contract.Transact(opts, "addHashes", artifactHash, sbomHash, scanHash)
}

// AddHashes is a paid mutator transaction binding the contract method 0x80e33eaa.
//
// Solidity: function addHashes(bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) returns()
func (_BaselineRegistry *BaselineRegistrySession) AddHashes(artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (*types.Transaction, error) {
	return _BaselineRegistry.Contract.AddHashes(&_BaselineRegistry.TransactOpts, artifactHash, sbomHash, scanHash)
}

// AddHashes is a paid mutator transaction binding the contract method 0x80e33eaa.
//
// Solidity: function addHashes(bytes32 artifactHash, bytes32 sbomHash, bytes32 scanHash) returns()
func (_BaselineRegistry *BaselineRegistryTransactorSession) AddHashes(artifactHash [32]byte, sbomHash [32]byte, scanHash [32]byte) (*types.Transaction, error) {
	return _BaselineRegistry.Contract.AddHashes(&_BaselineRegistry.TransactOpts, artifactHash, sbomHash, scanHash)
}

// BaselineRegistryHashesRecordedIterator is returned from FilterHashesRecorded and is used to iterate over the raw logs and unpacked data for HashesRecorded events raised by the BaselineRegistry contract.
type BaselineRegistryHashesRecordedIterator struct {
	Event *BaselineRegistryHashesRecorded // Event containing the contract specifics and raw log

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
func (it *BaselineRegistryHashesRecordedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BaselineRegistryHashesRecorded)
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
		it.Event = new(BaselineRegistryHashesRecorded)
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
func (it *BaselineRegistryHashesRecordedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BaselineRegistryHashesRecordedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BaselineRegistryHashesRecorded represents a HashesRecorded event raised by the BaselineRegistry contract.
type BaselineRegistryHashesRecorded struct {
	Key [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterHashesRecorded is a free log retrieval operation binding the contract event 0x793c72a4906951ada93a35c39b9a3dca5e411ac8578efef8393942516273d14a.
//
// Solidity: event HashesRecorded(bytes32 indexed key)
func (_BaselineRegistry *BaselineRegistryFilterer) FilterHashesRecorded(opts *bind.FilterOpts, key [][32]byte) (*BaselineRegistryHashesRecordedIterator, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}

	logs, sub, err := _BaselineRegistry.contract.FilterLogs(opts, "HashesRecorded", keyRule)
	if err != nil {
		return nil, err
	}
	return &BaselineRegistryHashesRecordedIterator{contract: _BaselineRegistry.contract, event: "HashesRecorded", logs: logs, sub: sub}, nil
}

// WatchHashesRecorded is a free log subscription operation binding the contract event 0x793c72a4906951ada93a35c39b9a3dca5e411ac8578efef8393942516273d14a.
//
// Solidity: event HashesRecorded(bytes32 indexed key)
func (_BaselineRegistry *BaselineRegistryFilterer) WatchHashesRecorded(opts *bind.WatchOpts, sink chan<- *BaselineRegistryHashesRecorded, key [][32]byte) (event.Subscription, error) {

	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}

	logs, sub, err := _BaselineRegistry.contract.WatchLogs(opts, "HashesRecorded", keyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BaselineRegistryHashesRecorded)
				if err := _BaselineRegistry.contract.UnpackLog(event, "HashesRecorded", log); err != nil {
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

// ParseHashesRecorded is a log parse operation binding the contract event 0x793c72a4906951ada93a35c39b9a3dca5e411ac8578efef8393942516273d14a.
//
// Solidity: event HashesRecorded(bytes32 indexed key)
func (_BaselineRegistry *BaselineRegistryFilterer) ParseHashesRecorded(log types.Log) (*BaselineRegistryHashesRecorded, error) {
	event := new(BaselineRegistryHashesRecorded)
	if err := _BaselineRegistry.contract.UnpackLog(event, "HashesRecorded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
