package index

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lajosdeme/biorepo/store"
)

// Onchain implements Index against a deployed BioRepository smart contract
// using the abigen-generated bindings in Biorepository.go.
//
// Construct with NewOnchain. The ethclient.Client must already be connected
// and the contract must be deployed at contractAddr.
type Onchain struct {
	contract *Biorepository     // abigen-generated binding
	client   *ethclient.Client  // used for receipt polling and block queries
	txOpts   *bind.TransactOpts // signer, gas settings — used for Commit()
}

// NewOnchain constructs an Onchain index.
//
// contractAddr is the deployed BioRepository address.
// ethClient must be connected to the correct network.
// txOpts must carry a valid signer for the agent's wallet address; it is
// used for every Commit() call. Use bind.NewKeyedTransactorWithChainID to
// construct it from a private key.
func NewOnchain(
	contractAddr common.Address,
	ethClient *ethclient.Client,
	txOpts *bind.TransactOpts,
) (*Onchain, error) {
	contract, err := NewBiorepository(contractAddr, ethClient)
	if err != nil {
		return nil, fmt.Errorf("onchain: binding contract at %s: %w", contractAddr.Hex(), err)
	}
	return &Onchain{
		contract: contract,
		client:   ethClient,
		txOpts:   txOpts,
	}, nil
}

// -------------------------------------------------------------------------------
// Index interface implementation
// -------------------------------------------------------------------------------

