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

// RegistryMetaData contains all meta data concerning the Registry contract.
var RegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"__baseURI\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"moduleAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"moduleAddr\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldImplementation\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"oldSelectors\",\"type\":\"bytes4[]\"},{\"indexed\":false,\"internalType\":\"bytes4[]\",\"name\":\"newSelectors\",\"type\":\"bytes4[]\"}],\"name\":\"ModuleUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"addModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"selectors\",\"type\":\"bytes4[]\"}],\"name\":\"removeModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oldImplementation\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes4[]\",\"name\":\"oldSelectors\",\"type\":\"bytes4[]\"},{\"internalType\":\"bytes4[]\",\"name\":\"newSelectors\",\"type\":\"bytes4[]\"}],\"name\":\"updateModule\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_adMintCost\",\"type\":\"uint256\"}],\"name\":\"setAdMintCost\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_dimoToken\",\"type\":\"address\"}],\"name\":\"setDimoToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_foundation\",\"type\":\"address\"}],\"name\":\"setFoundationAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_license\",\"type\":\"address\"}],\"name\":\"setLicense\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeType\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"aftermarketDeviceAddress\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceNodeMinted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDevicePaired\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeType\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"AttributeAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeType\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"}],\"name\":\"NodeMinted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addAftermarketDeviceAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"ownerSig\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"aftermarketDeviceSig\",\"type\":\"bytes\"}],\"name\":\"claimAftermarketDeviceSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[][]\",\"name\":\"infos\",\"type\":\"string[][]\"}],\"name\":\"mintAftermarketDeviceByManufacturerBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"pairAftermarketDeviceSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"setAftermarketDeviceInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"label\",\"type\":\"bytes\"}],\"name\":\"setAftermarketDeviceNodeType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"controller\",\"type\":\"address\"}],\"name\":\"ControllerSet\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addManufacturerAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isController\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isController\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isManufacturerMinted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"_isManufacturerMinted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"mintManufacturer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"names\",\"type\":\"string[]\"}],\"name\":\"mintManufacturerBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_controller\",\"type\":\"address\"}],\"name\":\"setController\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"setManufacturerInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"label\",\"type\":\"bytes\"}],\"name\":\"setManufacturerNodeType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"addVehicleAttribute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"mintVehicle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"manufacturerNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"mintVehicleSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"internalType\":\"string[]\",\"name\":\"attributes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"infos\",\"type\":\"string[]\"}],\"name\":\"setVehicleInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"label\",\"type\":\"bytes\"}],\"name\":\"setVehicleNodeType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"attribute\",\"type\":\"string\"}],\"name\":\"getInfo\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"info\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getNodeType\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"nodeType\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getParentNode\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"parentNode\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sourceNode\",\"type\":\"uint256\"}],\"name\":\"getLink\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"targetNode\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"baseURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_baseURI\",\"type\":\"string\"}],\"name\":\"setBaseURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_tokenURI\",\"type\":\"string\"}],\"name\":\"setTokenURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"vehicleNode\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"AftermarketDeviceUnpaired\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aftermarketDeviceNode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferAftermarketDeviceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"aftermarketDeviceNodes\",\"type\":\"uint256[]\"}],\"name\":\"unpairAftermarketDeviceByDeviceNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"vehicleNodes\",\"type\":\"uint256[]\"}],\"name\":\"unpairAftermarketDeviceByVehicleNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// RegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use RegistryMetaData.ABI instead.
var RegistryABI = RegistryMetaData.ABI

// Registry is an auto generated Go binding around an Ethereum contract.
type Registry struct {
	RegistryCaller     // Read-only binding to the contract
	RegistryTransactor // Write-only binding to the contract
	RegistryFilterer   // Log filterer for contract events
}

// RegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type RegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RegistrySession struct {
	Contract     *Registry         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RegistryCallerSession struct {
	Contract *RegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// RegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RegistryTransactorSession struct {
	Contract     *RegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// RegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type RegistryRaw struct {
	Contract *Registry // Generic contract binding to access the raw methods on
}

// RegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RegistryCallerRaw struct {
	Contract *RegistryCaller // Generic read-only contract binding to access the raw methods on
}

// RegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RegistryTransactorRaw struct {
	Contract *RegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRegistry creates a new instance of Registry, bound to a specific deployed contract.
func NewRegistry(address common.Address, backend bind.ContractBackend) (*Registry, error) {
	contract, err := bindRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Registry{RegistryCaller: RegistryCaller{contract: contract}, RegistryTransactor: RegistryTransactor{contract: contract}, RegistryFilterer: RegistryFilterer{contract: contract}}, nil
}

// NewRegistryCaller creates a new read-only instance of Registry, bound to a specific deployed contract.
func NewRegistryCaller(address common.Address, caller bind.ContractCaller) (*RegistryCaller, error) {
	contract, err := bindRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RegistryCaller{contract: contract}, nil
}

// NewRegistryTransactor creates a new write-only instance of Registry, bound to a specific deployed contract.
func NewRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*RegistryTransactor, error) {
	contract, err := bindRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RegistryTransactor{contract: contract}, nil
}

// NewRegistryFilterer creates a new log filterer instance of Registry, bound to a specific deployed contract.
func NewRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*RegistryFilterer, error) {
	contract, err := bindRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RegistryFilterer{contract: contract}, nil
}

// bindRegistry binds a generic wrapper to an already deployed contract.
func bindRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Registry *RegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Registry.Contract.RegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Registry *RegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Registry.Contract.RegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Registry *RegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Registry.Contract.RegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Registry *RegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Registry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Registry *RegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Registry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Registry *RegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Registry.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Registry *RegistryCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Registry *RegistrySession) BalanceOf(account common.Address) (*big.Int, error) {
	return _Registry.Contract.BalanceOf(&_Registry.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Registry *RegistryCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _Registry.Contract.BalanceOf(&_Registry.CallOpts, account)
}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_Registry *RegistryCaller) BaseURI(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "baseURI")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_Registry *RegistrySession) BaseURI() (string, error) {
	return _Registry.Contract.BaseURI(&_Registry.CallOpts)
}

// BaseURI is a free data retrieval call binding the contract method 0x6c0360eb.
//
// Solidity: function baseURI() view returns(string)
func (_Registry *RegistryCallerSession) BaseURI() (string, error) {
	return _Registry.Contract.BaseURI(&_Registry.CallOpts)
}

