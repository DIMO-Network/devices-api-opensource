// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package registry

import (
	"errors"
	"math/big"
	"strings"

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
)

// AbiMetaData contains all meta data concerning the Abi contract.
var AbiMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"__baseURI\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"moduleAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"moduleAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldImplementation\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"oldSelectors\",\"type\":\"bytes4[]\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"newSelectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"addModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"removeModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oldImplementation\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"oldSelectors\",\"type\":\"bytes4[]\"},{\"internalType\":\"bytes4[]\",\"name\":\"newSelectors\",\"type\":\"bytes4[]\"}],\"name\":\"updateModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_adMintCost\",\"type\":\"uint256\"}],\"name\":\"setAdMintCost\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_dimoToken\",\"type\":\"address\"}],\"name\":\"setDimoToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_foundation\",\"type\":\"address\"}],\"name\":\"setFoundationAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_license\",\"type\":\"address\"}],\"name\":\"setLicense\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeType\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"aftermarketDeviceAddress\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceNodeMinted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDevicePaired\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeType\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"AttributeAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeType\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"}],\"name\":\"NodeMinted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addAftermarketDeviceAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"ownerSig\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"aftermarketDeviceSig\",\"type\":\"bytes\"}],\"name\":\"claimAftermarketDeviceSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[][]\",\"name\":\"infos\",\"type\":\"string[][]\"}],\"name\":\"mintAftermarketDeviceByManufacturerBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"pairAftermarketDeviceSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"setAftermarketDeviceInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"label\",\"type\":\"bytes\"}],\"name\":\"setAftermarketDeviceNodeType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"ControllerSet\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addManufacturerAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isController\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isController\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isManufacturerMinted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isManufacturerMinted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"mintManufacturer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"names\",\"type\":\"string[]\"}],\"name\":\"mintManufacturerBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_controller\",\"type\":\"address\"}],\"name\":\"setController\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"setManufacturerInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"label\",\"type\":\"bytes\"}],\"name\":\"setManufacturerNodeType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addVehicleAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"mintVehicle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"mintVehicleSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"setVehicleInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"label\",\"type\":\"bytes\"}],\"name\":\"setVehicleNodeType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"getInfo\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getNodeType\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeType\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getParentNode\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"parentNode\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sourceNode\",\"type\":\"uint256\"}],\"name\":\"getLink\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"targetNode\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"baseURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_baseURI\",\"type\":\"string\"}],\"name\":\"setBaseURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_tokenURI\",\"type\":\"string\"}],\"name\":\"setTokenURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// AbiABI is the input ABI used to generate the binding from.
// Deprecated: Use AbiMetaData.ABI instead.
var AbiABI = AbiMetaData.ABI

// Abi is an auto generated Go binding around an Ethereum contract.
type Abi struct {
	AbiCaller     // Read-only binding to the contract
	AbiTransactor // Write-only binding to the contract
	AbiFilterer   // Log filterer for contract events
}