// Commit submits a commit transaction, waits for confirmation, and extracts
// the CommitId from the BioCommitCreated event in the receipt log.
func (o *Onchain) Commit(ctx context.Context, req CommitRequest) (CommitId, error) {
	// Copy txOpts so we don't mutate the shared instance across concurrent calls.
	opts := copyTransactOpts(o.txOpts)
	opts.Context = ctx

	tx, err := o.contract.Commit(
		opts,
		req.Parent,       // bytes32 parent
		req.ProblemTag,   // bytes32 problemTag
		req.FunctionTag,  // bytes32 functionTag
		req.Confidence,   // uint32 confidence
		req.CID.String(), // string cid
	)
	if err != nil {
		return CommitId{}, fmt.Errorf("onchain.Commit: submitting transaction: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, o.client, tx)
	if err != nil {
		return CommitId{}, fmt.Errorf("onchain.Commit: waiting for receipt (tx=%s): %w", tx.Hash().Hex(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return CommitId{}, fmt.Errorf("onchain.Commit: transaction reverted (tx=%s)", tx.Hash().Hex())
	}

	return o.extractCommitId(receipt)
}

// GetCommit calls getCommit() as a view call — no transaction, no gas.
func (o *Onchain) GetCommit(ctx context.Context, id CommitId) (BioCommit, error) {
	raw, err := o.contract.GetCommit(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return BioCommit{}, fmt.Errorf("onchain.GetCommit(%s): %w", CommitId(id).Hex(), err)
	}

	return BioCommit{
		ContentHash: raw.ContentHash,
		Parent:      CommitId(raw.Parent),
		Author:      raw.Author,
		Timestamp:   raw.Timestamp,
		ProblemTag:  raw.ProblemTag,
		FunctionTag: raw.FunctionTag,
		Confidence:  raw.Confidence,
	}, nil
}

// GetCID calls cidByCommit() as a view call to retrieve the Filecoin CID
// stored by the contract at commit time.
func (o *Onchain) GetCID(ctx context.Context, id CommitId) (store.CID, error) {
	cid, err := o.contract.CidByCommit(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return "", fmt.Errorf("onchain.GetCID(%s): %w", CommitId(id).Hex(), err)
	}
	if cid == "" {
		return "", fmt.Errorf("onchain.GetCID(%s): %w", CommitId(id).Hex(), ErrNotFound)
	}
	return store.CID(cid), nil
}

// GetChildren calls getChildren() as a view call.
func (o *Onchain) GetChildren(ctx context.Context, id CommitId) ([]CommitId, error) {
	raw, err := o.contract.GetChildren(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return nil, fmt.Errorf("onchain.GetChildren(%s): %w", CommitId(id).Hex(), err)
	}
	return toCommitIds(raw), nil
}

// GetCommitsByAuthor calls getCommitsByAuthor() as a view call.
func (o *Onchain) GetCommitsByAuthor(ctx context.Context, author common.Address) ([]CommitId, error) {
	raw, err := o.contract.GetCommitsByAuthor(&bind.CallOpts{Context: ctx}, author)
	if err != nil {
		return nil, fmt.Errorf("onchain.GetCommitsByAuthor(%s): %w", author.Hex(), err)
	}
	return toCommitIds(raw), nil
}

// SearchByTag returns CommitIds whose ProblemTag OR FunctionTag matches tag.
//
// Ethereum topic filtering is AND between fields, so matching either tag
// requires two separate filter calls — one per indexed field — merged and
// deduplicated. Both calls scan from the earliest block; callers doing
// high-frequency searches should maintain their own event cache rather than
// calling this in a hot path.
func (o *Onchain) SearchByTag(ctx context.Context, tag [32]byte, limit int, lookbackBlocks uint64) ([]CommitId, error) {
	seen := make(map[CommitId]struct{})
	var result []CommitId

	collect := func(iter *BiorepositoryBioCommitCreatedIterator) error {
		defer iter.Close()
		for iter.Next() {
			if limit > 0 && len(result) >= limit {
				return nil
			}
			id := CommitId(iter.Event.CommitId)
			if _, dup := seen[id]; !dup {
				seen[id] = struct{}{}
				result = append(result, id)
			}
		}
		return iter.Error()
	}

	header, err := o.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return result, err
	}

	latestBlock := header.Number.Uint64()
	startBlock := latestBlock - lookbackBlocks

	// Filter by problemTag.
	byProblem, err := o.contract.FilterBioCommitCreated(&bind.FilterOpts{
		Start: startBlock,
		End:   &latestBlock,
	}, nil, [][32]byte{tag}, nil)
	if err != nil {
		return nil, fmt.Errorf("onchain.SearchByTag: filtering by problemTag: %w", err)
	}
	if err := collect(byProblem); err != nil {
		return nil, fmt.Errorf("onchain.SearchByTag: iterating problemTag results: %w", err)
	}

	// Filter by functionTag — skip if limit already reached.
	if limit == 0 || len(result) < limit {
		byFunction, err := o.contract.FilterBioCommitCreated(
			&bind.FilterOpts{
				Start: startBlock,
				End:   &latestBlock,
			}, nil, nil, [][32]byte{tag})
		if err != nil {
			return nil, fmt.Errorf("onchain.SearchByTag: filtering by functionTag: %w", err)
		}
		if err := collect(byFunction); err != nil {
			return nil, fmt.Errorf("onchain.SearchByTag: iterating functionTag results: %w", err)
		}
	}

	return result, nil
}

// -------------------------------------------------------------------------------
// Internal helpers
// -------------------------------------------------------------------------------

// extractCommitId parses the BioCommitCreated event from a transaction receipt
// and returns the CommitId emitted by the contract.
func (o *Onchain) extractCommitId(receipt *types.Receipt) (CommitId, error) {
	// FilterBioCommitCreated requires a block range — use the receipt block.
	block := receipt.BlockNumber.Uint64()
	opts := &bind.FilterOpts{
		Start: block,
		End:   &block,
	}

	iter, err := o.contract.FilterBioCommitCreated(opts, nil, nil, nil)
	if err != nil {
		return CommitId{}, fmt.Errorf("extractCommitId: opening event filter: %w", err)
	}
	defer iter.Close()

	// Match the event whose transaction hash equals our tx — there may be
	// multiple commits in the same block from other callers.
	txHash := receipt.TxHash
	for iter.Next() {
		if iter.Event.Raw.TxHash == txHash {
			return CommitId(iter.Event.CommitId), nil
		}
	}
	if err := iter.Error(); err != nil {
		return CommitId{}, fmt.Errorf("extractCommitId: iterating events: %w", err)
	}

	return CommitId{}, fmt.Errorf("extractCommitId: BioCommitCreated event not found in tx %s", txHash.Hex())
}

// toCommitIds converts a [][ 32]byte from the ABI binding to []CommitId.
func toCommitIds(raw [][32]byte) []CommitId {
	out := make([]CommitId, len(raw))
	for i, r := range raw {
		out[i] = CommitId(r)
	}
	return out
}

// copyTransactOpts returns a shallow copy of opts so we can set Context
// per-call without mutating the shared instance.
func copyTransactOpts(opts *bind.TransactOpts) *bind.TransactOpts {
	copy := *opts
	if opts.GasPrice != nil {
		copy.GasPrice = new(big.Int).Set(opts.GasPrice)
	}
	return &copy
}