// GetInfo is a free data retrieval call binding the contract method 0x5cc148f3.
//
// Solidity: function getInfo(uint256 nodeId, string attribute) view returns(string info)
func (_Registry *RegistryCaller) GetInfo(opts *bind.CallOpts, nodeId *big.Int, attribute string) (string, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getInfo", nodeId, attribute)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetInfo is a free data retrieval call binding the contract method 0x5cc148f3.
//
// Solidity: function getInfo(uint256 nodeId, string attribute) view returns(string info)
func (_Registry *RegistrySession) GetInfo(nodeId *big.Int, attribute string) (string, error) {
	return _Registry.Contract.GetInfo(&_Registry.CallOpts, nodeId, attribute)
}

// GetInfo is a free data retrieval call binding the contract method 0x5cc148f3.
//
// Solidity: function getInfo(uint256 nodeId, string attribute) view returns(string info)
func (_Registry *RegistryCallerSession) GetInfo(nodeId *big.Int, attribute string) (string, error) {
	return _Registry.Contract.GetInfo(&_Registry.CallOpts, nodeId, attribute)
}

// GetLink is a free data retrieval call binding the contract method 0x393b6df3.
//
// Solidity: function getLink(uint256 sourceNode) view returns(uint256 targetNode)
func (_Registry *RegistryCaller) GetLink(opts *bind.CallOpts, sourceNode *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getLink", sourceNode)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLink is a free data retrieval call binding the contract method 0x393b6df3.
//
// Solidity: function getLink(uint256 sourceNode) view returns(uint256 targetNode)
func (_Registry *RegistrySession) GetLink(sourceNode *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetLink(&_Registry.CallOpts, sourceNode)
}

// GetLink is a free data retrieval call binding the contract method 0x393b6df3.
//
// Solidity: function getLink(uint256 sourceNode) view returns(uint256 targetNode)
func (_Registry *RegistryCallerSession) GetLink(sourceNode *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetLink(&_Registry.CallOpts, sourceNode)
}

// GetNodeType is a free data retrieval call binding the contract method 0x70c3e13b.
//
// Solidity: function getNodeType(uint256 tokenId) view returns(uint256 nodeType)
func (_Registry *RegistryCaller) GetNodeType(opts *bind.CallOpts, tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getNodeType", tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNodeType is a free data retrieval call binding the contract method 0x70c3e13b.
//
// Solidity: function getNodeType(uint256 tokenId) view returns(uint256 nodeType)
func (_Registry *RegistrySession) GetNodeType(tokenId *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetNodeType(&_Registry.CallOpts, tokenId)
}

// GetNodeType is a free data retrieval call binding the contract method 0x70c3e13b.
//
// Solidity: function getNodeType(uint256 tokenId) view returns(uint256 nodeType)
func (_Registry *RegistryCallerSession) GetNodeType(tokenId *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetNodeType(&_Registry.CallOpts, tokenId)
}

// GetParentNode is a free data retrieval call binding the contract method 0xc5e80c85.
//
// Solidity: function getParentNode(uint256 tokenId) view returns(uint256 parentNode)
func (_Registry *RegistryCaller) GetParentNode(opts *bind.CallOpts, tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getParentNode", tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetParentNode is a free data retrieval call binding the contract method 0xc5e80c85.
//
// Solidity: function getParentNode(uint256 tokenId) view returns(uint256 parentNode)
func (_Registry *RegistrySession) GetParentNode(tokenId *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetParentNode(&_Registry.CallOpts, tokenId)
}

// GetParentNode is a free data retrieval call binding the contract method 0xc5e80c85.
//
// Solidity: function getParentNode(uint256 tokenId) view returns(uint256 parentNode)
func (_Registry *RegistryCallerSession) GetParentNode(tokenId *big.Int) (*big.Int, error) {
	return _Registry.Contract.GetParentNode(&_Registry.CallOpts, tokenId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Registry *RegistryCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Registry *RegistrySession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Registry.Contract.GetRoleAdmin(&_Registry.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Registry *RegistryCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Registry.Contract.GetRoleAdmin(&_Registry.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Registry *RegistryCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Registry *RegistrySession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Registry.Contract.HasRole(&_Registry.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Registry *RegistryCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Registry.Contract.HasRole(&_Registry.CallOpts, role, account)
}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Registry *RegistryCaller) IsController(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "isController", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Registry *RegistrySession) IsController(addr common.Address) (bool, error) {
	return _Registry.Contract.IsController(&_Registry.CallOpts, addr)
}

// IsController is a free data retrieval call binding the contract method 0xb429afeb.
//
// Solidity: function isController(address addr) view returns(bool _isController)
func (_Registry *RegistryCallerSession) IsController(addr common.Address) (bool, error) {
	return _Registry.Contract.IsController(&_Registry.CallOpts, addr)
}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Registry *RegistryCaller) IsManufacturerMinted(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "isManufacturerMinted", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Registry *RegistrySession) IsManufacturerMinted(addr common.Address) (bool, error) {
	return _Registry.Contract.IsManufacturerMinted(&_Registry.CallOpts, addr)
}

// IsManufacturerMinted is a free data retrieval call binding the contract method 0x456bf169.
//
// Solidity: function isManufacturerMinted(address addr) view returns(bool _isManufacturerMinted)
func (_Registry *RegistryCallerSession) IsManufacturerMinted(addr common.Address) (bool, error) {
	return _Registry.Contract.IsManufacturerMinted(&_Registry.CallOpts, addr)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Registry *RegistryCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Registry *RegistrySession) Name() (string, error) {
	return _Registry.Contract.Name(&_Registry.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Registry *RegistryCallerSession) Name() (string, error) {
	return _Registry.Contract.Name(&_Registry.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Registry *RegistryCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Registry *RegistrySession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Registry.Contract.OwnerOf(&_Registry.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_Registry *RegistryCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _Registry.Contract.OwnerOf(&_Registry.CallOpts, tokenId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Registry *RegistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Registry *RegistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Registry.Contract.SupportsInterface(&_Registry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Registry *RegistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Registry.Contract.SupportsInterface(&_Registry.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Registry *RegistryCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Registry *RegistrySession) Symbol() (string, error) {
	return _Registry.Contract.Symbol(&_Registry.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Registry *RegistryCallerSession) Symbol() (string, error) {
	return _Registry.Contract.Symbol(&_Registry.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Registry *RegistryCaller) TokenURI(opts *bind.CallOpts, tokenId *big.Int) (string, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "tokenURI", tokenId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Registry *RegistrySession) TokenURI(tokenId *big.Int) (string, error) {
	return _Registry.Contract.TokenURI(&_Registry.CallOpts, tokenId)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenId) view returns(string)
func (_Registry *RegistryCallerSession) TokenURI(tokenId *big.Int) (string, error) {
	return _Registry.Contract.TokenURI(&_Registry.CallOpts, tokenId)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Registry *RegistryCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Registry.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Registry *RegistrySession) TotalSupply() (*big.Int, error) {
	return _Registry.Contract.TotalSupply(&_Registry.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Registry *RegistryCallerSession) TotalSupply() (*big.Int, error) {
	return _Registry.Contract.TotalSupply(&_Registry.CallOpts)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Registry *RegistryTransactor) AddAftermarketDeviceAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "addAftermarketDeviceAttribute", attribute)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Registry *RegistrySession) AddAftermarketDeviceAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddAftermarketDeviceAttribute(&_Registry.TransactOpts, attribute)
}

// AddAftermarketDeviceAttribute is a paid mutator transaction binding the contract method 0x6111afa3.
//
// Solidity: function addAftermarketDeviceAttribute(string attribute) returns()
func (_Registry *RegistryTransactorSession) AddAftermarketDeviceAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddAftermarketDeviceAttribute(&_Registry.TransactOpts, attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Registry *RegistryTransactor) AddManufacturerAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "addManufacturerAttribute", attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Registry *RegistrySession) AddManufacturerAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddManufacturerAttribute(&_Registry.TransactOpts, attribute)
}

// AddManufacturerAttribute is a paid mutator transaction binding the contract method 0x50300a3f.
//
// Solidity: function addManufacturerAttribute(string attribute) returns()
func (_Registry *RegistryTransactorSession) AddManufacturerAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddManufacturerAttribute(&_Registry.TransactOpts, attribute)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistryTransactor) AddModule(opts *bind.TransactOpts, implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "addModule", implementation, selectors)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistrySession) AddModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.AddModule(&_Registry.TransactOpts, implementation, selectors)
}

// AddModule is a paid mutator transaction binding the contract method 0x0df5b997.
//
// Solidity: function addModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistryTransactorSession) AddModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.AddModule(&_Registry.TransactOpts, implementation, selectors)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Registry *RegistryTransactor) AddVehicleAttribute(opts *bind.TransactOpts, attribute string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "addVehicleAttribute", attribute)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Registry *RegistrySession) AddVehicleAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddVehicleAttribute(&_Registry.TransactOpts, attribute)
}

// AddVehicleAttribute is a paid mutator transaction binding the contract method 0xf0d1a557.
//
// Solidity: function addVehicleAttribute(string attribute) returns()
func (_Registry *RegistryTransactorSession) AddVehicleAttribute(attribute string) (*types.Transaction, error) {
	return _Registry.Contract.AddVehicleAttribute(&_Registry.TransactOpts, attribute)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Registry *RegistryTransactor) ClaimAftermarketDeviceSign(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "claimAftermarketDeviceSign", aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Registry *RegistrySession) ClaimAftermarketDeviceSign(aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Registry.Contract.ClaimAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// ClaimAftermarketDeviceSign is a paid mutator transaction binding the contract method 0x89a841bb.
//
// Solidity: function claimAftermarketDeviceSign(uint256 aftermarketDeviceNode, address owner, bytes ownerSig, bytes aftermarketDeviceSig) returns()
func (_Registry *RegistryTransactorSession) ClaimAftermarketDeviceSign(aftermarketDeviceNode *big.Int, owner common.Address, ownerSig []byte, aftermarketDeviceSig []byte) (*types.Transaction, error) {
	return _Registry.Contract.ClaimAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, owner, ownerSig, aftermarketDeviceSig)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Registry *RegistrySession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.GrantRole(&_Registry.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.GrantRole(&_Registry.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Registry *RegistryTransactor) Initialize(opts *bind.TransactOpts, name string, version string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "initialize", name, version)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Registry *RegistrySession) Initialize(name string, version string) (*types.Transaction, error) {
	return _Registry.Contract.Initialize(&_Registry.TransactOpts, name, version)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name, string version) returns()
func (_Registry *RegistryTransactorSession) Initialize(name string, version string) (*types.Transaction, error) {
	return _Registry.Contract.Initialize(&_Registry.TransactOpts, name, version)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x6c155f2e.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, address[] addresses, string[] attributes, string[][] infos) returns()
func (_Registry *RegistryTransactor) MintAftermarketDeviceByManufacturerBatch(opts *bind.TransactOpts, manufacturerNode *big.Int, addresses []common.Address, attributes []string, infos [][]string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintAftermarketDeviceByManufacturerBatch", manufacturerNode, addresses, attributes, infos)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x6c155f2e.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, address[] addresses, string[] attributes, string[][] infos) returns()
func (_Registry *RegistrySession) MintAftermarketDeviceByManufacturerBatch(manufacturerNode *big.Int, addresses []common.Address, attributes []string, infos [][]string) (*types.Transaction, error) {
	return _Registry.Contract.MintAftermarketDeviceByManufacturerBatch(&_Registry.TransactOpts, manufacturerNode, addresses, attributes, infos)
}

// MintAftermarketDeviceByManufacturerBatch is a paid mutator transaction binding the contract method 0x6c155f2e.
//
// Solidity: function mintAftermarketDeviceByManufacturerBatch(uint256 manufacturerNode, address[] addresses, string[] attributes, string[][] infos) returns()
func (_Registry *RegistryTransactorSession) MintAftermarketDeviceByManufacturerBatch(manufacturerNode *big.Int, addresses []common.Address, attributes []string, infos [][]string) (*types.Transaction, error) {
	return _Registry.Contract.MintAftermarketDeviceByManufacturerBatch(&_Registry.TransactOpts, manufacturerNode, addresses, attributes, infos)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x29f47b90.
//
// Solidity: function mintManufacturer(address owner, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactor) MintManufacturer(opts *bind.TransactOpts, owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintManufacturer", owner, attributes, infos)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x29f47b90.
//
// Solidity: function mintManufacturer(address owner, string[] attributes, string[] infos) returns()
func (_Registry *RegistrySession) MintManufacturer(owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.MintManufacturer(&_Registry.TransactOpts, owner, attributes, infos)
}

// MintManufacturer is a paid mutator transaction binding the contract method 0x29f47b90.
//
// Solidity: function mintManufacturer(address owner, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactorSession) MintManufacturer(owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.MintManufacturer(&_Registry.TransactOpts, owner, attributes, infos)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Registry *RegistryTransactor) MintManufacturerBatch(opts *bind.TransactOpts, owner common.Address, names []string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintManufacturerBatch", owner, names)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Registry *RegistrySession) MintManufacturerBatch(owner common.Address, names []string) (*types.Transaction, error) {
	return _Registry.Contract.MintManufacturerBatch(&_Registry.TransactOpts, owner, names)
}

// MintManufacturerBatch is a paid mutator transaction binding the contract method 0x9abb3000.
//
// Solidity: function mintManufacturerBatch(address owner, string[] names) returns()
func (_Registry *RegistryTransactorSession) MintManufacturerBatch(owner common.Address, names []string) (*types.Transaction, error) {
	return _Registry.Contract.MintManufacturerBatch(&_Registry.TransactOpts, owner, names)
}

// MintVehicle is a paid mutator transaction binding the contract method 0xd7d1e236.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactor) MintVehicle(opts *bind.TransactOpts, manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintVehicle", manufacturerNode, owner, attributes, infos)
}

// MintVehicle is a paid mutator transaction binding the contract method 0xd7d1e236.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, string[] attributes, string[] infos) returns()
func (_Registry *RegistrySession) MintVehicle(manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.MintVehicle(&_Registry.TransactOpts, manufacturerNode, owner, attributes, infos)
}

// MintVehicle is a paid mutator transaction binding the contract method 0xd7d1e236.
//
// Solidity: function mintVehicle(uint256 manufacturerNode, address owner, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactorSession) MintVehicle(manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.MintVehicle(&_Registry.TransactOpts, manufacturerNode, owner, attributes, infos)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x9c4e7155.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, string[] attributes, string[] infos, bytes signature) returns()
func (_Registry *RegistryTransactor) MintVehicleSign(opts *bind.TransactOpts, manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string, signature []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "mintVehicleSign", manufacturerNode, owner, attributes, infos, signature)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x9c4e7155.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, string[] attributes, string[] infos, bytes signature) returns()
func (_Registry *RegistrySession) MintVehicleSign(manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.MintVehicleSign(&_Registry.TransactOpts, manufacturerNode, owner, attributes, infos, signature)
}

// MintVehicleSign is a paid mutator transaction binding the contract method 0x9c4e7155.
//
// Solidity: function mintVehicleSign(uint256 manufacturerNode, address owner, string[] attributes, string[] infos, bytes signature) returns()
func (_Registry *RegistryTransactorSession) MintVehicleSign(manufacturerNode *big.Int, owner common.Address, attributes []string, infos []string, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.MintVehicleSign(&_Registry.TransactOpts, manufacturerNode, owner, attributes, infos, signature)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistryTransactor) PairAftermarketDeviceSign(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "pairAftermarketDeviceSign", aftermarketDeviceNode, vehicleNode, signature)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistrySession) PairAftermarketDeviceSign(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.PairAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, vehicleNode, signature)
}

// PairAftermarketDeviceSign is a paid mutator transaction binding the contract method 0xcfe642dd.
//
// Solidity: function pairAftermarketDeviceSign(uint256 aftermarketDeviceNode, uint256 vehicleNode, bytes signature) returns()
func (_Registry *RegistryTransactorSession) PairAftermarketDeviceSign(aftermarketDeviceNode *big.Int, vehicleNode *big.Int, signature []byte) (*types.Transaction, error) {
	return _Registry.Contract.PairAftermarketDeviceSign(&_Registry.TransactOpts, aftermarketDeviceNode, vehicleNode, signature)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistryTransactor) RemoveModule(opts *bind.TransactOpts, implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "removeModule", implementation, selectors)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistrySession) RemoveModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.RemoveModule(&_Registry.TransactOpts, implementation, selectors)
}

// RemoveModule is a paid mutator transaction binding the contract method 0x9748a762.
//
// Solidity: function removeModule(address implementation, bytes4[] selectors) returns()
func (_Registry *RegistryTransactorSession) RemoveModule(implementation common.Address, selectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.RemoveModule(&_Registry.TransactOpts, implementation, selectors)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Registry *RegistrySession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.RenounceRole(&_Registry.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.RenounceRole(&_Registry.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Registry *RegistrySession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.RevokeRole(&_Registry.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Registry *RegistryTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Registry.Contract.RevokeRole(&_Registry.TransactOpts, role, account)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Registry *RegistryTransactor) SetAdMintCost(opts *bind.TransactOpts, _adMintCost *big.Int) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setAdMintCost", _adMintCost)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Registry *RegistrySession) SetAdMintCost(_adMintCost *big.Int) (*types.Transaction, error) {
	return _Registry.Contract.SetAdMintCost(&_Registry.TransactOpts, _adMintCost)
}

// SetAdMintCost is a paid mutator transaction binding the contract method 0x2390baa8.
//
// Solidity: function setAdMintCost(uint256 _adMintCost) returns()
func (_Registry *RegistryTransactorSession) SetAdMintCost(_adMintCost *big.Int) (*types.Transaction, error) {
	return _Registry.Contract.SetAdMintCost(&_Registry.TransactOpts, _adMintCost)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0xf4b64198.
//
// Solidity: function setAftermarketDeviceInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactor) SetAftermarketDeviceInfo(opts *bind.TransactOpts, nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setAftermarketDeviceInfo", nodeId, attributes, infos)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0xf4b64198.
//
// Solidity: function setAftermarketDeviceInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistrySession) SetAftermarketDeviceInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.SetAftermarketDeviceInfo(&_Registry.TransactOpts, nodeId, attributes, infos)
}

// SetAftermarketDeviceInfo is a paid mutator transaction binding the contract method 0xf4b64198.
//
// Solidity: function setAftermarketDeviceInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactorSession) SetAftermarketDeviceInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.SetAftermarketDeviceInfo(&_Registry.TransactOpts, nodeId, attributes, infos)
}

// SetAftermarketDeviceNodeType is a paid mutator transaction binding the contract method 0x656969da.
//
// Solidity: function setAftermarketDeviceNodeType(bytes label) returns()
func (_Registry *RegistryTransactor) SetAftermarketDeviceNodeType(opts *bind.TransactOpts, label []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setAftermarketDeviceNodeType", label)
}

// SetAftermarketDeviceNodeType is a paid mutator transaction binding the contract method 0x656969da.
//
// Solidity: function setAftermarketDeviceNodeType(bytes label) returns()
func (_Registry *RegistrySession) SetAftermarketDeviceNodeType(label []byte) (*types.Transaction, error) {
	return _Registry.Contract.SetAftermarketDeviceNodeType(&_Registry.TransactOpts, label)
}

// SetAftermarketDeviceNodeType is a paid mutator transaction binding the contract method 0x656969da.
//
// Solidity: function setAftermarketDeviceNodeType(bytes label) returns()
func (_Registry *RegistryTransactorSession) SetAftermarketDeviceNodeType(label []byte) (*types.Transaction, error) {
	return _Registry.Contract.SetAftermarketDeviceNodeType(&_Registry.TransactOpts, label)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string _baseURI) returns()
func (_Registry *RegistryTransactor) SetBaseURI(opts *bind.TransactOpts, _baseURI string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setBaseURI", _baseURI)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string _baseURI) returns()
func (_Registry *RegistrySession) SetBaseURI(_baseURI string) (*types.Transaction, error) {
	return _Registry.Contract.SetBaseURI(&_Registry.TransactOpts, _baseURI)
}

// SetBaseURI is a paid mutator transaction binding the contract method 0x55f804b3.
//
// Solidity: function setBaseURI(string _baseURI) returns()
func (_Registry *RegistryTransactorSession) SetBaseURI(_baseURI string) (*types.Transaction, error) {
	return _Registry.Contract.SetBaseURI(&_Registry.TransactOpts, _baseURI)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Registry *RegistryTransactor) SetController(opts *bind.TransactOpts, _controller common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setController", _controller)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Registry *RegistrySession) SetController(_controller common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetController(&_Registry.TransactOpts, _controller)
}

// SetController is a paid mutator transaction binding the contract method 0x92eefe9b.
//
// Solidity: function setController(address _controller) returns()
func (_Registry *RegistryTransactorSession) SetController(_controller common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetController(&_Registry.TransactOpts, _controller)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Registry *RegistryTransactor) SetDimoToken(opts *bind.TransactOpts, _dimoToken common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setDimoToken", _dimoToken)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Registry *RegistrySession) SetDimoToken(_dimoToken common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetDimoToken(&_Registry.TransactOpts, _dimoToken)
}

// SetDimoToken is a paid mutator transaction binding the contract method 0x5b6c1979.
//
// Solidity: function setDimoToken(address _dimoToken) returns()
func (_Registry *RegistryTransactorSession) SetDimoToken(_dimoToken common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetDimoToken(&_Registry.TransactOpts, _dimoToken)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Registry *RegistryTransactor) SetFoundationAddress(opts *bind.TransactOpts, _foundation common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setFoundationAddress", _foundation)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Registry *RegistrySession) SetFoundationAddress(_foundation common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetFoundationAddress(&_Registry.TransactOpts, _foundation)
}

// SetFoundationAddress is a paid mutator transaction binding the contract method 0xf41377ca.
//
// Solidity: function setFoundationAddress(address _foundation) returns()
func (_Registry *RegistryTransactorSession) SetFoundationAddress(_foundation common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetFoundationAddress(&_Registry.TransactOpts, _foundation)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Registry *RegistryTransactor) SetLicense(opts *bind.TransactOpts, _license common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setLicense", _license)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Registry *RegistrySession) SetLicense(_license common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetLicense(&_Registry.TransactOpts, _license)
}

// SetLicense is a paid mutator transaction binding the contract method 0x0fd21c17.
//
// Solidity: function setLicense(address _license) returns()
func (_Registry *RegistryTransactorSession) SetLicense(_license common.Address) (*types.Transaction, error) {
	return _Registry.Contract.SetLicense(&_Registry.TransactOpts, _license)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0xd89e7dbc.
//
// Solidity: function setManufacturerInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactor) SetManufacturerInfo(opts *bind.TransactOpts, nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setManufacturerInfo", nodeId, attributes, infos)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0xd89e7dbc.
//
// Solidity: function setManufacturerInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistrySession) SetManufacturerInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerInfo(&_Registry.TransactOpts, nodeId, attributes, infos)
}

// SetManufacturerInfo is a paid mutator transaction binding the contract method 0xd89e7dbc.
//
// Solidity: function setManufacturerInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactorSession) SetManufacturerInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerInfo(&_Registry.TransactOpts, nodeId, attributes, infos)
}

// SetManufacturerNodeType is a paid mutator transaction binding the contract method 0xca9ba80e.
//
// Solidity: function setManufacturerNodeType(bytes label) returns()
func (_Registry *RegistryTransactor) SetManufacturerNodeType(opts *bind.TransactOpts, label []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setManufacturerNodeType", label)
}

// SetManufacturerNodeType is a paid mutator transaction binding the contract method 0xca9ba80e.
//
// Solidity: function setManufacturerNodeType(bytes label) returns()
func (_Registry *RegistrySession) SetManufacturerNodeType(label []byte) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerNodeType(&_Registry.TransactOpts, label)
}

// SetManufacturerNodeType is a paid mutator transaction binding the contract method 0xca9ba80e.
//
// Solidity: function setManufacturerNodeType(bytes label) returns()
func (_Registry *RegistryTransactorSession) SetManufacturerNodeType(label []byte) (*types.Transaction, error) {
	return _Registry.Contract.SetManufacturerNodeType(&_Registry.TransactOpts, label)
}

// SetTokenURI is a paid mutator transaction binding the contract method 0x162094c4.
//
// Solidity: function setTokenURI(uint256 tokenId, string _tokenURI) returns()
func (_Registry *RegistryTransactor) SetTokenURI(opts *bind.TransactOpts, tokenId *big.Int, _tokenURI string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setTokenURI", tokenId, _tokenURI)
}

// SetTokenURI is a paid mutator transaction binding the contract method 0x162094c4.
//
// Solidity: function setTokenURI(uint256 tokenId, string _tokenURI) returns()
func (_Registry *RegistrySession) SetTokenURI(tokenId *big.Int, _tokenURI string) (*types.Transaction, error) {
	return _Registry.Contract.SetTokenURI(&_Registry.TransactOpts, tokenId, _tokenURI)
}

// SetTokenURI is a paid mutator transaction binding the contract method 0x162094c4.
//
// Solidity: function setTokenURI(uint256 tokenId, string _tokenURI) returns()
func (_Registry *RegistryTransactorSession) SetTokenURI(tokenId *big.Int, _tokenURI string) (*types.Transaction, error) {
	return _Registry.Contract.SetTokenURI(&_Registry.TransactOpts, tokenId, _tokenURI)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xc175eb46.
//
// Solidity: function setVehicleInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactor) SetVehicleInfo(opts *bind.TransactOpts, nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setVehicleInfo", nodeId, attributes, infos)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xc175eb46.
//
// Solidity: function setVehicleInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistrySession) SetVehicleInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.SetVehicleInfo(&_Registry.TransactOpts, nodeId, attributes, infos)
}

// SetVehicleInfo is a paid mutator transaction binding the contract method 0xc175eb46.
//
// Solidity: function setVehicleInfo(uint256 nodeId, string[] attributes, string[] infos) returns()
func (_Registry *RegistryTransactorSession) SetVehicleInfo(nodeId *big.Int, attributes []string, infos []string) (*types.Transaction, error) {
	return _Registry.Contract.SetVehicleInfo(&_Registry.TransactOpts, nodeId, attributes, infos)
}

// SetVehicleNodeType is a paid mutator transaction binding the contract method 0x63822b13.
//
// Solidity: function setVehicleNodeType(bytes label) returns()
func (_Registry *RegistryTransactor) SetVehicleNodeType(opts *bind.TransactOpts, label []byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "setVehicleNodeType", label)
}

// SetVehicleNodeType is a paid mutator transaction binding the contract method 0x63822b13.
//
// Solidity: function setVehicleNodeType(bytes label) returns()
func (_Registry *RegistrySession) SetVehicleNodeType(label []byte) (*types.Transaction, error) {
	return _Registry.Contract.SetVehicleNodeType(&_Registry.TransactOpts, label)
}

// SetVehicleNodeType is a paid mutator transaction binding the contract method 0x63822b13.
//
// Solidity: function setVehicleNodeType(bytes label) returns()
func (_Registry *RegistryTransactorSession) SetVehicleNodeType(label []byte) (*types.Transaction, error) {
	return _Registry.Contract.SetVehicleNodeType(&_Registry.TransactOpts, label)
}

// TransferAftermarketDeviceOwnership is a paid mutator transaction binding the contract method 0xff96b761.
//
// Solidity: function transferAftermarketDeviceOwnership(uint256 aftermarketDeviceNode, address newOwner) returns()
func (_Registry *RegistryTransactor) TransferAftermarketDeviceOwnership(opts *bind.TransactOpts, aftermarketDeviceNode *big.Int, newOwner common.Address) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "transferAftermarketDeviceOwnership", aftermarketDeviceNode, newOwner)
}

