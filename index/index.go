package index

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lajosdeme/biorepo/store"
)

// -------------------------------------------------------------------------------
// Core types — Go mirrors of the Solidity BioCommit struct and CommitId type.
// These are defined here rather than imported from types/ because the index
// package owns the onchain representation. The biorepository.Client composes
// these with Sequence data from the store to produce the richer Design type.
// -------------------------------------------------------------------------------

// CommitId mirrors `type CommitId is bytes32` in Solidity.
// It is the primary key for every design in the repository.
type CommitId [32]byte

// ZeroCommitId is the zero value of CommitId — used as the parent field
// for root commits (i.e. designs with no ancestor).
var ZeroCommitId CommitId

// IsZero reports whether this CommitId is unset.
func (c CommitId) IsZero() bool {
	return c == ZeroCommitId
}

// Hex returns the 0x-prefixed hex representation, matching Solidity's
// bytes32 display convention.
func (c CommitId) Hex() string {
	return fmt.Sprintf("0x%x", c)
}

// CommitRequest carries the fields the caller provides when creating a new
// commit. The contract derives CommitId and Timestamp — we do not pass them.
type CommitRequest struct {
	CID         store.CID // Filecoin CID string — contract derives contentHash from this
	Parent      CommitId
	Author      common.Address
	ProblemTag  [32]byte
	FunctionTag [32]byte
	Confidence  uint32
}

// CommitEvent mirrors the BioCommitCreated Solidity event. The onchain
// implementation populates this from the event log; the memory implementation
// constructs it directly.
type CommitEvent struct {
	CommitId    CommitId
	Parent      CommitId
	Author      common.Address
	ContentHash [32]byte
	ProblemTag  [32]byte
	FunctionTag [32]byte
	Confidence  uint32
	Timestamp   uint64
}

// -------------------------------------------------------------------------------
// Index interface
// -------------------------------------------------------------------------------

// Index is the Go interface over IBioRepository. It covers the four contract
// functions plus event-based search, which is handled offchain against the
// indexed event log rather than via a contract call.
//
// All methods accept a context.Context so callers can propagate deadlines and
// cancellation across both the memory and onchain implementations.
//
// Implementations must be safe for concurrent use.
type Index interface {
	// Commit writes a new BioCommit to the repository and returns the
	// CommitId assigned by the contract. For the onchain implementation
	// this submits a transaction and waits for confirmation; for the memory
	// implementation it is synchronous.
	//
	// Returns ErrParentNotFound if req.Parent is non-zero and unknown.
	// Returns ErrCommitExists if the derived CommitId is already present.
	Commit(ctx context.Context, req CommitRequest) (CommitId, error)

	// GetCommit retrieves a single BioCommit by its CommitId.
	// Returns ErrNotFound if id is unknown.
	GetCommit(ctx context.Context, id CommitId) (BioCommit, error)

	// GetChildren returns the CommitIds of all direct children of id —
	// i.e. all commits whose Parent field equals id. Returns an empty
	// slice (not an error) if id has no children.
	GetChildren(ctx context.Context, id CommitId) ([]CommitId, error)

	// GetCommitsByAuthor returns all CommitIds created by the given address,
	// in the order they were committed. Returns an empty slice if the author
	// has no commits.
	GetCommitsByAuthor(ctx context.Context, author common.Address) ([]CommitId, error)

	// SearchByTag returns CommitIds whose ProblemTag or FunctionTag matches
	// tag, up to limit results. Order is implementation-defined.
	//
	// The onchain implementation filters eth_getLogs by the indexed topic;
	// the memory implementation scans its event slice. In both cases the
	// result is a list of CommitIds — callers resolve them to full BioCommit
	// values via GetCommit.
	//
	// A limit of 0 means no limit.
	SearchByTag(ctx context.Context, tag [32]byte, limit int) ([]CommitId, error)

	// GetCID returns the Filecoin CID stored onchain for the given CommitId.
	// This is a view call to cidByCommit() on the contract.
	GetCID(ctx context.Context, id CommitId) (store.CID, error)
}

// Sentinel errors.
var (
	// ErrNotFound is returned when a CommitId is not present in the index.
	ErrNotFound = fmt.Errorf("index: commit not found")

	// ErrParentNotFound mirrors the Solidity ParentDoesNotExist error.
	ErrParentNotFound = fmt.Errorf("index: parent commit does not exist")

	// ErrCommitExists mirrors the Solidity CommitExists error.
	ErrCommitExists = fmt.Errorf("index: commit already exists")
)

// -------------------------------------------------------------------------------
// Memory implementation
// -------------------------------------------------------------------------------

// Memory is an in-memory Index implementation for tests and local development.
// It mirrors the semantics of IBioRepository exactly, including the parent
// existence check and the CommitExists guard.
//
// CommitIds are computed as keccak256(contentHash ‖ parent ‖ author ‖ timestamp)
// to approximate what a realistic Solidity implementation would produce. The
// exact derivation does not need to match the real contract — the memory
// implementation is never used alongside a live chain.
type Memory struct {
	mu          sync.RWMutex
	commits     map[CommitId]BioCommit  // primary store
	children    map[CommitId][]CommitId // parent → children index
	byAuthor    map[common.Address][]CommitId
	cidByCommit map[CommitId]store.CID
	events      []CommitEvent // ordered log, used by SearchByTag
}

