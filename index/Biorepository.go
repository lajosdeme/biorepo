// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package index

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
	_ = abi.ConvertType
)

// BioCommit is an auto generated low-level Go binding around an user-defined struct.
type BioCommit struct {
	ContentHash [32]byte
	Parent      [32]byte
	Author      common.Address
	Timestamp   uint64
	ProblemTag  [32]byte
	FunctionTag [32]byte
	Confidence  uint32
}

// BiorepositoryMetaData contains all meta data concerning the Biorepository contract.
var BiorepositoryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"children\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"cidByCommit\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"commit\",\"inputs\":[{\"name\":\"parent\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"},{\"name\":\"problemTag\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"functionTag\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"confidence\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"cid\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"commitId\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"commits\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}],\"outputs\":[{\"name\":\"contentHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"parent\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"},{\"name\":\"author\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"problemTag\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"functionTag\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"confidence\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"commitsByAuthor\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"exists\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChildren\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"CommitId[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCommit\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBioCommit\",\"components\":[{\"name\":\"contentHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"parent\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"},{\"name\":\"author\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"problemTag\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"functionTag\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"confidence\",\"type\":\"uint32\",\"internalType\":\"uint32\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCommitsByAuthor\",\"inputs\":[{\"name\":\"author\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"CommitId[]\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"BioCommitCreated\",\"inputs\":[{\"name\":\"commitId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"CommitId\"},{\"name\":\"parent\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"CommitId\"},{\"name\":\"author\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"contentHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"problemTag\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"functionTag\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"confidence\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"CommitExists\",\"inputs\":[{\"name\":\"commitId\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}]},{\"type\":\"error\",\"name\":\"ParentDoesNotExist\",\"inputs\":[{\"name\":\"parent\",\"type\":\"bytes32\",\"internalType\":\"CommitId\"}]}]",
}

// BiorepositoryABI is the input ABI used to generate the binding from.
// Deprecated: Use BiorepositoryMetaData.ABI instead.
var BiorepositoryABI = BiorepositoryMetaData.ABI

// Biorepository is an auto generated Go binding around an Ethereum contract.
type Biorepository struct {
	BiorepositoryCaller     // Read-only binding to the contract
	BiorepositoryTransactor // Write-only binding to the contract
	BiorepositoryFilterer   // Log filterer for contract events
}

// BiorepositoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type BiorepositoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BiorepositoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BiorepositoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BiorepositoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BiorepositoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BiorepositorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BiorepositorySession struct {
	Contract     *Biorepository    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BiorepositoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BiorepositoryCallerSession struct {
	Contract *BiorepositoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// BiorepositoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BiorepositoryTransactorSession struct {
	Contract     *BiorepositoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// BiorepositoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type BiorepositoryRaw struct {
	Contract *Biorepository // Generic contract binding to access the raw methods on
}

// BiorepositoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BiorepositoryCallerRaw struct {
	Contract *BiorepositoryCaller // Generic read-only contract binding to access the raw methods on
}

// BiorepositoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BiorepositoryTransactorRaw struct {
	Contract *BiorepositoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBiorepository creates a new instance of Biorepository, bound to a specific deployed contract.