// TransferAftermarketDeviceOwnership is a paid mutator transaction binding the contract method 0xff96b761.
//
// Solidity: function transferAftermarketDeviceOwnership(uint256 aftermarketDeviceNode, address newOwner) returns()
func (_Registry *RegistrySession) TransferAftermarketDeviceOwnership(aftermarketDeviceNode *big.Int, newOwner common.Address) (*types.Transaction, error) {
	return _Registry.Contract.TransferAftermarketDeviceOwnership(&_Registry.TransactOpts, aftermarketDeviceNode, newOwner)
}

// TransferAftermarketDeviceOwnership is a paid mutator transaction binding the contract method 0xff96b761.
//
// Solidity: function transferAftermarketDeviceOwnership(uint256 aftermarketDeviceNode, address newOwner) returns()
func (_Registry *RegistryTransactorSession) TransferAftermarketDeviceOwnership(aftermarketDeviceNode *big.Int, newOwner common.Address) (*types.Transaction, error) {
	return _Registry.Contract.TransferAftermarketDeviceOwnership(&_Registry.TransactOpts, aftermarketDeviceNode, newOwner)
}

// UnpairAftermarketDeviceByDeviceNode is a paid mutator transaction binding the contract method 0x71193956.
//
// Solidity: function unpairAftermarketDeviceByDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistryTransactor) UnpairAftermarketDeviceByDeviceNode(opts *bind.TransactOpts, aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "unpairAftermarketDeviceByDeviceNode", aftermarketDeviceNodes)
}

