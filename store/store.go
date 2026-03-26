package store

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"io"
	"strings"
)

// CID is a content identifier — a deterministic, content-addressed key
// derived from the bytes being stored. Modelled after IPFS CIDv1 but
// intentionally kept as an opaque string type so the Filecoin implementation
// can return the real IPFS CID and the memory implementation can return a
// locally-computed equivalent without either leaking into the interface.
//
// CIDs are always non-empty for successfully stored content. The zero value
// ("") is invalid and indicates an unset CID.
type CID string

// IsZero reports whether the CID is unset.
func (c CID) IsZero() bool {
	return c == ""
}

func (c CID) String() string {
	return string(c)
}

// ComputeCID derives the canonical CID for a byte slice without storing it.
// The memory store uses this directly; the Filecoin store uses it to verify
// the CID returned by the API matches the content that was uploaded.
//
// Algorithm: SHA-256 of the raw bytes, base32-encoded (lower, no padding),
// prefixed with "bafk" to visually resemble a real IPFS CIDv1 base32 CID.
// This is not a spec-compliant CIDv1 — it is intentionally a simplified
// stand-in that is stable, deterministic, and easy to compute in tests.
func ComputeCID(data []byte) CID {
	sum := sha256.Sum256(data)
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	encoded := strings.ToLower(enc.EncodeToString(sum[:]))
	return CID("bafk" + encoded)
}

// Store is the persistence interface for raw sequence bytes.
//
// The Store is intentionally narrow: it knows nothing about DNA, designs,
// or the onchain index. It is a content-addressed blob store. All semantic
// meaning is handled by the layers above it.
//
// Implementations must be safe for concurrent use.
type Store interface {
	// Put stores data and returns its CID. Put is idempotent — uploading
	// the same bytes twice returns the same CID and does not create a
	// duplicate entry. Callers may call Put without checking whether the
	// content already exists.
	Put(ctx context.Context, data []byte) (CID, error)

	// Get retrieves the raw bytes for a given CID. Returns ErrNotFound if
	// the CID is not present in this store.
	Get(ctx context.Context, cid CID) ([]byte, error)

	// Exists reports whether the given CID is present in the store without
	// fetching the associated data. Useful for deduplication checks before
	// a potentially expensive Put.
	Exists(ctx context.Context, cid CID) (bool, error)

	// List returns all CIDs currently held in the store. The order of
	// results is implementation-defined and must not be relied upon.
	// Intended for admin and sync tooling only — not for hot paths.
	List(ctx context.Context) ([]CID, error)

	// Delete removes the content for the given CID. Returns ErrNotFound if
	// the CID is not present. Implementations backed by immutable storage
	// (e.g. Filecoin) may return ErrNotSupported.
	Delete(ctx context.Context, cid CID) error
}

// WriterTo extends Store for callers that want to stream content out without
// loading it fully into memory. Not all implementations support this — check
// for the interface before using.
type WriterTo interface {
	// GetWriter writes the content for cid to w. Returns ErrNotFound if
	// the CID is not present.
	GetWriter(ctx context.Context, cid CID, w io.Writer) error
}

// Sentinel errors returned by Store implementations.
var (
	// ErrNotFound is returned by Get, Exists, and Delete when the given
	// CID is not present in the store.
	ErrNotFound = fmt.Errorf("store: CID not found")

	// ErrNotSupported is returned by operations the implementation does
	// not support (e.g. Delete on an immutable backend).
	ErrNotSupported = fmt.Errorf("store: operation not supported")

	// ErrCIDMismatch is returned when a remote store returns a CID that
	// does not match the locally-computed expected CID. This indicates
	// either data corruption in transit or a backend behaving unexpectedly.
	ErrCIDMismatch = fmt.Errorf("store: CID returned by backend does not match expected")
)