// AbiCaller is an auto generated read-only Go binding around an Ethereum contract.
type AbiCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AbiTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AbiTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AbiFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AbiFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AbiSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AbiSession struct {
	Contract     *Abi              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AbiCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AbiCallerSession struct {
	Contract *AbiCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// AbiTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AbiTransactorSession struct {
	Contract     *AbiTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AbiRaw is an auto generated low-level Go binding around an Ethereum contract.
type AbiRaw struct {
	Contract *Abi // Generic contract binding to access the raw methods on
}

// AbiCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AbiCallerRaw struct {
	Contract *AbiCaller // Generic read-only contract binding to access the raw methods on
}

// AbiTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AbiTransactorRaw struct {
	Contract *AbiTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAbi creates a new instance of Abi, bound to a specific deployed contract.
func NewAbi(address common.Address, backend bind.ContractBackend) (*Abi, error) {
	contract, err := bindAbi(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Abi{AbiCaller: AbiCaller{contract: contract}, AbiTransactor: AbiTransactor{contract: contract}, AbiFilterer: AbiFilterer{contract: contract}}, nil
}

// NewAbiCaller creates a new read-only instance of Abi, bound to a specific deployed contract.
func NewAbiCaller(address common.Address, caller bind.ContractCaller) (*AbiCaller, error) {
	contract, err := bindAbi(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AbiCaller{contract: contract}, nil
}

// NewAbiTransactor creates a new write-only instance of Abi, bound to a specific deployed contract.
func NewAbiTransactor(address common.Address, transactor bind.ContractTransactor) (*AbiTransactor, error) {
	contract, err := bindAbi(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AbiTransactor{contract: contract}, nil
}

// NewAbiFilterer creates a new log filterer instance of Abi, bound to a specific deployed contract.
func NewAbiFilterer(address common.Address, filterer bind.ContractFilterer) (*AbiFilterer, error) {
	contract, err := bindAbi(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AbiFilterer{contract: contract}, nil
}

// bindAbi binds a generic wrapper to an already deployed contract.
func bindAbi(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AbiABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Abi *AbiRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Abi.Contract.AbiCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Abi *AbiRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Abi.Contract.AbiTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Abi *AbiRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Abi.Contract.AbiTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Abi *AbiCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Abi.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Abi *AbiTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Abi.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Abi *AbiTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Abi.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Abi *AbiCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Abi *AbiSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _Abi.Contract.BalanceOf(&_Abi.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Abi *AbiCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _Abi.Contract.BalanceOf(&_Abi.CallOpts, account)
}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_Abi *AbiCaller) BaseURI(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "baseURI")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_Abi *AbiSession) BaseURI() (string, error) {
	return _Abi.Contract.BaseURI(&_Abi.CallOpts)
}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_Abi *AbiCallerSession) BaseURI() (string, error) {
	return _Abi.Contract.BaseURI(&_Abi.CallOpts)
}

// GetInfo is a free data retrieval call binding the contract method 0x5cc148f3.
//
// Solidity: function getInfo(uint256 nodeId, string attribute) view returns(string info)
func (_Abi *AbiCaller) GetInfo(opts *bind.CallOpts, nodeId *big.Int, attribute string) (string, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "getInfo", nodeId, attribute)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetInfo is a free data retrieval call binding the contract method 0x5cc148f3.
//
// Solidity: function getInfo(uint256 nodeId, string attribute) view returns(string info)
func (_Abi *AbiSession) GetInfo(nodeId *big.Int, attribute string) (string, error) {
	return _Abi.Contract.GetInfo(&_Abi.CallOpts, nodeId, attribute)
}

// GetInfo is a free data retrieval call binding the contract method 0x5cc148f3.
//
// Solidity: function getInfo(uint256 nodeId, string attribute) view returns(string info)
func (_Abi *AbiCallerSession) GetInfo(nodeId *big.Int, attribute string) (string, error) {
	return _Abi.Contract.GetInfo(&_Abi.CallOpts, nodeId, attribute)
}

// GetLink is a free data retrieval call binding the contract method 0x393b6df3.
//
// Solidity: function getLink(uint256 sourceNode) view returns(uint256 targetNode)
func (_Abi *AbiCaller) GetLink(opts *bind.CallOpts, sourceNode *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "getLink", sourceNode)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLink is a free data retrieval call binding the contract method 0x393b6df3.
//
// Solidity: function getLink(uint256 sourceNode) view returns(uint256 targetNode)
func (_Abi *AbiSession) GetLink(sourceNode *big.Int) (*big.Int, error) {
	return _Abi.Contract.GetLink(&_Abi.CallOpts, sourceNode)
}

// GetLink is a free data retrieval call binding the contract method 0x393b6df3.
//
// Solidity: function getLink(uint256 sourceNode) view returns(uint256 targetNode)
func (_Abi *AbiCallerSession) GetLink(sourceNode *big.Int) (*big.Int, error) {
	return _Abi.Contract.GetLink(&_Abi.CallOpts, sourceNode)
}

// GetNodeType is a free data retrieval call binding the contract method 0x70c3e13b.
//
// Solidity: function getNodeType(uint256 tokenId) view returns(uint256 nodeType)
func (_Abi *AbiCaller) GetNodeType(opts *bind.CallOpts, tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "getNodeType", tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNodeType is a free data retrieval call binding the contract method 0x70c3e13b.
//
// Solidity: function getNodeType(uint256 tokenId) view returns(uint256 nodeType)
func (_Abi *AbiSession) GetNodeType(tokenId *big.Int) (*big.Int, error) {
	return _Abi.Contract.GetNodeType(&_Abi.CallOpts, tokenId)
}

// GetNodeType is a free data retrieval call binding the contract method 0x70c3e13b.
//
// Solidity: function getNodeType(uint256 tokenId) view returns(uint256 nodeType)
func (_Abi *AbiCallerSession) GetNodeType(tokenId *big.Int) (*big.Int, error) {
	return _Abi.Contract.GetNodeType(&_Abi.CallOpts, tokenId)
}

// GetParentNode is a free data retrieval call binding the contract method 0xc5e80c85.
//
// Solidity: function getParentNode(uint256 tokenId) view returns(uint256 parentNode)
func (_Abi *AbiCaller) GetParentNode(opts *bind.CallOpts, tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "getParentNode", tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetParentNode is a free data retrieval call binding the contract method 0xc5e80c85.
//
// Solidity: function getParentNode(uint256 tokenId) view returns(uint256 parentNode)
func (_Abi *AbiSession) GetParentNode(tokenId *big.Int) (*big.Int, error) {
	return _Abi.Contract.GetParentNode(&_Abi.CallOpts, tokenId)
}

// GetParentNode is a free data retrieval call binding the contract method 0xc5e80c85.
//
// Solidity: function getParentNode(uint256 tokenId) view returns(uint256 parentNode)
func (_Abi *AbiCallerSession) GetParentNode(tokenId *big.Int) (*big.Int, error) {
	return _Abi.Contract.GetParentNode(&_Abi.CallOpts, tokenId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Abi *AbiCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Abi *AbiSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Abi.Contract.GetRoleAdmin(&_Abi.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Abi *AbiCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Abi.Contract.GetRoleAdmin(&_Abi.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Abi *AbiCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Abi *AbiSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Abi.Contract.HasRole(&_Abi.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Abi *AbiCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Abi.Contract.HasRole(&_Abi.CallOpts, role, account)
}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Abi *AbiCaller) IsController(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "isController", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Abi *AbiSession) IsController(addr common.Address) (bool, error) {
	return _Abi.Contract.IsController(&_Abi.CallOpts, addr)
}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Abi *AbiCallerSession) IsController(addr common.Address) (bool, error) {
	return _Abi.Contract.IsController(&_Abi.CallOpts, addr)
}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Abi *AbiCaller) IsManufacturerMinted(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "isManufacturerMinted", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Abi *AbiSession) IsManufacturerMinted(addr common.Address) (bool, error) {
	return _Abi.Contract.IsManufacturerMinted(&_Abi.CallOpts, addr)
}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Abi *AbiCallerSession) IsManufacturerMinted(addr common.Address) (bool, error) {
	return _Abi.Contract.IsManufacturerMinted(&_Abi.CallOpts, addr)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Abi *AbiCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Abi *AbiSession) Name() (string, error) {
	return _Abi.Contract.Name(&_Abi.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Abi *AbiCallerSession) Name() (string, error) {
	return _Abi.Contract.Name(&_Abi.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Abi *AbiCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Abi *AbiSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Abi.Contract.OwnerOf(&_Abi.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Abi *AbiCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Abi.Contract.OwnerOf(&_Abi.CallOpts, tokenId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Abi *AbiCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Abi *AbiSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Abi.Contract.SupportsInterface(&_Abi.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Abi *AbiCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Abi.Contract.SupportsInterface(&_Abi.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Abi *AbiCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Abi *AbiSession) Symbol() (string, error) {
	return _Abi.Contract.Symbol(&_Abi.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Abi *AbiCallerSession) Symbol() (string, error) {
	return _Abi.Contract.Symbol(&_Abi.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Abi *AbiCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Abi *AbiSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Abi.Contract.TokenURI(&_Abi.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Abi *AbiCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Abi.Contract.TokenURI(&_Abi.CallOpts, tokenId)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Abi *AbiCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Abi *AbiSession) TotalSupply() (*big.Int, error) {
	return _Abi.Contract.TotalSupply(&_Abi.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Abi *AbiCallerSession) TotalSupply() (*big.Int, error) {
	return _Abi.Contract.TotalSupply(&_Abi.CallOpts)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Abi *AbiTransactor) AddAftermarketDeviceAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "addAftermarketDeviceAttribute", attribute)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Abi *AbiSession) AddAftermarketDeviceAttribute(attribute string) (*types.Transaction, error) {
	return _Abi.Contract.AddAftermarketDeviceAttribute(&_Abi.TransactOpts, attribute)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Abi *AbiTransactorSession) AddAftermarketDeviceAttribute(attribute string) (*types.Transaction, error) {
	return _Abi.Contract.AddAftermarketDeviceAttribute(&_Abi.TransactOpts, attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Abi *AbiTransactor) AddManufacturerAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "addManufacturerAttribute", attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Abi *AbiSession) AddManufacturerAttribute(attribute string) (*types.Transaction, error) {
	return _Abi.Contract.AddManufacturerAttribute(&_Abi.TransactOpts, attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Abi *AbiTransactorSession) AddManufacturerAttribute(attribute string) (*types.Transaction, error) {
	return _Abi.Contract.AddManufacturerAttribute(&_Abi.TransactOpts, attribute)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Abi *AbiTransactor) AddModule(opts *bind.TransactOpts, implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "addModule", implementation, selectors)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Abi *AbiSession) AddModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Abi.Contract.AddModule(&_Abi.TransactOpts, implementation, selectors)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Abi *AbiTransactorSession) AddModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Abi.Contract.AddModule(&_Abi.TransactOpts, implementation, selectors)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Abi *AbiTransactor) AddVehicleAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "addVehicleAttribute", attribute)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Abi *AbiSession) AddVehicleAttribute(attribute string) (*types.Transaction, error) {
	return _Abi.Contract.AddVehicleAttribute(&_Abi.TransactOpts, attribute)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Abi *AbiTransactorSession) AddVehicleAttribute(attribute string) (*types.Transaction, error) {
	return _Abi.Contract.AddVehicleAttribute(&_Abi.TransactOpts, attribute)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Abi *AbiTransactor) ClaimAftermarketDeviceSign(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "claimAftermarketDeviceSign", aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Abi *AbiSession) ClaimAftermarketDeviceSign(aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Abi.Contract.ClaimAftermarketDeviceSign(&_Abi.TransactOpts, aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Abi *AbiTransactorSession) ClaimAftermarketDeviceSign(aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Abi.Contract.ClaimAftermarketDeviceSign(&_Abi.TransactOpts, aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Abi *AbiTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Abi *AbiSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.Contract.GrantRole(&_Abi.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Abi *AbiTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.Contract.GrantRole(&_Abi.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Abi *AbiTransactor) Initialize(opts *bind.TransactOpts, name string, version string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "initialize", name, version)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Abi *AbiSession) Initialize(name string, version string) (*types.Transaction, error) {
	return _Abi.Contract.Initialize(&_Abi.TransactOpts, name, version)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Abi *AbiTransactorSession) Initialize(name string, version string) (*types.Transaction, error) {
	return _Abi.Contract.Initialize(&_Abi.TransactOpts, name, version)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x6c155f2e.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, address[] addresses, string[] attributes, string[][] infos) returns()
func (_Abi *AbiTransactor) MintAftermarketDeviceByManufacturerBatch(opts *bind.TransactOpts, manufacturerNode *big.Int, addresses []common.Address, attributes []string, infos [][]string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "mintAftermarketDeviceByManufacturerBatch", manufacturerNode, addresses, attributes, infos)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x6c155f2e.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, address[] addresses, string[] attributes, string[][] infos) returns()
func (_Abi *AbiSession) MintAftermarketDeviceByManufacturerBatch(manufacturerNode *big.Int, addresses []common.Address, attributes []string, infos [][]string) (*types.Transaction, error) {
	return _Abi.Contract.MintAftermarketDeviceByManufacturerBatch(&_Abi.TransactOpts, manufacturerNode, addresses, attributes, infos)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x6c155f2e.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, address[] addresses, string[] attributes, string[][] infos) returns()
func (_Abi *AbiTransactorSession) MintAftermarketDeviceByManufacturerBatch(manufacturerNode *big.Int, addresses []common.Address, attributes []string, infos [][]string) (*types.Transaction, error) {
	return _Abi.Contract.MintAftermarketDeviceByManufacturerBatch(&_Abi.TransactOpts, manufacturerNode, addresses, attributes, infos)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x29f47b90.
//
// Solidity: function mintManufacturer(address owner, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactor) MintManufacturer(opts *bind.TransactOpts, owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "mintManufacturer", owner, attributes, infos)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x29f47b90.
//
// Solidity: function mintManufacturer(address owner, string[] attributes, string[] infos) returns()
func (_Abi *AbiSession) MintManufacturer(owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.MintManufacturer(&_Abi.TransactOpts, owner, attributes, infos)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x29f47b90.
//
// Solidity: function mintManufacturer(address owner, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactorSession) MintManufacturer(owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.MintManufacturer(&_Abi.TransactOpts, owner, attributes, infos)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Abi *AbiTransactor) MintManufacturerBatch(opts *bind.TransactOpts, owner common.Address, names []string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "mintManufacturerBatch", owner, names)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Abi *AbiSession) MintManufacturerBatch(owner common.Address, names []string) (*types.Transaction, error) {
	return _Abi.Contract.MintManufacturerBatch(&_Abi.TransactOpts, owner, names)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Abi *AbiTransactorSession) MintManufacturerBatch(owner common.Address, names []string) (*types.Transaction, error) {
	return _Abi.Contract.MintManufacturerBatch(&_Abi.TransactOpts, owner, names)
}

// MintVehicle is a paid mutator transaction binding the contract method 0xd7d1e236.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactor) MintVehicle(opts *bind.TransactOpts, manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "mintVehicle", manufacturerNode, owner, attributes, infos)
}

// MintVehicle is a paid mutator transaction binding the contract method 0xd7d1e236.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, string[] attributes, string[] infos) returns()
func (_Abi *AbiSession) MintVehicle(manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.MintVehicle(&_Abi.TransactOpts, manufacturerNode, owner, attributes, infos)
}

// MintVehicle is a paid mutator transaction binding the contract method 0xd7d1e236.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactorSession) MintVehicle(manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.MintVehicle(&_Abi.TransactOpts, manufacturerNode, owner, attributes, infos)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x9c4e7155.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, string[] attributes, string[] infos, bytes signature) returns()
func (_Abi *AbiTransactor) MintVehicleSign(opts *bind.TransactOpts, manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string, signature []byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "mintVehicleSign", manufacturerNode, owner, attributes, infos, signature)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x9c4e7155.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, string[] attributes, string[] infos, bytes signature) returns()
func (_Abi *AbiSession) MintVehicleSign(manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string, signature []byte) (*types.Transaction, error) {
	return _Abi.Contract.MintVehicleSign(&_Abi.TransactOpts, manufacturerNode, owner, attributes, infos, signature)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x9c4e7155.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, string[] attributes, string[] infos, bytes signature) returns()
func (_Abi *AbiTransactorSession) MintVehicleSign(manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string, signature []byte) (*types.Transaction, error) {
	return _Abi.Contract.MintVehicleSign(&_Abi.TransactOpts, manufacturerNode, owner, attributes, infos, signature)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Abi *AbiTransactor) PairAftermarketDeviceSign(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "pairAftermarketDeviceSign", aftermarketDeviceNode, vehicleNode, signature)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Abi *AbiSession) PairAftermarketDeviceSign(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Abi.Contract.PairAftermarketDeviceSign(&_Abi.TransactOpts, aftermarketDeviceNode, vehicleNode, signature)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Abi *AbiTransactorSession) PairAftermarketDeviceSign(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Abi.Contract.PairAftermarketDeviceSign(&_Abi.TransactOpts, aftermarketDeviceNode, vehicleNode, signature)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Abi *AbiTransactor) RemoveModule(opts *bind.TransactOpts, implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "removeModule", implementation, selectors)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Abi *AbiSession) RemoveModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Abi.Contract.RemoveModule(&_Abi.TransactOpts, implementation, selectors)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Abi *AbiTransactorSession) RemoveModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Abi.Contract.RemoveModule(&_Abi.TransactOpts, implementation, selectors)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Abi *AbiTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Abi *AbiSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.Contract.RenounceRole(&_Abi.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Abi *AbiTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.Contract.RenounceRole(&_Abi.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Abi *AbiTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Abi *AbiSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.Contract.RevokeRole(&_Abi.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Abi *AbiTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Abi.Contract.RevokeRole(&_Abi.TransactOpts, role, account)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Abi *AbiTransactor) SetAdMintCost(opts *bind.TransactOpts, _adMintCost *big.Int) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setAdMintCost", _adMintCost)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Abi *AbiSession) SetAdMintCost(_adMintCost *big.Int) (*types.Transaction, error) {
	return _Abi.Contract.SetAdMintCost(&_Abi.TransactOpts, _adMintCost)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Abi *AbiTransactorSession) SetAdMintCost(_adMintCost *big.Int) (*types.Transaction, error) {
	return _Abi.Contract.SetAdMintCost(&_Abi.TransactOpts, _adMintCost)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0xf4b64198.
//
// Solidity: function setAftermarketDeviceInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactor) SetAftermarketDeviceInfo(opts *bind.TransactOpts, nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setAftermarketDeviceInfo", nodeId, attributes, infos)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0xf4b64198.
//
// Solidity: function setAftermarketDeviceInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiSession) SetAftermarketDeviceInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.SetAftermarketDeviceInfo(&_Abi.TransactOpts, nodeId, attributes, infos)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0xf4b64198.
//
// Solidity: function setAftermarketDeviceInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactorSession) SetAftermarketDeviceInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.SetAftermarketDeviceInfo(&_Abi.TransactOpts, nodeId, attributes, infos)
}

// SetAftermarketDeviceNodeType is a paid mutator transaction binding the contract method 0x656969da.
//
// Solidity: function setAftermarketDeviceNodeType(bytes label) returns()
func (_Abi *AbiTransactor) SetAftermarketDeviceNodeType(opts *bind.TransactOpts, label []byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setAftermarketDeviceNodeType", label)
}

// SetAftermarketDeviceNodeType is a paid mutator transaction binding the contract method 0x656969da.
//
// Solidity: function setAftermarketDeviceNodeType(bytes label) returns()
func (_Abi *AbiSession) SetAftermarketDeviceNodeType(label []byte) (*types.Transaction, error) {
	return _Abi.Contract.SetAftermarketDeviceNodeType(&_Abi.TransactOpts, label)
}

// SetAftermarketDeviceNodeType is a paid mutator transaction binding the contract method 0x656969da.
//
// Solidity: function setAftermarketDeviceNodeType(bytes label) returns()
func (_Abi *AbiTransactorSession) SetAftermarketDeviceNodeType(label []byte) (*types.Transaction, error) {
	return _Abi.Contract.SetAftermarketDeviceNodeType(&_Abi.TransactOpts, label)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string _baseURI) returns()
func (_Abi *AbiTransactor) SetBaseURI(opts *bind.TransactOpts, _baseURI string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setBaseURI", _baseURI)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string _baseURI) returns()
func (_Abi *AbiSession) SetBaseURI(_baseURI string) (*types.Transaction, error) {
	return _Abi.Contract.SetBaseURI(&_Abi.TransactOpts, _baseURI)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string _baseURI) returns()
func (_Abi *AbiTransactorSession) SetBaseURI(_baseURI string) (*types.Transaction, error) {
	return _Abi.Contract.SetBaseURI(&_Abi.TransactOpts, _baseURI)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Abi *AbiTransactor) SetController(opts *bind.TransactOpts, _controller common.Address) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setController", _controller)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Abi *AbiSession) SetController(_controller common.Address) (*types.Transaction, error) {
	return _Abi.Contract.SetController(&_Abi.TransactOpts, _controller)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Abi *AbiTransactorSession) SetController(_controller common.Address) (*types.Transaction, error) {
	return _Abi.Contract.SetController(&_Abi.TransactOpts, _controller)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Abi *AbiTransactor) SetDimoToken(opts *bind.TransactOpts, _dimoToken common.Address) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setDimoToken", _dimoToken)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Abi *AbiSession) SetDimoToken(_dimoToken common.Address) (*types.Transaction, error) {
	return _Abi.Contract.SetDimoToken(&_Abi.TransactOpts, _dimoToken)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Abi *AbiTransactorSession) SetDimoToken(_dimoToken common.Address) (*types.Transaction, error) {
	return _Abi.Contract.SetDimoToken(&_Abi.TransactOpts, _dimoToken)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Abi *AbiTransactor) SetFoundationAddress(opts *bind.TransactOpts, _foundation common.Address) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setFoundationAddress", _foundation)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Abi *AbiSession) SetFoundationAddress(_foundation common.Address) (*types.Transaction, error) {
	return _Abi.Contract.SetFoundationAddress(&_Abi.TransactOpts, _foundation)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Abi *AbiTransactorSession) SetFoundationAddress(_foundation common.Address) (*types.Transaction, error) {
	return _Abi.Contract.SetFoundationAddress(&_Abi.TransactOpts, _foundation)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Abi *AbiTransactor) SetLicense(opts *bind.TransactOpts, _license common.Address) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setLicense", _license)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Abi *AbiSession) SetLicense(_license common.Address) (*types.Transaction, error) {
	return _Abi.Contract.SetLicense(&_Abi.TransactOpts, _license)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Abi *AbiTransactorSession) SetLicense(_license common.Address) (*types.Transaction, error) {
	return _Abi.Contract.SetLicense(&_Abi.TransactOpts, _license)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0xd89e7dbc.
//
// Solidity: function setManufacturerInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactor) SetManufacturerInfo(opts *bind.TransactOpts, nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setManufacturerInfo", nodeId, attributes, infos)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0xd89e7dbc.
//
// Solidity: function setManufacturerInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiSession) SetManufacturerInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.SetManufacturerInfo(&_Abi.TransactOpts, nodeId, attributes, infos)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0xd89e7dbc.
//
// Solidity: function setManufacturerInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactorSession) SetManufacturerInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.SetManufacturerInfo(&_Abi.TransactOpts, nodeId, attributes, infos)
}

// SetManufacturerNodeType is a paid mutator transaction binding the contract method 0xca9ba80e.
//
// Solidity: function setManufacturerNodeType(bytes label) returns()
func (_Abi *AbiTransactor) SetManufacturerNodeType(opts *bind.TransactOpts, label []byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setManufacturerNodeType", label)
}

// SetManufacturerNodeType is a paid mutator transaction binding the contract method 0xca9ba80e.
//
// Solidity: function setManufacturerNodeType(bytes label) returns()
func (_Abi *AbiSession) SetManufacturerNodeType(label []byte) (*types.Transaction, error) {
	return _Abi.Contract.SetManufacturerNodeType(&_Abi.TransactOpts, label)
}

// SetManufacturerNodeType is a paid mutator transaction binding the contract method 0xca9ba80e.
//
// Solidity: function setManufacturerNodeType(bytes label) returns()
func (_Abi *AbiTransactorSession) SetManufacturerNodeType(label []byte) (*types.Transaction, error) {
	return _Abi.Contract.SetManufacturerNodeType(&_Abi.TransactOpts, label)
}

// SetTokenURI is a paid mutator transaction binding the contract method 0x162094c4.
//
// Solidity: function setTokenURI(uint256 tokenId, string _tokenURI) returns()
func (_Abi *AbiTransactor) SetTokenURI(opts *bind.TransactOpts, tokenId *big.Int, _tokenURI string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setTokenURI", tokenId, _tokenURI)
}

// SetTokenURI is a paid mutator transaction binding the contract method 0x162094c4.
//
// Solidity: function setTokenURI(uint256 tokenId, string _tokenURI) returns()
func (_Abi *AbiSession) SetTokenURI(tokenId *big.Int, _tokenURI string) (*types.Transaction, error) {
	return _Abi.Contract.SetTokenURI(&_Abi.TransactOpts, tokenId, _tokenURI)
}

// SetTokenURI is a paid mutator transaction binding the contract method 0x162094c4.
//
// Solidity: function setTokenURI(uint256 tokenId, string _tokenURI) returns()
func (_Abi *AbiTransactorSession) SetTokenURI(tokenId *big.Int, _tokenURI string) (*types.Transaction, error) {
	return _Abi.Contract.SetTokenURI(&_Abi.TransactOpts, tokenId, _tokenURI)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xc175eb46.
//
// Solidity: function setVehicleInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactor) SetVehicleInfo(opts *bind.TransactOpts, nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setVehicleInfo", nodeId, attributes, infos)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xc175eb46.
//
// Solidity: function setVehicleInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiSession) SetVehicleInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.SetVehicleInfo(&_Abi.TransactOpts, nodeId, attributes, infos)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xc175eb46.
//
// Solidity: function setVehicleInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Abi *AbiTransactorSession) SetVehicleInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Abi.Contract.SetVehicleInfo(&_Abi.TransactOpts, nodeId, attributes, infos)
}

// SetVehicleNodeType is a paid mutator transaction binding the contract method 0x63822b13.
//
// Solidity: function setVehicleNodeType(bytes label) returns()
func (_Abi *AbiTransactor) SetVehicleNodeType(opts *bind.TransactOpts, label []byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "setVehicleNodeType", label)
}

// SetVehicleNodeType is a paid mutator transaction binding the contract method 0x63822b13.
//
// Solidity: function setVehicleNodeType(bytes label) returns()
func (_Abi *AbiSession) SetVehicleNodeType(label []byte) (*types.Transaction, error) {
	return _Abi.Contract.SetVehicleNodeType(&_Abi.TransactOpts, label)
}

// SetVehicleNodeType is a paid mutator transaction binding the contract method 0x63822b13.
//
// Solidity: function setVehicleNodeType(bytes label) returns()
func (_Abi *AbiTransactorSession) SetVehicleNodeType(label []byte) (*types.Transaction, error) {
	return _Abi.Contract.SetVehicleNodeType(&_Abi.TransactOpts, label)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Abi *AbiTransactor) UpdateModule(opts *bind.TransactOpts, oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "updateModule", oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Abi *AbiSession) UpdateModule(oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Abi.Contract.UpdateModule(&_Abi.TransactOpts, oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Abi *AbiTransactorSession) UpdateModule(oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Abi.Contract.UpdateModule(&_Abi.TransactOpts, oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Abi *AbiTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Abi.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Abi *AbiSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Abi.Contract.Fallback(&_Abi.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Abi *AbiTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Abi.Contract.Fallback(&_Abi.TransactOpts, calldata)
}

// AbiAftermarketDeviceClaimedIterator is returned from FilterAftermarketDeviceClaimed and is used to iterate over the raw logs and unpacked data for AftermarketDeviceClaimed events raised by the Abi contract.
type AbiAftermarketDeviceClaimedIterator struct {
	Event *AbiAftermarketDeviceClaimed // Event containing the contract specifics and raw log

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
func (it *AbiAftermarketDeviceClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiAftermarketDeviceClaimed)
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
		it.Event = new(AbiAftermarketDeviceClaimed)
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
func (it *AbiAftermarketDeviceClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiAftermarketDeviceClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiAftermarketDeviceClaimed represents a AftermarketDeviceClaimed event raised by the Abi contract.
type AbiAftermarketDeviceClaimed struct {
	AftermarketDeviceNode *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceClaimed is a free log retrieval operation binding the contract event 0x8468d811e5090d3b1a07e28af524e66c128f624e16b07638f419012c779f76ec.
//
// Solidity: event AftermarketDeviceClaimed(uint256 indexed aftermarketDeviceNode, address indexed owner)
func (_Abi *AbiFilterer) FilterAftermarketDeviceClaimed(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int, owner []common.Address) (*AbiAftermarketDeviceClaimedIterator, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "AftermarketDeviceClaimed", aftermarketDeviceNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &AbiAftermarketDeviceClaimedIterator{contract: _Abi.contract, event: "AftermarketDeviceClaimed", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceClaimed is a free log subscription operation binding the contract event 0x8468d811e5090d3b1a07e28af524e66c128f624e16b07638f419012c779f76ec.
//
// Solidity: event AftermarketDeviceClaimed(uint256 indexed aftermarketDeviceNode, address indexed owner)
func (_Abi *AbiFilterer) WatchAftermarketDeviceClaimed(opts *bind.WatchOpts, sink chan<- *AbiAftermarketDeviceClaimed, aftermarketDeviceNode []*big.Int, owner []common.Address) (event.Subscription, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "AftermarketDeviceClaimed", aftermarketDeviceNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiAftermarketDeviceClaimed)
				if err := _Abi.contract.UnpackLog(event, "AftermarketDeviceClaimed", log); err != nil {
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

// ParseAftermarketDeviceClaimed is a log parse operation binding the contract event 0x8468d811e5090d3b1a07e28af524e66c128f624e16b07638f419012c779f76ec.
//
// Solidity: event AftermarketDeviceClaimed(uint256 indexed aftermarketDeviceNode, address indexed owner)
func (_Abi *AbiFilterer) ParseAftermarketDeviceClaimed(log types.Log) (*AbiAftermarketDeviceClaimed, error) {
	event := new(AbiAftermarketDeviceClaimed)
	if err := _Abi.contract.UnpackLog(event, "AftermarketDeviceClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiAftermarketDeviceNodeMintedIterator is returned from FilterAftermarketDeviceNodeMinted and is used to iterate over the raw logs and unpacked data for AftermarketDeviceNodeMinted events raised by the Abi contract.
type AbiAftermarketDeviceNodeMintedIterator struct {
	Event *AbiAftermarketDeviceNodeMinted // Event containing the contract specifics and raw log

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
func (it *AbiAftermarketDeviceNodeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiAftermarketDeviceNodeMinted)
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
		it.Event = new(AbiAftermarketDeviceNodeMinted)
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
func (it *AbiAftermarketDeviceNodeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiAftermarketDeviceNodeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiAftermarketDeviceNodeMinted represents a AftermarketDeviceNodeMinted event raised by the Abi contract.
type AbiAftermarketDeviceNodeMinted struct {
	NodeType                 *big.Int
	NodeId                   *big.Int
	AftermarketDeviceAddress common.Address
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceNodeMinted is a free log retrieval operation binding the contract event 0x7856cab6942610b9c833a5d8a87a2c88deb168f56f6da3599900da04c13705e9.
//
// Solidity: event AftermarketDeviceNodeMinted(uint256 indexed nodeType, uint256 indexed nodeId, address indexed aftermarketDeviceAddress)
func (_Abi *AbiFilterer) FilterAftermarketDeviceNodeMinted(opts *bind.FilterOpts, nodeType []*big.Int, nodeId []*big.Int, aftermarketDeviceAddress []common.Address) (*AbiAftermarketDeviceNodeMintedIterator, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var nodeIdRule []interface{}
	for _, nodeIdItem := range nodeId {
		nodeIdRule = append(nodeIdRule, nodeIdItem)
	}
	var aftermarketDeviceAddressRule []interface{}
	for _, aftermarketDeviceAddressItem := range aftermarketDeviceAddress {
		aftermarketDeviceAddressRule = append(aftermarketDeviceAddressRule, aftermarketDeviceAddressItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "AftermarketDeviceNodeMinted", nodeTypeRule, nodeIdRule, aftermarketDeviceAddressRule)
	if err != nil {
		return nil, err
	}
	return &AbiAftermarketDeviceNodeMintedIterator{contract: _Abi.contract, event: "AftermarketDeviceNodeMinted", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceNodeMinted is a free log subscription operation binding the contract event 0x7856cab6942610b9c833a5d8a87a2c88deb168f56f6da3599900da04c13705e9.
//
// Solidity: event AftermarketDeviceNodeMinted(uint256 indexed nodeType, uint256 indexed nodeId, address indexed aftermarketDeviceAddress)
func (_Abi *AbiFilterer) WatchAftermarketDeviceNodeMinted(opts *bind.WatchOpts, sink chan<- *AbiAftermarketDeviceNodeMinted, nodeType []*big.Int, nodeId []*big.Int, aftermarketDeviceAddress []common.Address) (event.Subscription, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var nodeIdRule []interface{}
	for _, nodeIdItem := range nodeId {
		nodeIdRule = append(nodeIdRule, nodeIdItem)
	}
	var aftermarketDeviceAddressRule []interface{}
	for _, aftermarketDeviceAddressItem := range aftermarketDeviceAddress {
		aftermarketDeviceAddressRule = append(aftermarketDeviceAddressRule, aftermarketDeviceAddressItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "AftermarketDeviceNodeMinted", nodeTypeRule, nodeIdRule, aftermarketDeviceAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiAftermarketDeviceNodeMinted)
				if err := _Abi.contract.UnpackLog(event, "AftermarketDeviceNodeMinted", log); err != nil {
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

// ParseAftermarketDeviceNodeMinted is a log parse operation binding the contract event 0x7856cab6942610b9c833a5d8a87a2c88deb168f56f6da3599900da04c13705e9.
//
// Solidity: event AftermarketDeviceNodeMinted(uint256 indexed nodeType, uint256 indexed nodeId, address indexed aftermarketDeviceAddress)
func (_Abi *AbiFilterer) ParseAftermarketDeviceNodeMinted(log types.Log) (*AbiAftermarketDeviceNodeMinted, error) {
	event := new(AbiAftermarketDeviceNodeMinted)
	if err := _Abi.contract.UnpackLog(event, "AftermarketDeviceNodeMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiAftermarketDevicePairedIterator is returned from FilterAftermarketDevicePaired and is used to iterate over the raw logs and unpacked data for AftermarketDevicePaired events raised by the Abi contract.
type AbiAftermarketDevicePairedIterator struct {
	Event *AbiAftermarketDevicePaired // Event containing the contract specifics and raw log

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
func (it *AbiAftermarketDevicePairedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiAftermarketDevicePaired)
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
		it.Event = new(AbiAftermarketDevicePaired)
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
func (it *AbiAftermarketDevicePairedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiAftermarketDevicePairedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiAftermarketDevicePaired represents a AftermarketDevicePaired event raised by the Abi contract.
type AbiAftermarketDevicePaired struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDevicePaired is a free log retrieval operation binding the contract event 0x89ec132808bbf01af00b90fd34e04fd6cfb8dba2813ca5446a415500b83c7938.
//
// Solidity: event AftermarketDevicePaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Abi *AbiFilterer) FilterAftermarketDevicePaired(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int, vehicleNode []*big.Int, owner []common.Address) (*AbiAftermarketDevicePairedIterator, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var vehicleNodeRule []interface{}
	for _, vehicleNodeItem := range vehicleNode {
		vehicleNodeRule = append(vehicleNodeRule, vehicleNodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "AftermarketDevicePaired", aftermarketDeviceNodeRule, vehicleNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &AbiAftermarketDevicePairedIterator{contract: _Abi.contract, event: "AftermarketDevicePaired", logs: logs, sub: sub}, nil
}

// WatchAftermarketDevicePaired is a free log subscription operation binding the contract event 0x89ec132808bbf01af00b90fd34e04fd6cfb8dba2813ca5446a415500b83c7938.
//
// Solidity: event AftermarketDevicePaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Abi *AbiFilterer) WatchAftermarketDevicePaired(opts *bind.WatchOpts, sink chan<- *AbiAftermarketDevicePaired, aftermarketDeviceNode []*big.Int, vehicleNode []*big.Int, owner []common.Address) (event.Subscription, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var vehicleNodeRule []interface{}
	for _, vehicleNodeItem := range vehicleNode {
		vehicleNodeRule = append(vehicleNodeRule, vehicleNodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "AftermarketDevicePaired", aftermarketDeviceNodeRule, vehicleNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiAftermarketDevicePaired)
				if err := _Abi.contract.UnpackLog(event, "AftermarketDevicePaired", log); err != nil {
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

// ParseAftermarketDevicePaired is a log parse operation binding the contract event 0x89ec132808bbf01af00b90fd34e04fd6cfb8dba2813ca5446a415500b83c7938.
//
// Solidity: event AftermarketDevicePaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Abi *AbiFilterer) ParseAftermarketDevicePaired(log types.Log) (*AbiAftermarketDevicePaired, error) {
	event := new(AbiAftermarketDevicePaired)
	if err := _Abi.contract.UnpackLog(event, "AftermarketDevicePaired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Abi contract.
type AbiApprovalIterator struct {
	Event *AbiApproval // Event containing the contract specifics and raw log

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
func (it *AbiApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiApproval)
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
		it.Event = new(AbiApproval)
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
func (it *AbiApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiApproval represents a Approval event raised by the Abi contract.
type AbiApproval struct {
	Owner    common.Address
	Operator common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed operator, uint256 indexed tokenId)
func (_Abi *AbiFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, operator []common.Address, tokenId []*big.Int) (*AbiApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "Approval", ownerRule, operatorRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AbiApprovalIterator{contract: _Abi.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed operator, uint256 indexed tokenId)
func (_Abi *AbiFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *AbiApproval, owner []common.Address, operator []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "Approval", ownerRule, operatorRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiApproval)
				if err := _Abi.contract.UnpackLog(event, "Approval", log); err != nil {
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
// Solidity: event Approval(address indexed owner, address indexed operator, uint256 indexed tokenId)
func (_Abi *AbiFilterer) ParseApproval(log types.Log) (*AbiApproval, error) {
	event := new(AbiApproval)
	if err := _Abi.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the Abi contract.
type AbiApprovalForAllIterator struct {
	Event *AbiApprovalForAll // Event containing the contract specifics and raw log

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
func (it *AbiApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiApprovalForAll)
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
		it.Event = new(AbiApprovalForAll)
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
func (it *AbiApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiApprovalForAll represents a ApprovalForAll event raised by the Abi contract.
type AbiApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Abi *AbiFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*AbiApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &AbiApprovalForAllIterator{contract: _Abi.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Abi *AbiFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *AbiApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiApprovalForAll)
				if err := _Abi.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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
func (_Abi *AbiFilterer) ParseApprovalForAll(log types.Log) (*AbiApprovalForAll, error) {
	event := new(AbiApprovalForAll)
	if err := _Abi.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiAttributeAddedIterator is returned from FilterAttributeAdded and is used to iterate over the raw logs and unpacked data for AttributeAdded events raised by the Abi contract.
type AbiAttributeAddedIterator struct {
	Event *AbiAttributeAdded // Event containing the contract specifics and raw log

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
func (it *AbiAttributeAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiAttributeAdded)
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
		it.Event = new(AbiAttributeAdded)
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
func (it *AbiAttributeAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiAttributeAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiAttributeAdded represents a AttributeAdded event raised by the Abi contract.
type AbiAttributeAdded struct {
	NodeType  *big.Int
	Attribute common.Hash
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAttributeAdded is a free log retrieval operation binding the contract event 0xdee1f2fc87d9c834bee1095ebfc0b81ae1b364a7c74060167ab8a82623b22f9c.
//
// Solidity: event AttributeAdded(uint256 indexed nodeType, string indexed attribute)
func (_Abi *AbiFilterer) FilterAttributeAdded(opts *bind.FilterOpts, nodeType []*big.Int, attribute []string) (*AbiAttributeAddedIterator, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var attributeRule []interface{}
	for _, attributeItem := range attribute {
		attributeRule = append(attributeRule, attributeItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "AttributeAdded", nodeTypeRule, attributeRule)
	if err != nil {
		return nil, err
	}
	return &AbiAttributeAddedIterator{contract: _Abi.contract, event: "AttributeAdded", logs: logs, sub: sub}, nil
}

// WatchAttributeAdded is a free log subscription operation binding the contract event 0xdee1f2fc87d9c834bee1095ebfc0b81ae1b364a7c74060167ab8a82623b22f9c.
//
// Solidity: event AttributeAdded(uint256 indexed nodeType, string indexed attribute)
func (_Abi *AbiFilterer) WatchAttributeAdded(opts *bind.WatchOpts, sink chan<- *AbiAttributeAdded, nodeType []*big.Int, attribute []string) (event.Subscription, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var attributeRule []interface{}
	for _, attributeItem := range attribute {
		attributeRule = append(attributeRule, attributeItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "AttributeAdded", nodeTypeRule, attributeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiAttributeAdded)
				if err := _Abi.contract.UnpackLog(event, "AttributeAdded", log); err != nil {
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

// ParseAttributeAdded is a log parse operation binding the contract event 0xdee1f2fc87d9c834bee1095ebfc0b81ae1b364a7c74060167ab8a82623b22f9c.
//
// Solidity: event AttributeAdded(uint256 indexed nodeType, string indexed attribute)
func (_Abi *AbiFilterer) ParseAttributeAdded(log types.Log) (*AbiAttributeAdded, error) {
	event := new(AbiAttributeAdded)
	if err := _Abi.contract.UnpackLog(event, "AttributeAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiControllerSetIterator is returned from FilterControllerSet and is used to iterate over the raw logs and unpacked data for ControllerSet events raised by the Abi contract.
type AbiControllerSetIterator struct {
	Event *AbiControllerSet // Event containing the contract specifics and raw log

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
func (it *AbiControllerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiControllerSet)
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
		it.Event = new(AbiControllerSet)
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
func (it *AbiControllerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiControllerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiControllerSet represents a ControllerSet event raised by the Abi contract.
type AbiControllerSet struct {
	Controller common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterControllerSet is a free log retrieval operation binding the contract event 0x79f74fd5964b6943d8a1865abfb7f668c92fa3f32c0a2e3195da7d0946703ad7.
//
// Solidity: event ControllerSet(address indexed controller)
func (_Abi *AbiFilterer) FilterControllerSet(opts *bind.FilterOpts, controller []common.Address) (*AbiControllerSetIterator, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "ControllerSet", controllerRule)
	if err != nil {
		return nil, err
	}
	return &AbiControllerSetIterator{contract: _Abi.contract, event: "ControllerSet", logs: logs, sub: sub}, nil
}

// WatchControllerSet is a free log subscription operation binding the contract event 0x79f74fd5964b6943d8a1865abfb7f668c92fa3f32c0a2e3195da7d0946703ad7.
//
// Solidity: event ControllerSet(address indexed controller)
func (_Abi *AbiFilterer) WatchControllerSet(opts *bind.WatchOpts, sink chan<- *AbiControllerSet, controller []common.Address) (event.Subscription, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "ControllerSet", controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiControllerSet)
				if err := _Abi.contract.UnpackLog(event, "ControllerSet", log); err != nil {
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

// ParseControllerSet is a log parse operation binding the contract event 0x79f74fd5964b6943d8a1865abfb7f668c92fa3f32c0a2e3195da7d0946703ad7.
//
// Solidity: event ControllerSet(address indexed controller)
func (_Abi *AbiFilterer) ParseControllerSet(log types.Log) (*AbiControllerSet, error) {
	event := new(AbiControllerSet)
	if err := _Abi.contract.UnpackLog(event, "ControllerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiModuleAddedIterator is returned from FilterModuleAdded and is used to iterate over the raw logs and unpacked data for ModuleAdded events raised by the Abi contract.
type AbiModuleAddedIterator struct {
	Event *AbiModuleAdded // Event containing the contract specifics and raw log

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
func (it *AbiModuleAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiModuleAdded)
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
		it.Event = new(AbiModuleAdded)
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
func (it *AbiModuleAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiModuleAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiModuleAdded represents a ModuleAdded event raised by the Abi contract.
type AbiModuleAdded struct {
	ModuleAddr common.Address
	Selectors  [][4]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterModuleAdded is a free log retrieval operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_Abi *AbiFilterer) FilterModuleAdded(opts *bind.FilterOpts, moduleAddr []common.Address) (*AbiModuleAddedIterator, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "ModuleAdded", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return &AbiModuleAddedIterator{contract: _Abi.contract, event: "ModuleAdded", logs: logs, sub: sub}, nil
}

// WatchModuleAdded is a free log subscription operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_Abi *AbiFilterer) WatchModuleAdded(opts *bind.WatchOpts, sink chan<- *AbiModuleAdded, moduleAddr []common.Address) (event.Subscription, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "ModuleAdded", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiModuleAdded)
				if err := _Abi.contract.UnpackLog(event, "ModuleAdded", log); err != nil {
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

// ParseModuleAdded is a log parse operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_Abi *AbiFilterer) ParseModuleAdded(log types.Log) (*AbiModuleAdded, error) {
	event := new(AbiModuleAdded)
	if err := _Abi.contract.UnpackLog(event, "ModuleAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiModuleRemovedIterator is returned from FilterModuleRemoved and is used to iterate over the raw logs and unpacked data for ModuleRemoved events raised by the Abi contract.
type AbiModuleRemovedIterator struct {
	Event *AbiModuleRemoved // Event containing the contract specifics and raw log

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
func (it *AbiModuleRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiModuleRemoved)
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
		it.Event = new(AbiModuleRemoved)
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
func (it *AbiModuleRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiModuleRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiModuleRemoved represents a ModuleRemoved event raised by the Abi contract.
type AbiModuleRemoved struct {
	ModuleAddr common.Address
	Selectors  [][4]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterModuleRemoved is a free log retrieval operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_Abi *AbiFilterer) FilterModuleRemoved(opts *bind.FilterOpts, moduleAddr []common.Address) (*AbiModuleRemovedIterator, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "ModuleRemoved", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return &AbiModuleRemovedIterator{contract: _Abi.contract, event: "ModuleRemoved", logs: logs, sub: sub}, nil
}

// WatchModuleRemoved is a free log subscription operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_Abi *AbiFilterer) WatchModuleRemoved(opts *bind.WatchOpts, sink chan<- *AbiModuleRemoved, moduleAddr []common.Address) (event.Subscription, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "ModuleRemoved", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiModuleRemoved)
				if err := _Abi.contract.UnpackLog(event, "ModuleRemoved", log); err != nil {
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

// ParseModuleRemoved is a log parse operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_Abi *AbiFilterer) ParseModuleRemoved(log types.Log) (*AbiModuleRemoved, error) {
	event := new(AbiModuleRemoved)
	if err := _Abi.contract.UnpackLog(event, "ModuleRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiModuleUpdatedIterator is returned from FilterModuleUpdated and is used to iterate over the raw logs and unpacked data for ModuleUpdated events raised by the Abi contract.
type AbiModuleUpdatedIterator struct {
	Event *AbiModuleUpdated // Event containing the contract specifics and raw log

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
func (it *AbiModuleUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiModuleUpdated)
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
		it.Event = new(AbiModuleUpdated)
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
func (it *AbiModuleUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiModuleUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiModuleUpdated represents a ModuleUpdated event raised by the Abi contract.
type AbiModuleUpdated struct {
	OldImplementation common.Address
	NewImplementation common.Address
	OldSelectors      [][4]byte
	NewSelectors      [][4]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterModuleUpdated is a free log retrieval operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_Abi *AbiFilterer) FilterModuleUpdated(opts *bind.FilterOpts, oldImplementation []common.Address, newImplementation []common.Address) (*AbiModuleUpdatedIterator, error) {

	var oldImplementationRule []interface{}
	for _, oldImplementationItem := range oldImplementation {
		oldImplementationRule = append(oldImplementationRule, oldImplementationItem)
	}
	var newImplementationRule []interface{}
	for _, newImplementationItem := range newImplementation {
		newImplementationRule = append(newImplementationRule, newImplementationItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "ModuleUpdated", oldImplementationRule, newImplementationRule)
	if err != nil {
		return nil, err
	}
	return &AbiModuleUpdatedIterator{contract: _Abi.contract, event: "ModuleUpdated", logs: logs, sub: sub}, nil
}

// WatchModuleUpdated is a free log subscription operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_Abi *AbiFilterer) WatchModuleUpdated(opts *bind.WatchOpts, sink chan<- *AbiModuleUpdated, oldImplementation []common.Address, newImplementation []common.Address) (event.Subscription, error) {

	var oldImplementationRule []interface{}
	for _, oldImplementationItem := range oldImplementation {
		oldImplementationRule = append(oldImplementationRule, oldImplementationItem)
	}
	var newImplementationRule []interface{}
	for _, newImplementationItem := range newImplementation {
		newImplementationRule = append(newImplementationRule, newImplementationItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "ModuleUpdated", oldImplementationRule, newImplementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiModuleUpdated)
				if err := _Abi.contract.UnpackLog(event, "ModuleUpdated", log); err != nil {
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

// ParseModuleUpdated is a log parse operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_Abi *AbiFilterer) ParseModuleUpdated(log types.Log) (*AbiModuleUpdated, error) {
	event := new(AbiModuleUpdated)
	if err := _Abi.contract.UnpackLog(event, "ModuleUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiNodeMintedIterator is returned from FilterNodeMinted and is used to iterate over the raw logs and unpacked data for NodeMinted events raised by the Abi contract.
type AbiNodeMintedIterator struct {
	Event *AbiNodeMinted // Event containing the contract specifics and raw log

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
func (it *AbiNodeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiNodeMinted)
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
		it.Event = new(AbiNodeMinted)
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
func (it *AbiNodeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiNodeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiNodeMinted represents a NodeMinted event raised by the Abi contract.
type AbiNodeMinted struct {
	NodeType *big.Int
	NodeId   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNodeMinted is a free log retrieval operation binding the contract event 0x0c2616265c4fd089569533525abc7b19b9f82b423d7cdb61801490b8f9e0ce59.
//
// Solidity: event NodeMinted(uint256 indexed nodeType, uint256 indexed nodeId)
func (_Abi *AbiFilterer) FilterNodeMinted(opts *bind.FilterOpts, nodeType []*big.Int, nodeId []*big.Int) (*AbiNodeMintedIterator, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var nodeIdRule []interface{}
	for _, nodeIdItem := range nodeId {
		nodeIdRule = append(nodeIdRule, nodeIdItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "NodeMinted", nodeTypeRule, nodeIdRule)
	if err != nil {
		return nil, err
	}
	return &AbiNodeMintedIterator{contract: _Abi.contract, event: "NodeMinted", logs: logs, sub: sub}, nil
}

// WatchNodeMinted is a free log subscription operation binding the contract event 0x0c2616265c4fd089569533525abc7b19b9f82b423d7cdb61801490b8f9e0ce59.
//
// Solidity: event NodeMinted(uint256 indexed nodeType, uint256 indexed nodeId)
func (_Abi *AbiFilterer) WatchNodeMinted(opts *bind.WatchOpts, sink chan<- *AbiNodeMinted, nodeType []*big.Int, nodeId []*big.Int) (event.Subscription, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var nodeIdRule []interface{}
	for _, nodeIdItem := range nodeId {
		nodeIdRule = append(nodeIdRule, nodeIdItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "NodeMinted", nodeTypeRule, nodeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiNodeMinted)
				if err := _Abi.contract.UnpackLog(event, "NodeMinted", log); err != nil {
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

// ParseNodeMinted is a log parse operation binding the contract event 0x0c2616265c4fd089569533525abc7b19b9f82b423d7cdb61801490b8f9e0ce59.
//
// Solidity: event NodeMinted(uint256 indexed nodeType, uint256 indexed nodeId)
func (_Abi *AbiFilterer) ParseNodeMinted(log types.Log) (*AbiNodeMinted, error) {
	event := new(AbiNodeMinted)
	if err := _Abi.contract.UnpackLog(event, "NodeMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Abi contract.
type AbiRoleAdminChangedIterator struct {
	Event *AbiRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *AbiRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiRoleAdminChanged)
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
		it.Event = new(AbiRoleAdminChanged)
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
func (it *AbiRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiRoleAdminChanged represents a RoleAdminChanged event raised by the Abi contract.
type AbiRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Abi *AbiFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*AbiRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &AbiRoleAdminChangedIterator{contract: _Abi.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Abi *AbiFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *AbiRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiRoleAdminChanged)
				if err := _Abi.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Abi *AbiFilterer) ParseRoleAdminChanged(log types.Log) (*AbiRoleAdminChanged, error) {
	event := new(AbiRoleAdminChanged)
	if err := _Abi.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Abi contract.
type AbiRoleGrantedIterator struct {
	Event *AbiRoleGranted // Event containing the contract specifics and raw log

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
func (it *AbiRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiRoleGranted)
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
		it.Event = new(AbiRoleGranted)
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
func (it *AbiRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiRoleGranted represents a RoleGranted event raised by the Abi contract.
type AbiRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Abi *AbiFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AbiRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AbiRoleGrantedIterator{contract: _Abi.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Abi *AbiFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *AbiRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiRoleGranted)
				if err := _Abi.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Abi *AbiFilterer) ParseRoleGranted(log types.Log) (*AbiRoleGranted, error) {
	event := new(AbiRoleGranted)
	if err := _Abi.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Abi contract.
type AbiRoleRevokedIterator struct {
	Event *AbiRoleRevoked // Event containing the contract specifics and raw log

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
func (it *AbiRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiRoleRevoked)
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
		it.Event = new(AbiRoleRevoked)
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
func (it *AbiRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiRoleRevoked represents a RoleRevoked event raised by the Abi contract.
type AbiRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Abi *AbiFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AbiRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Abi.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AbiRoleRevokedIterator{contract: _Abi.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Abi *AbiFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *AbiRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Abi.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiRoleRevoked)
				if err := _Abi.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Abi *AbiFilterer) ParseRoleRevoked(log types.Log) (*AbiRoleRevoked, error) {
	event := new(AbiRoleRevoked)
	if err := _Abi.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Abi contract.
type AbiTransferIterator struct {
	Event *AbiTransfer // Event containing the contract specifics and raw log

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
func (it *AbiTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiTransfer)
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
		it.Event = new(AbiTransfer)
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
func (it *AbiTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiTransfer represents a Transfer event raised by the Abi contract.
type AbiTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Abi *AbiFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*AbiTransferIterator, error) {

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

	logs, sub, err := _Abi.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &AbiTransferIterator{contract: _Abi.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Abi *AbiFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *AbiTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Abi.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiTransfer)
				if err := _Abi.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_Abi *AbiFilterer) ParseTransfer(log types.Log) (*AbiTransfer, error) {
	event := new(AbiTransfer)
	if err := _Abi.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