// UnpairAftermarketDeviceByDeviceNode is a paid mutator transaction binding the contract method 0x71193956.
//
// Solidity: function unpairAftermarketDeviceByDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistrySession) UnpairAftermarketDeviceByDeviceNode(aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceByDeviceNode(&_Registry.TransactOpts, aftermarketDeviceNodes)
}

// UnpairAftermarketDeviceByDeviceNode is a paid mutator transaction binding the contract method 0x71193956.
//
// Solidity: function unpairAftermarketDeviceByDeviceNode(uint256[] aftermarketDeviceNodes) returns()
func (_Registry *RegistryTransactorSession) UnpairAftermarketDeviceByDeviceNode(aftermarketDeviceNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceByDeviceNode(&_Registry.TransactOpts, aftermarketDeviceNodes)
}

// UnpairAftermarketDeviceByVehicleNode is a paid mutator transaction binding the contract method 0x8c2ee9bb.
//
// Solidity: function unpairAftermarketDeviceByVehicleNode(uint256[] vehicleNodes) returns()
func (_Registry *RegistryTransactor) UnpairAftermarketDeviceByVehicleNode(opts *bind.TransactOpts, vehicleNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "unpairAftermarketDeviceByVehicleNode", vehicleNodes)
}

// UnpairAftermarketDeviceByVehicleNode is a paid mutator transaction binding the contract method 0x8c2ee9bb.
//
// Solidity: function unpairAftermarketDeviceByVehicleNode(uint256[] vehicleNodes) returns()
func (_Registry *RegistrySession) UnpairAftermarketDeviceByVehicleNode(vehicleNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceByVehicleNode(&_Registry.TransactOpts, vehicleNodes)
}