func NewBiorepository(address common.Address, backend bind.ContractBackend) (*Biorepository, error) {
	contract, err := bindBiorepository(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Biorepository{BiorepositoryCaller: BiorepositoryCaller{contract: contract}, BiorepositoryTransactor: BiorepositoryTransactor{contract: contract}, BiorepositoryFilterer: BiorepositoryFilterer{contract: contract}}, nil
}

// NewBiorepositoryCaller creates a new read-only instance of Biorepository, bound to a specific deployed contract.
func NewBiorepositoryCaller(address common.Address, caller bind.ContractCaller) (*BiorepositoryCaller, error) {
	contract, err := bindBiorepository(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BiorepositoryCaller{contract: contract}, nil
}

// NewBiorepositoryTransactor creates a new write-only instance of Biorepository, bound to a specific deployed contract.
func NewBiorepositoryTransactor(address common.Address, transactor bind.ContractTransactor) (*BiorepositoryTransactor, error) {
	contract, err := bindBiorepository(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BiorepositoryTransactor{contract: contract}, nil
}

// NewBiorepositoryFilterer creates a new log filterer instance of Biorepository, bound to a specific deployed contract.
func NewBiorepositoryFilterer(address common.Address, filterer bind.ContractFilterer) (*BiorepositoryFilterer, error) {
	contract, err := bindBiorepository(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BiorepositoryFilterer{contract: contract}, nil
}

// bindBiorepository binds a generic wrapper to an already deployed contract.
func bindBiorepository(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BiorepositoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Biorepository *BiorepositoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Biorepository.Contract.BiorepositoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Biorepository *BiorepositoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Biorepository.Contract.BiorepositoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Biorepository *BiorepositoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Biorepository.Contract.BiorepositoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Biorepository *BiorepositoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Biorepository.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Biorepository *BiorepositoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Biorepository.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Biorepository *BiorepositoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Biorepository.Contract.contract.Transact(opts, method, params...)
}

// Children is a free data retrieval call binding the contract method 0x9e96f3cd.
//
// Solidity: function children(bytes32 , uint256 ) view returns(bytes32)
func (_Biorepository *BiorepositoryCaller) Children(opts *bind.CallOpts, arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Biorepository.contract.Call(opts, &out, "children", arg0, arg1)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Children is a free data retrieval call binding the contract method 0x9e96f3cd.
//
// Solidity: function children(bytes32 , uint256 ) view returns(bytes32)
func (_Biorepository *BiorepositorySession) Children(arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	return _Biorepository.Contract.Children(&_Biorepository.CallOpts, arg0, arg1)
}

// Children is a free data retrieval call binding the contract method 0x9e96f3cd.
//
// Solidity: function children(bytes32 , uint256 ) view returns(bytes32)
func (_Biorepository *BiorepositoryCallerSession) Children(arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	return _Biorepository.Contract.Children(&_Biorepository.CallOpts, arg0, arg1)
}

// CidByCommit is a free data retrieval call binding the contract method 0xec165dd4.
//
// Solidity: function cidByCommit(bytes32 ) view returns(string)
func (_Biorepository *BiorepositoryCaller) CidByCommit(opts *bind.CallOpts, arg0 [32]byte) (string, error) {
	var out []interface{}
	err := _Biorepository.contract.Call(opts, &out, "cidByCommit", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// CidByCommit is a free data retrieval call binding the contract method 0xec165dd4.
//
// Solidity: function cidByCommit(bytes32 ) view returns(string)
func (_Biorepository *BiorepositorySession) CidByCommit(arg0 [32]byte) (string, error) {
	return _Biorepository.Contract.CidByCommit(&_Biorepository.CallOpts, arg0)
}

// CidByCommit is a free data retrieval call binding the contract method 0xec165dd4.
//
// Solidity: function cidByCommit(bytes32 ) view returns(string)
func (_Biorepository *BiorepositoryCallerSession) CidByCommit(arg0 [32]byte) (string, error) {
	return _Biorepository.Contract.CidByCommit(&_Biorepository.CallOpts, arg0)
}

// Commits is a free data retrieval call binding the contract method 0x47885781.
//
// Solidity: function commits(bytes32 ) view returns(bytes32 contentHash, bytes32 parent, address author, uint64 timestamp, bytes32 problemTag, bytes32 functionTag, uint32 confidence)
func (_Biorepository *BiorepositoryCaller) Commits(opts *bind.CallOpts, arg0 [32]byte) (struct {
	ContentHash [32]byte
	Parent      [32]byte
	Author      common.Address
	Timestamp   uint64
	ProblemTag  [32]byte
	FunctionTag [32]byte
	Confidence  uint32
}, error) {
	var out []interface{}
	err := _Biorepository.contract.Call(opts, &out, "commits", arg0)

	outstruct := new(struct {
		ContentHash [32]byte
		Parent      [32]byte
		Author      common.Address
		Timestamp   uint64
		ProblemTag  [32]byte
		FunctionTag [32]byte
		Confidence  uint32
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ContentHash = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Parent = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.Author = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.Timestamp = *abi.ConvertType(out[3], new(uint64)).(*uint64)
	outstruct.ProblemTag = *abi.ConvertType(out[4], new([32]byte)).(*[32]byte)
	outstruct.FunctionTag = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Confidence = *abi.ConvertType(out[6], new(uint32)).(*uint32)

	return *outstruct, err

}

// Commits is a free data retrieval call binding the contract method 0x47885781.
//
// Solidity: function commits(bytes32 ) view returns(bytes32 contentHash, bytes32 parent, address author, uint64 timestamp, bytes32 problemTag, bytes32 functionTag, uint32 confidence)
func (_Biorepository *BiorepositorySession) Commits(arg0 [32]byte) (struct {
	ContentHash [32]byte
	Parent      [32]byte
	Author      common.Address
	Timestamp   uint64
	ProblemTag  [32]byte
	FunctionTag [32]byte
	Confidence  uint32
}, error) {
	return _Biorepository.Contract.Commits(&_Biorepository.CallOpts, arg0)
}

// Commits is a free data retrieval call binding the contract method 0x47885781.
//
// Solidity: function commits(bytes32 ) view returns(bytes32 contentHash, bytes32 parent, address author, uint64 timestamp, bytes32 problemTag, bytes32 functionTag, uint32 confidence)
func (_Biorepository *BiorepositoryCallerSession) Commits(arg0 [32]byte) (struct {
	ContentHash [32]byte
	Parent      [32]byte
	Author      common.Address
	Timestamp   uint64
	ProblemTag  [32]byte
	FunctionTag [32]byte
	Confidence  uint32
}, error) {
	return _Biorepository.Contract.Commits(&_Biorepository.CallOpts, arg0)
}

// CommitsByAuthor is a free data retrieval call binding the contract method 0xc27f80cd.
//
// Solidity: function commitsByAuthor(address , uint256 ) view returns(bytes32)
func (_Biorepository *BiorepositoryCaller) CommitsByAuthor(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Biorepository.contract.Call(opts, &out, "commitsByAuthor", arg0, arg1)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CommitsByAuthor is a free data retrieval call binding the contract method 0xc27f80cd.
//
// Solidity: function commitsByAuthor(address , uint256 ) view returns(bytes32)
func (_Biorepository *BiorepositorySession) CommitsByAuthor(arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	return _Biorepository.Contract.CommitsByAuthor(&_Biorepository.CallOpts, arg0, arg1)
}

// CommitsByAuthor is a free data retrieval call binding the contract method 0xc27f80cd.
//
// Solidity: function commitsByAuthor(address , uint256 ) view returns(bytes32)
func (_Biorepository *BiorepositoryCallerSession) CommitsByAuthor(arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	return _Biorepository.Contract.CommitsByAuthor(&_Biorepository.CallOpts, arg0, arg1)
}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_Biorepository *BiorepositoryCaller) Exists(opts *bind.CallOpts, id [32]byte) (bool, error) {
	var out []interface{}
	err := _Biorepository.contract.Call(opts, &out, "exists", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_Biorepository *BiorepositorySession) Exists(id [32]byte) (bool, error) {
	return _Biorepository.Contract.Exists(&_Biorepository.CallOpts, id)
}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_Biorepository *BiorepositoryCallerSession) Exists(id [32]byte) (bool, error) {
	return _Biorepository.Contract.Exists(&_Biorepository.CallOpts, id)
}

// GetChildren is a free data retrieval call binding the contract method 0xd37684ff.
//
// Solidity: function getChildren(bytes32 id) view returns(bytes32[])
func (_Biorepository *BiorepositoryCaller) GetChildren(opts *bind.CallOpts, id [32]byte) ([][32]byte, error) {
	var out []interface{}
	err := _Biorepository.contract.Call(opts, &out, "getChildren", id)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetChildren is a free data retrieval call binding the contract method 0xd37684ff.
//
// Solidity: function getChildren(bytes32 id) view returns(bytes32[])
func (_Biorepository *BiorepositorySession) GetChildren(id [32]byte) ([][32]byte, error) {
	return _Biorepository.Contract.GetChildren(&_Biorepository.CallOpts, id)
}

// GetChildren is a free data retrieval call binding the contract method 0xd37684ff.
//
// Solidity: function getChildren(bytes32 id) view returns(bytes32[])
func (_Biorepository *BiorepositoryCallerSession) GetChildren(id [32]byte) ([][32]byte, error) {
	return _Biorepository.Contract.GetChildren(&_Biorepository.CallOpts, id)
}

// GetCommit is a free data retrieval call binding the contract method 0x8784ea96.
//
// Solidity: function getCommit(bytes32 id) view returns((bytes32,bytes32,address,uint64,bytes32,bytes32,uint32))
func (_Biorepository *BiorepositoryCaller) GetCommit(opts *bind.CallOpts, id [32]byte) (BioCommit, error) {
	var out []interface{}
	err := _Biorepository.contract.Call(opts, &out, "getCommit", id)

	if err != nil {
		return *new(BioCommit), err
	}

	out0 := *abi.ConvertType(out[0], new(BioCommit)).(*BioCommit)

	return out0, err

}

// GetCommit is a free data retrieval call binding the contract method 0x8784ea96.
//
// Solidity: function getCommit(bytes32 id) view returns((bytes32,bytes32,address,uint64,bytes32,bytes32,uint32))
func (_Biorepository *BiorepositorySession) GetCommit(id [32]byte) (BioCommit, error) {
	return _Biorepository.Contract.GetCommit(&_Biorepository.CallOpts, id)
}

// GetCommit is a free data retrieval call binding the contract method 0x8784ea96.
//
// Solidity: function getCommit(bytes32 id) view returns((bytes32,bytes32,address,uint64,bytes32,bytes32,uint32))
func (_Biorepository *BiorepositoryCallerSession) GetCommit(id [32]byte) (BioCommit, error) {
	return _Biorepository.Contract.GetCommit(&_Biorepository.CallOpts, id)
}

// GetCommitsByAuthor is a free data retrieval call binding the contract method 0x1eb44767.
//
// Solidity: function getCommitsByAuthor(address author) view returns(bytes32[])
func (_Biorepository *BiorepositoryCaller) GetCommitsByAuthor(opts *bind.CallOpts, author common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _Biorepository.contract.Call(opts, &out, "getCommitsByAuthor", author)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetCommitsByAuthor is a free data retrieval call binding the contract method 0x1eb44767.
//
// Solidity: function getCommitsByAuthor(address author) view returns(bytes32[])
func (_Biorepository *BiorepositorySession) GetCommitsByAuthor(author common.Address) ([][32]byte, error) {
	return _Biorepository.Contract.GetCommitsByAuthor(&_Biorepository.CallOpts, author)
}

// GetCommitsByAuthor is a free data retrieval call binding the contract method 0x1eb44767.
//
// Solidity: function getCommitsByAuthor(address author) view returns(bytes32[])
func (_Biorepository *BiorepositoryCallerSession) GetCommitsByAuthor(author common.Address) ([][32]byte, error) {
	return _Biorepository.Contract.GetCommitsByAuthor(&_Biorepository.CallOpts, author)
}

// Commit is a paid mutator transaction binding the contract method 0xb129ea0c.
//
// Solidity: function commit(bytes32 parent, bytes32 problemTag, bytes32 functionTag, uint32 confidence, string cid) returns(bytes32 commitId)
func (_Biorepository *BiorepositoryTransactor) Commit(opts *bind.TransactOpts, parent [32]byte, problemTag [32]byte, functionTag [32]byte, confidence uint32, cid string) (*types.Transaction, error) {
	return _Biorepository.contract.Transact(opts, "commit", parent, problemTag, functionTag, confidence, cid)
}

// Commit is a paid mutator transaction binding the contract method 0xb129ea0c.
//
// Solidity: function commit(bytes32 parent, bytes32 problemTag, bytes32 functionTag, uint32 confidence, string cid) returns(bytes32 commitId)
func (_Biorepository *BiorepositorySession) Commit(parent [32]byte, problemTag [32]byte, functionTag [32]byte, confidence uint32, cid string) (*types.Transaction, error) {
	return _Biorepository.Contract.Commit(&_Biorepository.TransactOpts, parent, problemTag, functionTag, confidence, cid)
}

// Commit is a paid mutator transaction binding the contract method 0xb129ea0c.
//
// Solidity: function commit(bytes32 parent, bytes32 problemTag, bytes32 functionTag, uint32 confidence, string cid) returns(bytes32 commitId)
func (_Biorepository *BiorepositoryTransactorSession) Commit(parent [32]byte, problemTag [32]byte, functionTag [32]byte, confidence uint32, cid string) (*types.Transaction, error) {
	return _Biorepository.Contract.Commit(&_Biorepository.TransactOpts, parent, problemTag, functionTag, confidence, cid)
}

// BiorepositoryBioCommitCreatedIterator is returned from FilterBioCommitCreated and is used to iterate over the raw logs and unpacked data for BioCommitCreated events raised by the Biorepository contract.
type BiorepositoryBioCommitCreatedIterator struct {
	Event *BiorepositoryBioCommitCreated // Event containing the contract specifics and raw log

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
func (it *BiorepositoryBioCommitCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BiorepositoryBioCommitCreated)
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
		it.Event = new(BiorepositoryBioCommitCreated)
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
func (it *BiorepositoryBioCommitCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BiorepositoryBioCommitCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BiorepositoryBioCommitCreated represents a BioCommitCreated event raised by the Biorepository contract.
type BiorepositoryBioCommitCreated struct {
	CommitId    [32]byte
	Parent      [32]byte
	Author      common.Address
	ContentHash [32]byte
	ProblemTag  [32]byte
	FunctionTag [32]byte
	Confidence  uint32
	Timestamp   *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterBioCommitCreated is a free log retrieval operation binding the contract event 0x20a7a356faa8535a788f1223856a99841acb384f4312bea6642a965717fe6e0b.
//
// Solidity: event BioCommitCreated(bytes32 indexed commitId, bytes32 parent, address author, bytes32 contentHash, bytes32 indexed problemTag, bytes32 indexed functionTag, uint32 confidence, uint256 timestamp)
func (_Biorepository *BiorepositoryFilterer) FilterBioCommitCreated(opts *bind.FilterOpts, commitId [][32]byte, problemTag [][32]byte, functionTag [][32]byte) (*BiorepositoryBioCommitCreatedIterator, error) {

	var commitIdRule []interface{}
	for _, commitIdItem := range commitId {
		commitIdRule = append(commitIdRule, commitIdItem)
	}

	var problemTagRule []interface{}
	for _, problemTagItem := range problemTag {
		problemTagRule = append(problemTagRule, problemTagItem)
	}
	var functionTagRule []interface{}
	for _, functionTagItem := range functionTag {
		functionTagRule = append(functionTagRule, functionTagItem)
	}

	logs, sub, err := _Biorepository.contract.FilterLogs(opts, "BioCommitCreated", commitIdRule, problemTagRule, functionTagRule)
	if err != nil {
		return nil, err
	}
	return &BiorepositoryBioCommitCreatedIterator{contract: _Biorepository.contract, event: "BioCommitCreated", logs: logs, sub: sub}, nil
}

// WatchBioCommitCreated is a free log subscription operation binding the contract event 0x20a7a356faa8535a788f1223856a99841acb384f4312bea6642a965717fe6e0b.
//
// Solidity: event BioCommitCreated(bytes32 indexed commitId, bytes32 parent, address author, bytes32 contentHash, bytes32 indexed problemTag, bytes32 indexed functionTag, uint32 confidence, uint256 timestamp)
func (_Biorepository *BiorepositoryFilterer) WatchBioCommitCreated(opts *bind.WatchOpts, sink chan<- *BiorepositoryBioCommitCreated, commitId [][32]byte, problemTag [][32]byte, functionTag [][32]byte) (event.Subscription, error) {

	var commitIdRule []interface{}
	for _, commitIdItem := range commitId {
		commitIdRule = append(commitIdRule, commitIdItem)
	}

	var problemTagRule []interface{}
	for _, problemTagItem := range problemTag {
		problemTagRule = append(problemTagRule, problemTagItem)
	}
	var functionTagRule []interface{}
	for _, functionTagItem := range functionTag {
		functionTagRule = append(functionTagRule, functionTagItem)
	}

	logs, sub, err := _Biorepository.contract.WatchLogs(opts, "BioCommitCreated", commitIdRule, problemTagRule, functionTagRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BiorepositoryBioCommitCreated)
				if err := _Biorepository.contract.UnpackLog(event, "BioCommitCreated", log); err != nil {
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

// ParseBioCommitCreated is a log parse operation binding the contract event 0x20a7a356faa8535a788f1223856a99841acb384f4312bea6642a965717fe6e0b.
//
// Solidity: event BioCommitCreated(bytes32 indexed commitId, bytes32 parent, address author, bytes32 contentHash, bytes32 indexed problemTag, bytes32 indexed functionTag, uint32 confidence, uint256 timestamp)
func (_Biorepository *BiorepositoryFilterer) ParseBioCommitCreated(log types.Log) (*BiorepositoryBioCommitCreated, error) {
	event := new(BiorepositoryBioCommitCreated)
	if err := _Biorepository.contract.UnpackLog(event, "BioCommitCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