// NewMemory returns an initialised, empty in-memory index.
func NewMemory() *Memory {
	return &Memory{
		commits:     make(map[CommitId]BioCommit),
		children:    make(map[CommitId][]CommitId),
		byAuthor:    make(map[common.Address][]CommitId),
		cidByCommit: make(map[CommitId]store.CID),
	}
}

// Commit stores a new BioCommit and returns its CommitId.
func (m *Memory) Commit(_ context.Context, req CommitRequest) (CommitId, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// mirror the Solidity ParentDoesNotExist check
	if !req.Parent.IsZero() {
		if _, ok := m.commits[req.Parent]; !ok {
			return CommitId{}, fmt.Errorf("%w: %s", ErrParentNotFound, req.Parent.Hex())
		}
	}

	now := uint64(time.Now().Unix())
	id := deriveCommitId(req, now)

	// mirror the Solidity CommitExists check
	if _, ok := m.commits[id]; ok {
		return CommitId{}, fmt.Errorf("%w: %s", ErrCommitExists, id.Hex())
	}

	commit := BioCommit{
		ContentHash: crypto.Keccak256Hash([]byte(req.CID)),
		Parent:      req.Parent,
		Author:      req.Author,
		Timestamp:   now,
		ProblemTag:  req.ProblemTag,
		FunctionTag: req.FunctionTag,
		Confidence:  req.Confidence,
	}

	m.commits[id] = commit
	m.children[req.Parent] = append(m.children[req.Parent], id)
	m.byAuthor[req.Author] = append(m.byAuthor[req.Author], id)
	m.events = append(m.events, CommitEvent{
		CommitId:    id,
		Parent:      req.Parent,
		Author:      req.Author,
		ContentHash: crypto.Keccak256Hash([]byte(req.CID)),
		ProblemTag:  req.ProblemTag,
		FunctionTag: req.FunctionTag,
		Confidence:  req.Confidence,
		Timestamp:   now,
	})
	m.cidByCommit[id] = req.CID

	return id, nil
}

// GetCommit retrieves a BioCommit by id.
func (m *Memory) GetCommit(_ context.Context, id CommitId) (BioCommit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.commits[id]
	if !ok {
		return BioCommit{}, fmt.Errorf("%w: %s", ErrNotFound, id.Hex())
	}
	return c, nil
}

// GetChildren returns all direct children of id.
func (m *Memory) GetChildren(_ context.Context, id CommitId) ([]CommitId, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	children := m.children[id]
	if len(children) == 0 {
		return []CommitId{}, nil
	}
	// return a copy to prevent callers from mutating the internal slice
	out := make([]CommitId, len(children))
	copy(out, children)
	return out, nil
}

// GetCommitsByAuthor returns all CommitIds created by author, in commit order.
func (m *Memory) GetCommitsByAuthor(_ context.Context, author common.Address) ([]CommitId, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := m.byAuthor[author]
	if len(ids) == 0 {
		return []CommitId{}, nil
	}
	out := make([]CommitId, len(ids))
	copy(out, ids)
	return out, nil
}

// SearchByTag returns CommitIds whose ProblemTag or FunctionTag matches tag.
func (m *Memory) SearchByTag(_ context.Context, tag [32]byte, limit int) ([]CommitId, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []CommitId
	for _, ev := range m.events {
		if ev.ProblemTag == tag || ev.FunctionTag == tag {
			out = append(out, ev.CommitId)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	if out == nil {
		return []CommitId{}, nil
	}
	return out, nil
}

func (m *Memory) GetCID(_ context.Context, id CommitId) (store.CID, error) {
	cid := m.cidByCommit[id]

	if cid.IsZero() {
		return cid, errors.New("cid doesn't exist")
	}

	return cid, nil
}

// Len returns the number of commits in the index. Intended for tests.
func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.commits)
}

// -------------------------------------------------------------------------------
// CommitId derivation
// -------------------------------------------------------------------------------

// deriveCommitId computes a CommitId from a CommitRequest and a timestamp.
// This approximates what the real Solidity contract does internally.
//
// The input to the hash is:
//
//	contentHash (32 bytes)
//	‖ parent     (32 bytes)
//	‖ author     (20 bytes)
//	‖ timestamp  ( 8 bytes, big-endian uint64)
//
// This is intentionally simple — correctness of the memory implementation
// matters; matching the exact contract derivation does not, because the
// onchain implementation reads the CommitId from the transaction receipt
// rather than computing it locally.
func deriveCommitId(req CommitRequest, timestamp uint64) CommitId {
	var buf [92]byte // 32 + 32 + 20 + 8
	copy(buf[0:32], crypto.Keccak256([]byte(req.CID)))
	copy(buf[32:64], req.Parent[:])
	copy(buf[64:84], req.Author[:])
	binary.BigEndian.PutUint64(buf[84:92], timestamp)
	return CommitId(sha256.Sum256(buf[:]))
}
