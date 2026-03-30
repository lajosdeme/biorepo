package biorepo

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/lajosdeme/biorepo/index"
	"github.com/lajosdeme/biorepo/store"
	"github.com/lajosdeme/biorepo/types"
)

// -------------------------------------------------------------------------------
// Design — the top-level object callers work with.
// -------------------------------------------------------------------------------

// Design is the complete representation of a synthetic biology design.
// It joins the onchain BioCommit record with the actual Sequence bytes
// from Filecoin. Neither backend knows about the other — the Client is
// the only place they are composed.
type Design struct {
	CommitId    index.CommitId
	Commit      index.BioCommit
	Sequence    types.Sequence
	FilecoinCID store.CID
}

// NewDesign carries the caller-supplied fields needed to publish a design.
// Fields derived by the backends (CommitId, Timestamp, ContentHash) are absent.
type NewDesign struct {
	Sequence    types.Sequence
	Parent      index.CommitId // ZeroCommitId for root commits
	Author      common.Address
	ProblemTag  [32]byte
	FunctionTag [32]byte

	// Confidence is the Evo 2 fitness score in [0, 1]. Scaled internally
	// to the contract's 0–1_000_000 range before writing onchain.
	Confidence float64
}

// -------------------------------------------------------------------------------
// Client
// -------------------------------------------------------------------------------

// Client is the single entry point for all biorepository operations.
// It composes a Store (Filecoin, for sequence bytes) and an Index
// (onchain contract, for metadata and CID lookup).
//
// Construct with New. Both backends must be non-nil.
//
// Example — production:
//
//	s := filecoinstore.New(apiToken)
//	i := onchainindex.New(ethClient, contractAddr, txOpts)
//	client := biorepository.New(s, i)
//
// Example — tests:
//
//	client := biorepository.New(store.NewMemory(), index.NewMemory())
type Client struct {
	store store.Store
	index index.Index
}

// New constructs a Client from a Store and an Index implementation.
func New(s store.Store, i index.Index) *Client {
	if s == nil {
		panic("biorepository.New: store must not be nil")
	}
	if i == nil {
		panic("biorepository.New: index must not be nil")
	}
	return &Client{store: s, index: i}
}

// -------------------------------------------------------------------------------
// Write operations
// -------------------------------------------------------------------------------

// PublishDesign stores the sequence on Filecoin and commits its metadata
// onchain. The Filecoin CID is passed to the contract, which derives the
// contentHash as keccak256(abi.encode(cid)) and stores the CID in
// cidByCommit for later retrieval.
//
// PublishDesign validates the sequence before touching either backend.
// It is safe to retry: the Store's idempotent Put means a second upload
// of the same sequence is a no-op, and the contract's CommitExists guard
// means a duplicate commit returns ErrCommitExists rather than writing twice.
func (c *Client) PublishDesign(ctx context.Context, d NewDesign) (index.CommitId, error) {
	if err := d.Sequence.Validate(); err != nil {
		return index.CommitId{}, fmt.Errorf("publishDesign: invalid sequence: %w", err)
	}

	// 1. Encode as FASTA and upload to Filecoin.
	var buf bytes.Buffer
	if err := d.Sequence.EncodeFASTA(&buf); err != nil {
		return index.CommitId{}, fmt.Errorf("publishDesign: encoding FASTA: %w", err)
	}
	cid, err := c.store.Put(ctx, buf.Bytes())
	if err != nil {
		return index.CommitId{}, fmt.Errorf("publishDesign: storing sequence: %w", err)
	}

	// 2. Commit metadata onchain. The contract derives contentHash from the
	// CID string and stores the CID in cidByCommit — we pass cid directly.
	commitId, err := c.index.Commit(ctx, index.CommitRequest{
		CID:         cid,
		Parent:      d.Parent,
		Author:      d.Author,
		ProblemTag:  d.ProblemTag,
		FunctionTag: d.FunctionTag,
		Confidence:  scaleConfidence(d.Confidence),
	})
	if err != nil {
		return index.CommitId{}, fmt.Errorf("publishDesign: committing to index: %w", err)
	}

	return commitId, nil
}

// ForkDesign creates a new commit whose Parent is the given CommitId.
// The parent must already exist in the index. The forked design carries
// an entirely new sequence — use types.Sequence methods (Subsequence,
// ReverseComplement) to derive the new sequence before calling ForkDesign.
func (c *Client) ForkDesign(ctx context.Context, parentId index.CommitId, d NewDesign) (index.CommitId, error) {
	if _, err := c.index.GetCommit(ctx, parentId); err != nil {
		return index.CommitId{}, fmt.Errorf("forkDesign: parent not found: %w", err)
	}
	d.Parent = parentId
	return c.PublishDesign(ctx, d)
}

// -------------------------------------------------------------------------------
// Read operations
// -------------------------------------------------------------------------------