// UnpairAftermarketDeviceByVehicleNode is a paid mutator transaction binding the contract method 0x8c2ee9bb.
//
// Solidity: function unpairAftermarketDeviceByVehicleNode(uint256[] vehicleNodes) returns()
func (_Registry *RegistryTransactorSession) UnpairAftermarketDeviceByVehicleNode(vehicleNodes []*big.Int) (*types.Transaction, error) {
	return _Registry.Contract.UnpairAftermarketDeviceByVehicleNode(&_Registry.TransactOpts, vehicleNodes)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Registry *RegistryTransactor) UpdateModule(opts *bind.TransactOpts, oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Registry.contract.Transact(opts, "updateModule", oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Registry *RegistrySession) UpdateModule(oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.UpdateModule(&_Registry.TransactOpts, oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// UpdateModule is a paid mutator transaction binding the contract method 0x06d1d2a1.
//
// Solidity: function updateModule(address oldImplementation, address newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors) returns()
func (_Registry *RegistryTransactorSession) UpdateModule(oldImplementation common.Address, newImplementation common.Address, oldSelectors [][4]byte, newSelectors [][4]byte) (*types.Transaction, error) {
	return _Registry.Contract.UpdateModule(&_Registry.TransactOpts, oldImplementation, newImplementation, oldSelectors, newSelectors)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Registry *RegistryTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Registry.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Registry *RegistrySession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Registry.Contract.Fallback(&_Registry.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_Registry *RegistryTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Registry.Contract.Fallback(&_Registry.TransactOpts, calldata)
}

// RegistryAftermarketDeviceClaimedIterator is returned from FilterAftermarketDeviceClaimed and is used to iterate over the raw logs and unpacked data for AftermarketDeviceClaimed events raised by the Registry contract.
type RegistryAftermarketDeviceClaimedIterator struct {
	Event *RegistryAftermarketDeviceClaimed // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceClaimed)
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
		it.Event = new(RegistryAftermarketDeviceClaimed)
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
func (it *RegistryAftermarketDeviceClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceClaimed represents a AftermarketDeviceClaimed event raised by the Registry contract.
type RegistryAftermarketDeviceClaimed struct {
	AftermarketDeviceNode *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceClaimed is a free log retrieval operation binding the contract event 0x8468d811e5090d3b1a07e28af524e66c128f624e16b07638f419012c779f76ec.
//
// Solidity: event AftermarketDeviceClaimed(uint256 indexed aftermarketDeviceNode, address indexed owner)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceClaimed(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int, owner []common.Address) (*RegistryAftermarketDeviceClaimedIterator, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceClaimed", aftermarketDeviceNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceClaimedIterator{contract: _Registry.contract, event: "AftermarketDeviceClaimed", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceClaimed is a free log subscription operation binding the contract event 0x8468d811e5090d3b1a07e28af524e66c128f624e16b07638f419012c779f76ec.
//
// Solidity: event AftermarketDeviceClaimed(uint256 indexed aftermarketDeviceNode, address indexed owner)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceClaimed(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceClaimed, aftermarketDeviceNode []*big.Int, owner []common.Address) (event.Subscription, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceClaimed", aftermarketDeviceNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceClaimed)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceClaimed", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseAftermarketDeviceClaimed(log types.Log) (*RegistryAftermarketDeviceClaimed, error) {
	event := new(RegistryAftermarketDeviceClaimed)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceNodeMintedIterator is returned from FilterAftermarketDeviceNodeMinted and is used to iterate over the raw logs and unpacked data for AftermarketDeviceNodeMinted events raised by the Registry contract.
type RegistryAftermarketDeviceNodeMintedIterator struct {
	Event *RegistryAftermarketDeviceNodeMinted // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceNodeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceNodeMinted)
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
		it.Event = new(RegistryAftermarketDeviceNodeMinted)
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
func (it *RegistryAftermarketDeviceNodeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceNodeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceNodeMinted represents a AftermarketDeviceNodeMinted event raised by the Registry contract.
type RegistryAftermarketDeviceNodeMinted struct {
	NodeType                 *big.Int
	NodeId                   *big.Int
	AftermarketDeviceAddress common.Address
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceNodeMinted is a free log retrieval operation binding the contract event 0x7856cab6942610b9c833a5d8a87a2c88deb168f56f6da3599900da04c13705e9.
//
// Solidity: event AftermarketDeviceNodeMinted(uint256 indexed nodeType, uint256 indexed nodeId, address indexed aftermarketDeviceAddress)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceNodeMinted(opts *bind.FilterOpts, nodeType []*big.Int, nodeId []*big.Int, aftermarketDeviceAddress []common.Address) (*RegistryAftermarketDeviceNodeMintedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceNodeMinted", nodeTypeRule, nodeIdRule, aftermarketDeviceAddressRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceNodeMintedIterator{contract: _Registry.contract, event: "AftermarketDeviceNodeMinted", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceNodeMinted is a free log subscription operation binding the contract event 0x7856cab6942610b9c833a5d8a87a2c88deb168f56f6da3599900da04c13705e9.
//
// Solidity: event AftermarketDeviceNodeMinted(uint256 indexed nodeType, uint256 indexed nodeId, address indexed aftermarketDeviceAddress)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceNodeMinted(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceNodeMinted, nodeType []*big.Int, nodeId []*big.Int, aftermarketDeviceAddress []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceNodeMinted", nodeTypeRule, nodeIdRule, aftermarketDeviceAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceNodeMinted)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceNodeMinted", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseAftermarketDeviceNodeMinted(log types.Log) (*RegistryAftermarketDeviceNodeMinted, error) {
	event := new(RegistryAftermarketDeviceNodeMinted)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceNodeMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDevicePairedIterator is returned from FilterAftermarketDevicePaired and is used to iterate over the raw logs and unpacked data for AftermarketDevicePaired events raised by the Registry contract.
type RegistryAftermarketDevicePairedIterator struct {
	Event *RegistryAftermarketDevicePaired // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDevicePairedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDevicePaired)
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
		it.Event = new(RegistryAftermarketDevicePaired)
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
func (it *RegistryAftermarketDevicePairedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDevicePairedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDevicePaired represents a AftermarketDevicePaired event raised by the Registry contract.
type RegistryAftermarketDevicePaired struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDevicePaired is a free log retrieval operation binding the contract event 0x89ec132808bbf01af00b90fd34e04fd6cfb8dba2813ca5446a415500b83c7938.
//
// Solidity: event AftermarketDevicePaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) FilterAftermarketDevicePaired(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int, vehicleNode []*big.Int, owner []common.Address) (*RegistryAftermarketDevicePairedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDevicePaired", aftermarketDeviceNodeRule, vehicleNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDevicePairedIterator{contract: _Registry.contract, event: "AftermarketDevicePaired", logs: logs, sub: sub}, nil
}

// WatchAftermarketDevicePaired is a free log subscription operation binding the contract event 0x89ec132808bbf01af00b90fd34e04fd6cfb8dba2813ca5446a415500b83c7938.
//
// Solidity: event AftermarketDevicePaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) WatchAftermarketDevicePaired(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDevicePaired, aftermarketDeviceNode []*big.Int, vehicleNode []*big.Int, owner []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDevicePaired", aftermarketDeviceNodeRule, vehicleNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDevicePaired)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDevicePaired", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseAftermarketDevicePaired(log types.Log) (*RegistryAftermarketDevicePaired, error) {
	event := new(RegistryAftermarketDevicePaired)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDevicePaired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceTransferredIterator is returned from FilterAftermarketDeviceTransferred and is used to iterate over the raw logs and unpacked data for AftermarketDeviceTransferred events raised by the Registry contract.
type RegistryAftermarketDeviceTransferredIterator struct {
	Event *RegistryAftermarketDeviceTransferred // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceTransferred)
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
		it.Event = new(RegistryAftermarketDeviceTransferred)
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
func (it *RegistryAftermarketDeviceTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceTransferred represents a AftermarketDeviceTransferred event raised by the Registry contract.
type RegistryAftermarketDeviceTransferred struct {
	AftermarketDeviceNode *big.Int
	OldOwner              common.Address
	NewOwner              common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceTransferred is a free log retrieval operation binding the contract event 0x1d2e88640b58e7fc67878851d97e2cfae3bc7eb7db3226dec94b1c499d631637.
//
// Solidity: event AftermarketDeviceTransferred(uint256 indexed aftermarketDeviceNode, address indexed oldOwner, address indexed newOwner)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceTransferred(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int, oldOwner []common.Address, newOwner []common.Address) (*RegistryAftermarketDeviceTransferredIterator, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var oldOwnerRule []interface{}
	for _, oldOwnerItem := range oldOwner {
		oldOwnerRule = append(oldOwnerRule, oldOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceTransferred", aftermarketDeviceNodeRule, oldOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceTransferredIterator{contract: _Registry.contract, event: "AftermarketDeviceTransferred", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceTransferred is a free log subscription operation binding the contract event 0x1d2e88640b58e7fc67878851d97e2cfae3bc7eb7db3226dec94b1c499d631637.
//
// Solidity: event AftermarketDeviceTransferred(uint256 indexed aftermarketDeviceNode, address indexed oldOwner, address indexed newOwner)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceTransferred(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceTransferred, aftermarketDeviceNode []*big.Int, oldOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var aftermarketDeviceNodeRule []interface{}
	for _, aftermarketDeviceNodeItem := range aftermarketDeviceNode {
		aftermarketDeviceNodeRule = append(aftermarketDeviceNodeRule, aftermarketDeviceNodeItem)
	}
	var oldOwnerRule []interface{}
	for _, oldOwnerItem := range oldOwner {
		oldOwnerRule = append(oldOwnerRule, oldOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceTransferred", aftermarketDeviceNodeRule, oldOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceTransferred)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceTransferred", log); err != nil {
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

// ParseAftermarketDeviceTransferred is a log parse operation binding the contract event 0x1d2e88640b58e7fc67878851d97e2cfae3bc7eb7db3226dec94b1c499d631637.
//
// Solidity: event AftermarketDeviceTransferred(uint256 indexed aftermarketDeviceNode, address indexed oldOwner, address indexed newOwner)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceTransferred(log types.Log) (*RegistryAftermarketDeviceTransferred, error) {
	event := new(RegistryAftermarketDeviceTransferred)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAftermarketDeviceUnpairedIterator is returned from FilterAftermarketDeviceUnpaired and is used to iterate over the raw logs and unpacked data for AftermarketDeviceUnpaired events raised by the Registry contract.
type RegistryAftermarketDeviceUnpairedIterator struct {
	Event *RegistryAftermarketDeviceUnpaired // Event containing the contract specifics and raw log

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
func (it *RegistryAftermarketDeviceUnpairedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAftermarketDeviceUnpaired)
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
		it.Event = new(RegistryAftermarketDeviceUnpaired)
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
func (it *RegistryAftermarketDeviceUnpairedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAftermarketDeviceUnpairedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAftermarketDeviceUnpaired represents a AftermarketDeviceUnpaired event raised by the Registry contract.
type RegistryAftermarketDeviceUnpaired struct {
	AftermarketDeviceNode *big.Int
	VehicleNode           *big.Int
	Owner                 common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterAftermarketDeviceUnpaired is a free log retrieval operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) FilterAftermarketDeviceUnpaired(opts *bind.FilterOpts, aftermarketDeviceNode []*big.Int, vehicleNode []*big.Int, owner []common.Address) (*RegistryAftermarketDeviceUnpairedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AftermarketDeviceUnpaired", aftermarketDeviceNodeRule, vehicleNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAftermarketDeviceUnpairedIterator{contract: _Registry.contract, event: "AftermarketDeviceUnpaired", logs: logs, sub: sub}, nil
}

// WatchAftermarketDeviceUnpaired is a free log subscription operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) WatchAftermarketDeviceUnpaired(opts *bind.WatchOpts, sink chan<- *RegistryAftermarketDeviceUnpaired, aftermarketDeviceNode []*big.Int, vehicleNode []*big.Int, owner []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AftermarketDeviceUnpaired", aftermarketDeviceNodeRule, vehicleNodeRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAftermarketDeviceUnpaired)
				if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceUnpaired", log); err != nil {
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

// ParseAftermarketDeviceUnpaired is a log parse operation binding the contract event 0xd9135724aa6cdaa5b3ea73e3e0d74cb1a3a6d3cddcb9d58583f05f17bac82a8e.
//
// Solidity: event AftermarketDeviceUnpaired(uint256 indexed aftermarketDeviceNode, uint256 indexed vehicleNode, address indexed owner)
func (_Registry *RegistryFilterer) ParseAftermarketDeviceUnpaired(log types.Log) (*RegistryAftermarketDeviceUnpaired, error) {
	event := new(RegistryAftermarketDeviceUnpaired)
	if err := _Registry.contract.UnpackLog(event, "AftermarketDeviceUnpaired", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Registry contract.
type RegistryApprovalIterator struct {
	Event *RegistryApproval // Event containing the contract specifics and raw log

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
func (it *RegistryApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryApproval)
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
		it.Event = new(RegistryApproval)
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
func (it *RegistryApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryApproval represents a Approval event raised by the Registry contract.
type RegistryApproval struct {
	Owner    common.Address
	Operator common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed operator, uint256 indexed tokenId)
func (_Registry *RegistryFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, operator []common.Address, tokenId []*big.Int) (*RegistryApprovalIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "Approval", ownerRule, operatorRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &RegistryApprovalIterator{contract: _Registry.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed operator, uint256 indexed tokenId)
func (_Registry *RegistryFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *RegistryApproval, owner []common.Address, operator []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "Approval", ownerRule, operatorRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryApproval)
				if err := _Registry.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseApproval(log types.Log) (*RegistryApproval, error) {
	event := new(RegistryApproval)
	if err := _Registry.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the Registry contract.
type RegistryApprovalForAllIterator struct {
	Event *RegistryApprovalForAll // Event containing the contract specifics and raw log

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
func (it *RegistryApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryApprovalForAll)
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
		it.Event = new(RegistryApprovalForAll)
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
func (it *RegistryApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryApprovalForAll represents a ApprovalForAll event raised by the Registry contract.
type RegistryApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Registry *RegistryFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*RegistryApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &RegistryApprovalForAllIterator{contract: _Registry.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_Registry *RegistryFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *RegistryApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryApprovalForAll)
				if err := _Registry.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseApprovalForAll(log types.Log) (*RegistryApprovalForAll, error) {
	event := new(RegistryApprovalForAll)
	if err := _Registry.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryAttributeAddedIterator is returned from FilterAttributeAdded and is used to iterate over the raw logs and unpacked data for AttributeAdded events raised by the Registry contract.
type RegistryAttributeAddedIterator struct {
	Event *RegistryAttributeAdded // Event containing the contract specifics and raw log

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
func (it *RegistryAttributeAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryAttributeAdded)
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
		it.Event = new(RegistryAttributeAdded)
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
func (it *RegistryAttributeAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryAttributeAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryAttributeAdded represents a AttributeAdded event raised by the Registry contract.
type RegistryAttributeAdded struct {
	NodeType  *big.Int
	Attribute common.Hash
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAttributeAdded is a free log retrieval operation binding the contract event 0xdee1f2fc87d9c834bee1095ebfc0b81ae1b364a7c74060167ab8a82623b22f9c.
//
// Solidity: event AttributeAdded(uint256 indexed nodeType, string indexed attribute)
func (_Registry *RegistryFilterer) FilterAttributeAdded(opts *bind.FilterOpts, nodeType []*big.Int, attribute []string) (*RegistryAttributeAddedIterator, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var attributeRule []interface{}
	for _, attributeItem := range attribute {
		attributeRule = append(attributeRule, attributeItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "AttributeAdded", nodeTypeRule, attributeRule)
	if err != nil {
		return nil, err
	}
	return &RegistryAttributeAddedIterator{contract: _Registry.contract, event: "AttributeAdded", logs: logs, sub: sub}, nil
}

// WatchAttributeAdded is a free log subscription operation binding the contract event 0xdee1f2fc87d9c834bee1095ebfc0b81ae1b364a7c74060167ab8a82623b22f9c.
//
// Solidity: event AttributeAdded(uint256 indexed nodeType, string indexed attribute)
func (_Registry *RegistryFilterer) WatchAttributeAdded(opts *bind.WatchOpts, sink chan<- *RegistryAttributeAdded, nodeType []*big.Int, attribute []string) (event.Subscription, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var attributeRule []interface{}
	for _, attributeItem := range attribute {
		attributeRule = append(attributeRule, attributeItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "AttributeAdded", nodeTypeRule, attributeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryAttributeAdded)
				if err := _Registry.contract.UnpackLog(event, "AttributeAdded", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseAttributeAdded(log types.Log) (*RegistryAttributeAdded, error) {
	event := new(RegistryAttributeAdded)
	if err := _Registry.contract.UnpackLog(event, "AttributeAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryControllerSetIterator is returned from FilterControllerSet and is used to iterate over the raw logs and unpacked data for ControllerSet events raised by the Registry contract.
type RegistryControllerSetIterator struct {
	Event *RegistryControllerSet // Event containing the contract specifics and raw log

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
func (it *RegistryControllerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryControllerSet)
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
		it.Event = new(RegistryControllerSet)
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
func (it *RegistryControllerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryControllerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryControllerSet represents a ControllerSet event raised by the Registry contract.
type RegistryControllerSet struct {
	Controller common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterControllerSet is a free log retrieval operation binding the contract event 0x79f74fd5964b6943d8a1865abfb7f668c92fa3f32c0a2e3195da7d0946703ad7.
//
// Solidity: event ControllerSet(address indexed controller)
func (_Registry *RegistryFilterer) FilterControllerSet(opts *bind.FilterOpts, controller []common.Address) (*RegistryControllerSetIterator, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ControllerSet", controllerRule)
	if err != nil {
		return nil, err
	}
	return &RegistryControllerSetIterator{contract: _Registry.contract, event: "ControllerSet", logs: logs, sub: sub}, nil
}

// WatchControllerSet is a free log subscription operation binding the contract event 0x79f74fd5964b6943d8a1865abfb7f668c92fa3f32c0a2e3195da7d0946703ad7.
//
// Solidity: event ControllerSet(address indexed controller)
func (_Registry *RegistryFilterer) WatchControllerSet(opts *bind.WatchOpts, sink chan<- *RegistryControllerSet, controller []common.Address) (event.Subscription, error) {

	var controllerRule []interface{}
	for _, controllerItem := range controller {
		controllerRule = append(controllerRule, controllerItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ControllerSet", controllerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryControllerSet)
				if err := _Registry.contract.UnpackLog(event, "ControllerSet", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseControllerSet(log types.Log) (*RegistryControllerSet, error) {
	event := new(RegistryControllerSet)
	if err := _Registry.contract.UnpackLog(event, "ControllerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryModuleAddedIterator is returned from FilterModuleAdded and is used to iterate over the raw logs and unpacked data for ModuleAdded events raised by the Registry contract.
type RegistryModuleAddedIterator struct {
	Event *RegistryModuleAdded // Event containing the contract specifics and raw log

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
func (it *RegistryModuleAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryModuleAdded)
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
		it.Event = new(RegistryModuleAdded)
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
func (it *RegistryModuleAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryModuleAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryModuleAdded represents a ModuleAdded event raised by the Registry contract.
type RegistryModuleAdded struct {
	ModuleAddr common.Address
	Selectors  [][4]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterModuleAdded is a free log retrieval operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_Registry *RegistryFilterer) FilterModuleAdded(opts *bind.FilterOpts, moduleAddr []common.Address) (*RegistryModuleAddedIterator, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ModuleAdded", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return &RegistryModuleAddedIterator{contract: _Registry.contract, event: "ModuleAdded", logs: logs, sub: sub}, nil
}

// WatchModuleAdded is a free log subscription operation binding the contract event 0x02d0c334c706cd2f08faf7bc03674fc7f3970dd8921776c655069cde33b7fb29.
//
// Solidity: event ModuleAdded(address indexed moduleAddr, bytes4[] selectors)
func (_Registry *RegistryFilterer) WatchModuleAdded(opts *bind.WatchOpts, sink chan<- *RegistryModuleAdded, moduleAddr []common.Address) (event.Subscription, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ModuleAdded", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryModuleAdded)
				if err := _Registry.contract.UnpackLog(event, "ModuleAdded", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseModuleAdded(log types.Log) (*RegistryModuleAdded, error) {
	event := new(RegistryModuleAdded)
	if err := _Registry.contract.UnpackLog(event, "ModuleAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryModuleRemovedIterator is returned from FilterModuleRemoved and is used to iterate over the raw logs and unpacked data for ModuleRemoved events raised by the Registry contract.
type RegistryModuleRemovedIterator struct {
	Event *RegistryModuleRemoved // Event containing the contract specifics and raw log

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
func (it *RegistryModuleRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryModuleRemoved)
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
		it.Event = new(RegistryModuleRemoved)
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
func (it *RegistryModuleRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryModuleRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryModuleRemoved represents a ModuleRemoved event raised by the Registry contract.
type RegistryModuleRemoved struct {
	ModuleAddr common.Address
	Selectors  [][4]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterModuleRemoved is a free log retrieval operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_Registry *RegistryFilterer) FilterModuleRemoved(opts *bind.FilterOpts, moduleAddr []common.Address) (*RegistryModuleRemovedIterator, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ModuleRemoved", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return &RegistryModuleRemovedIterator{contract: _Registry.contract, event: "ModuleRemoved", logs: logs, sub: sub}, nil
}

// WatchModuleRemoved is a free log subscription operation binding the contract event 0x7c3eb4f9083f75cbed2bd3f703e24b4bbcb77d345d3c50945f3abf3e967755cb.
//
// Solidity: event ModuleRemoved(address indexed moduleAddr, bytes4[] selectors)
func (_Registry *RegistryFilterer) WatchModuleRemoved(opts *bind.WatchOpts, sink chan<- *RegistryModuleRemoved, moduleAddr []common.Address) (event.Subscription, error) {

	var moduleAddrRule []interface{}
	for _, moduleAddrItem := range moduleAddr {
		moduleAddrRule = append(moduleAddrRule, moduleAddrItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ModuleRemoved", moduleAddrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryModuleRemoved)
				if err := _Registry.contract.UnpackLog(event, "ModuleRemoved", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseModuleRemoved(log types.Log) (*RegistryModuleRemoved, error) {
	event := new(RegistryModuleRemoved)
	if err := _Registry.contract.UnpackLog(event, "ModuleRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryModuleUpdatedIterator is returned from FilterModuleUpdated and is used to iterate over the raw logs and unpacked data for ModuleUpdated events raised by the Registry contract.
type RegistryModuleUpdatedIterator struct {
	Event *RegistryModuleUpdated // Event containing the contract specifics and raw log

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
func (it *RegistryModuleUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryModuleUpdated)
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
		it.Event = new(RegistryModuleUpdated)
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
func (it *RegistryModuleUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryModuleUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryModuleUpdated represents a ModuleUpdated event raised by the Registry contract.
type RegistryModuleUpdated struct {
	OldImplementation common.Address
	NewImplementation common.Address
	OldSelectors      [][4]byte
	NewSelectors      [][4]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterModuleUpdated is a free log retrieval operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_Registry *RegistryFilterer) FilterModuleUpdated(opts *bind.FilterOpts, oldImplementation []common.Address, newImplementation []common.Address) (*RegistryModuleUpdatedIterator, error) {

	var oldImplementationRule []interface{}
	for _, oldImplementationItem := range oldImplementation {
		oldImplementationRule = append(oldImplementationRule, oldImplementationItem)
	}
	var newImplementationRule []interface{}
	for _, newImplementationItem := range newImplementation {
		newImplementationRule = append(newImplementationRule, newImplementationItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "ModuleUpdated", oldImplementationRule, newImplementationRule)
	if err != nil {
		return nil, err
	}
	return &RegistryModuleUpdatedIterator{contract: _Registry.contract, event: "ModuleUpdated", logs: logs, sub: sub}, nil
}

// WatchModuleUpdated is a free log subscription operation binding the contract event 0xa062c2c046aa14dc9284b13bde77061cb034f0aa820f20057af6b164651eaa08.
//
// Solidity: event ModuleUpdated(address indexed oldImplementation, address indexed newImplementation, bytes4[] oldSelectors, bytes4[] newSelectors)
func (_Registry *RegistryFilterer) WatchModuleUpdated(opts *bind.WatchOpts, sink chan<- *RegistryModuleUpdated, oldImplementation []common.Address, newImplementation []common.Address) (event.Subscription, error) {

	var oldImplementationRule []interface{}
	for _, oldImplementationItem := range oldImplementation {
		oldImplementationRule = append(oldImplementationRule, oldImplementationItem)
	}
	var newImplementationRule []interface{}
	for _, newImplementationItem := range newImplementation {
		newImplementationRule = append(newImplementationRule, newImplementationItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "ModuleUpdated", oldImplementationRule, newImplementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryModuleUpdated)
				if err := _Registry.contract.UnpackLog(event, "ModuleUpdated", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseModuleUpdated(log types.Log) (*RegistryModuleUpdated, error) {
	event := new(RegistryModuleUpdated)
	if err := _Registry.contract.UnpackLog(event, "ModuleUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryNodeMintedIterator is returned from FilterNodeMinted and is used to iterate over the raw logs and unpacked data for NodeMinted events raised by the Registry contract.
type RegistryNodeMintedIterator struct {
	Event *RegistryNodeMinted // Event containing the contract specifics and raw log

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
func (it *RegistryNodeMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryNodeMinted)
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
		it.Event = new(RegistryNodeMinted)
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
func (it *RegistryNodeMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryNodeMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryNodeMinted represents a NodeMinted event raised by the Registry contract.
type RegistryNodeMinted struct {
	NodeType *big.Int
	NodeId   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNodeMinted is a free log retrieval operation binding the contract event 0x0c2616265c4fd089569533525abc7b19b9f82b423d7cdb61801490b8f9e0ce59.
//
// Solidity: event NodeMinted(uint256 indexed nodeType, uint256 indexed nodeId)
func (_Registry *RegistryFilterer) FilterNodeMinted(opts *bind.FilterOpts, nodeType []*big.Int, nodeId []*big.Int) (*RegistryNodeMintedIterator, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var nodeIdRule []interface{}
	for _, nodeIdItem := range nodeId {
		nodeIdRule = append(nodeIdRule, nodeIdItem)
	}

	logs, sub, err := _Registry.contract.FilterLogs(opts, "NodeMinted", nodeTypeRule, nodeIdRule)
	if err != nil {
		return nil, err
	}
	return &RegistryNodeMintedIterator{contract: _Registry.contract, event: "NodeMinted", logs: logs, sub: sub}, nil
}

// WatchNodeMinted is a free log subscription operation binding the contract event 0x0c2616265c4fd089569533525abc7b19b9f82b423d7cdb61801490b8f9e0ce59.
//
// Solidity: event NodeMinted(uint256 indexed nodeType, uint256 indexed nodeId)
func (_Registry *RegistryFilterer) WatchNodeMinted(opts *bind.WatchOpts, sink chan<- *RegistryNodeMinted, nodeType []*big.Int, nodeId []*big.Int) (event.Subscription, error) {

	var nodeTypeRule []interface{}
	for _, nodeTypeItem := range nodeType {
		nodeTypeRule = append(nodeTypeRule, nodeTypeItem)
	}
	var nodeIdRule []interface{}
	for _, nodeIdItem := range nodeId {
		nodeIdRule = append(nodeIdRule, nodeIdItem)
	}

	logs, sub, err := _Registry.contract.WatchLogs(opts, "NodeMinted", nodeTypeRule, nodeIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryNodeMinted)
				if err := _Registry.contract.UnpackLog(event, "NodeMinted", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseNodeMinted(log types.Log) (*RegistryNodeMinted, error) {
	event := new(RegistryNodeMinted)
	if err := _Registry.contract.UnpackLog(event, "NodeMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Registry contract.
type RegistryRoleAdminChangedIterator struct {
	Event *RegistryRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *RegistryRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryRoleAdminChanged)
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
		it.Event = new(RegistryRoleAdminChanged)
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
func (it *RegistryRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryRoleAdminChanged represents a RoleAdminChanged event raised by the Registry contract.
type RegistryRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Registry *RegistryFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*RegistryRoleAdminChangedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &RegistryRoleAdminChangedIterator{contract: _Registry.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Registry *RegistryFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *RegistryRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryRoleAdminChanged)
				if err := _Registry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseRoleAdminChanged(log types.Log) (*RegistryRoleAdminChanged, error) {
	event := new(RegistryRoleAdminChanged)
	if err := _Registry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Registry contract.
type RegistryRoleGrantedIterator struct {
	Event *RegistryRoleGranted // Event containing the contract specifics and raw log

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
func (it *RegistryRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryRoleGranted)
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
		it.Event = new(RegistryRoleGranted)
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
func (it *RegistryRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryRoleGranted represents a RoleGranted event raised by the Registry contract.
type RegistryRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Registry *RegistryFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*RegistryRoleGrantedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &RegistryRoleGrantedIterator{contract: _Registry.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Registry *RegistryFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *RegistryRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryRoleGranted)
				if err := _Registry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseRoleGranted(log types.Log) (*RegistryRoleGranted, error) {
	event := new(RegistryRoleGranted)
	if err := _Registry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Registry contract.
type RegistryRoleRevokedIterator struct {
	Event *RegistryRoleRevoked // Event containing the contract specifics and raw log

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
func (it *RegistryRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryRoleRevoked)
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
		it.Event = new(RegistryRoleRevoked)
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
func (it *RegistryRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryRoleRevoked represents a RoleRevoked event raised by the Registry contract.
type RegistryRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Registry *RegistryFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*RegistryRoleRevokedIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &RegistryRoleRevokedIterator{contract: _Registry.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Registry *RegistryFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *RegistryRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryRoleRevoked)
				if err := _Registry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseRoleRevoked(log types.Log) (*RegistryRoleRevoked, error) {
	event := new(RegistryRoleRevoked)
	if err := _Registry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RegistryTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Registry contract.
type RegistryTransferIterator struct {
	Event *RegistryTransfer // Event containing the contract specifics and raw log

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
func (it *RegistryTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RegistryTransfer)
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
		it.Event = new(RegistryTransfer)
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
func (it *RegistryTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RegistryTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RegistryTransfer represents a Transfer event raised by the Registry contract.
type RegistryTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Registry *RegistryFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*RegistryTransferIterator, error) {

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

	logs, sub, err := _Registry.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &RegistryTransferIterator{contract: _Registry.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_Registry *RegistryFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *RegistryTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Registry.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RegistryTransfer)
				if err := _Registry.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_Registry *RegistryFilterer) ParseTransfer(log types.Log) (*RegistryTransfer, error) {
	event := new(RegistryTransfer)
	if err := _Registry.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