// GetDesign retrieves the full Design for a given CommitId.
//
// Flow:
//  1. Fetch BioCommit from the onchain index.
//  2. Fetch the Filecoin CID via cidByCommit (a contract view call).
//  3. Fetch sequence bytes from Filecoin using the CID.
//  4. Decode FASTA bytes into a types.Sequence.
func (c *Client) GetDesign(ctx context.Context, id index.CommitId) (Design, error) {
	// 1. Fetch onchain metadata.
	commit, err := c.index.GetCommit(ctx, id)
	if err != nil {
		return Design{}, fmt.Errorf("getDesign: fetching commit: %w", err)
	}

	// 2. Retrieve the Filecoin CID stored by the contract.
	cid, err := c.index.GetCID(ctx, id)
	if err != nil {
		return Design{}, fmt.Errorf("getDesign: fetching CID for commit %s: %w", id.Hex(), err)
	}

	// 3. Fetch sequence bytes from Filecoin.
	seqBytes, err := c.store.Get(ctx, cid)
	if err != nil {
		return Design{}, fmt.Errorf("getDesign: fetching sequence (cid=%s): %w", cid, err)
	}

	// 4. Decode FASTA.
	seq, err := types.DecodeFASTA(bytes.NewReader(seqBytes))
	if err != nil {
		return Design{}, fmt.Errorf("getDesign: decoding FASTA: %w", err)
	}

	return Design{
		CommitId:    id,
		Commit:      commit,
		Sequence:    seq,
		FilecoinCID: cid,
	}, nil
}

// GetLineage walks parent pointers from id back to the root, returning
// designs ordered root → id. A root design (no parent) returns a slice
// of length 1.
func (c *Client) GetLineage(ctx context.Context, id index.CommitId) ([]Design, error) {
	// Walk up the parent chain collecting CommitIds in reverse order.
	var chain []index.CommitId
	current := id
	for {
		chain = append(chain, current)
		commit, err := c.index.GetCommit(ctx, current)
		if err != nil {
			return nil, fmt.Errorf("getLineage: fetching commit %s: %w", current.Hex(), err)
		}

		if index.CommitId(commit.Parent).IsZero() {
			break
		}
		current = commit.Parent
	}

	// Reverse so the slice runs root → id.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return c.resolveIds(ctx, chain, "getLineage")
}

// GetChildren returns the immediate children of id as full Design values.
// Returns an empty slice (not an error) if id is a leaf commit.
func (c *Client) GetChildren(ctx context.Context, id index.CommitId) ([]Design, error) {
	childIds, err := c.index.GetChildren(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getChildren: %w", err)
	}
	return c.resolveIds(ctx, childIds, "getChildren")
}

// GetCommitsByAuthor returns all designs published by the given address,
// in the order they were committed.
func (c *Client) GetCommitsByAuthor(ctx context.Context, author common.Address) ([]Design, error) {
	ids, err := c.index.GetCommitsByAuthor(ctx, author)
	if err != nil {
		return nil, fmt.Errorf("getCommitsByAuthor: %w", err)
	}
	return c.resolveIds(ctx, ids, "getCommitsByAuthor")
}

// -------------------------------------------------------------------------------
// Search operations
// -------------------------------------------------------------------------------

// SearchByTag returns designs whose ProblemTag or FunctionTag matches tag,
// up to limit results (0 = no limit). Use types.TagFromString to produce
// the tag from a human-readable label.
//
// Example:
//
//	designs, err := client.SearchByTag(ctx, types.TagFromString("drought-resistance"), 10)
func (c *Client) SearchByTag(ctx context.Context, tag [32]byte, limit int, lookbackBlocks uint64) ([]Design, error) {
	ids, err := c.index.SearchByTag(ctx, tag, limit, lookbackBlocks)
	if err != nil {
		return nil, fmt.Errorf("searchByTag: %w", err)
	}
	return c.resolveIds(ctx, ids, "searchByTag")
}

// -------------------------------------------------------------------------------
// Internal helpers
// -------------------------------------------------------------------------------

// resolveIds fetches full Design values for a slice of CommitIds, preserving order.
func (c *Client) resolveIds(ctx context.Context, ids []index.CommitId, caller string) ([]Design, error) {
	designs := make([]Design, 0, len(ids))
	for _, id := range ids {
		d, err := c.GetDesign(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("%s: resolving design %s: %w", caller, id.Hex(), err)
		}
		designs = append(designs, d)
	}
	return designs, nil
}

// PutRaw uploads arbitrary bytes to Filecoin and returns the CID.
// Used for non-sequence artifacts such as the autoresearch report JSON
// and Markdown files. The bytes are not indexed onchain.
func (c *Client) PutRaw(ctx context.Context, data []byte) (store.CID, error) {
	cid, err := c.store.Put(ctx, data)
	if err != nil {
		return "", fmt.Errorf("putRaw: %w", err)
	}
	return cid, nil
}

// GetCID returns the Filecoin CID stored onchain for a CommitId.
// This is a thin wrapper over index.GetCID, exposed so the publisher
// can record the CID in the report without accessing the index directly.
func (c *Client) GetCID(ctx context.Context, id index.CommitId) (store.CID, error) {
	cid, err := c.index.GetCID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("getCID(%s): %w", id.Hex(), err)
	}
	return cid, nil
}

// scaleConfidence converts a float64 in [0, 1] to a uint32 in [0, 1_000_000],
// clamping out-of-range values rather than erroring.
func scaleConfidence(f float64) uint32 {
	if f <= 0 {
		return 0
	}
	if f >= 1 {
		return 1_000_000
	}
	return uint32(f * 1_000_000)
}
