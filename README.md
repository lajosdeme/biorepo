# biorepo

The `biorepo` module is the foundational data layer of the godmode.bio protocol. It is the **GitHub for synthetic biology** — a content-addressed, onchain-indexed store for AI-generated and researcher-authored DNA designs.

This module is a standalone Go library. It is imported by the `autoresearch` agent, the godmode.bio web UI, and any other tooling that needs to read from or write to the protocol. It has no opinion about how designs are generated or evaluated — that belongs to the layers above it.

---

## Architecture

Every design in the protocol lives in two places simultaneously:

```
                ┌─────────────────────────────────────────┐
                │            biorepo.Client          │
                └────────────────┬────────────────────────┘
                                 │ composes
               ┌─────────────────┴──────────────────┐
               │                                    │
        store.Store                           index.Index
        (Filecoin)                         (BioRepository.sol)
               │                                    │
     sequence bytes                         onchain metadata
     addressed by CID                    CommitId → BioCommit
                                          CommitId → CID string
```

**Store** — a content-addressed blob store. Holds raw FASTA-encoded sequence bytes. The Filecoin implementation uses the [lighthouse.storage](https://lighthouse.storage) HTTP API; the memory implementation holds bytes in a `sync.RWMutex`-protected map.

**Index** — the onchain record. Each design produces one `BioCommit` on the `BioRepository` contract, identified by a `CommitId` (`bytes32`). The contract stores the Filecoin CID in a `cidByCommit` mapping so that any `CommitId` can be resolved back to its sequence without any offchain registry. The onchain implementation wraps the `abigen`-generated bindings; the memory implementation mirrors the contract's semantics exactly, including parent existence checks and the `CommitExists` guard.

**Client** — the public surface. It owns no state beyond its two backends. All persistence is delegated downward. The `Client` is the only place where a `BioCommit` and a `Sequence` are joined into a `Design`.

---

## Installation

```bash
go get github.com/lajosdeme/biorepo
```

Requires Go 1.21 or later.

---

## Quick start

### Production

```go
import (
    "github.com/lajosdeme/biorepo"
    filecoinstore "github.com/lajosdeme/biorepo/store/filecoin"
    onchainindex "github.com/lajosdeme/biorepo/index/onchain"
)

// Store — Filecoin via lighthouse.storage
s := filecoinstore.New(os.Getenv("WEB3_STORAGE_TOKEN"))

// Index — deployed BioRepository contract
ethClient, _ := ethclient.Dial(os.Getenv("ETH_RPC_URL"))
txOpts, _    := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
i, _         := onchainindex.NewOnchain(contractAddr, ethClient, txOpts)

// Client
client := biorepo.New(s, i)
```

### Tests and local development

```go
import (
    "github.com/lajosdeme/biorepo"
    "github.com/lajosdeme/biorepo/store"
    "github.com/lajosdeme/biorepo/index"
)

client := biorepo.New(store.NewMemory(), index.NewMemory())
```

Both memory implementations are safe for concurrent use and mirror the semantics of their production counterparts exactly. No network, no chain, no configuration required.

---

## Core concepts

### Design

A `Design` is the top-level object the `Client` returns. It combines the onchain `BioCommit` record with the actual `Sequence` bytes fetched from Filecoin.

```go
type Design struct {
    CommitId    index.CommitId  // the onchain primary key
    Commit      index.BioCommit // metadata stored on the contract
    Sequence    types.Sequence  // DNA sequence bytes + metadata
    FilecoinCID store.CID       // the Filecoin content address
}
```

### CommitId

A `CommitId` is a `bytes32` derived by the contract from the commit's content. It is the Git commit SHA equivalent — globally unique, deterministic, and permanent. Every operation that returns a design starts from a `CommitId`.

### BioCommit

The onchain record written by the contract for every design:

```go
type BioCommit struct {
    ContentHash [32]byte       // keccak256(abi.encode(cid)) — onchain integrity proof
    Parent      CommitId       // ZeroCommitId for root designs
    Author      common.Address // agent wallet or researcher wallet
    Timestamp   uint64         // unix seconds, set by the contract
    ProblemTag  [32]byte       // keccak256 of the problem label
    FunctionTag [32]byte       // keccak256 of the function label
    Confidence  uint32         // Evo 2 fitness score, scaled 0–1_000_000
}
```

### Tags

Tags are `bytes32` values derived from human-readable label strings via `keccak256`. Use `types.TagFromString` to convert, or the predefined constants for Season 1 themes:

```go
types.TagFromString("drought-resistance")  // → [32]byte

// predefined Season 1 constants
types.TagDroughtResistance
types.TagCarbonSequestration
types.TagNitrogenFixation
types.TagPlasticDegradation
types.TagMethaneReduction
```

### Confidence

Confidence is the Evo 2 fitness score for a design, expressed as a `float64` in `[0, 1]` on the `NewDesign` struct. The `Client` scales it to `uint32` in `[0, 1_000_000]` before writing onchain, matching the contract's representation. To convert back: `float64(commit.Confidence) / 1_000_000`.

---

## Usage

### Publishing a design

```go
seq := types.NewSequenceNormalized("ACGTACGT...", types.SequenceMeta{
    Description: "drought-tolerance locus variant",
    Organism:    "Zea mays",
    Strand:      types.StrandSense,
})

commitId, err := client.PublishDesign(ctx, biorepo.NewDesign{
    Sequence:    seq,
    Parent:      index.ZeroCommitId, // root — no ancestor
    Author:      agentWalletAddress,
    ProblemTag:  types.TagDroughtResistance,
    FunctionTag: types.TagFromString("osmotic-stress-response"),
    Confidence:  0.847,
})
```

`PublishDesign` validates the sequence, encodes it as FASTA, uploads to Filecoin, and commits metadata onchain. It is safe to retry — the store's idempotent `Put` and the contract's `CommitExists` guard make duplicate calls harmless.

### Forking a design

```go
// extend an existing design — parent must exist in the index
commitId, err := client.ForkDesign(ctx, parentCommitId, biorepo.NewDesign{
    Sequence:    derivedSeq,
    Author:      agentWalletAddress,
    ProblemTag:  types.TagDroughtResistance,
    FunctionTag: types.TagFromString("root-architecture"),
    Confidence:  0.901,
})
```

### Retrieving a design

```go
design, err := client.GetDesign(ctx, commitId)
fmt.Println(design.Sequence.Bases())       // raw DNA string
fmt.Println(design.Sequence.GCContent())   // float64 in [0, 1]
fmt.Println(design.FilecoinCID)            // "bafybeig..."
```

### Walking the lineage

```go
// returns all ancestors from root → commitId, inclusive
lineage, err := client.GetLineage(ctx, commitId)
for _, d := range lineage {
    fmt.Printf("%s  author=%s  confidence=%.3f\n",
        d.CommitId.Hex(),
        d.Commit.Author.Hex(),
        float64(d.Commit.Confidence)/1_000_000,
    )
}
```

### Searching by tag

```go
designs, err := client.SearchByTag(ctx, types.TagDroughtResistance, 10)
```

The onchain implementation issues two `eth_getLogs` calls (one per indexed topic field) and deduplicates results. For high-frequency search, maintain an event cache using the `WatchBioCommitCreated` filterer rather than calling `SearchByTag` in a loop.

### Getting all designs by an author

```go
// useful for the UI to show an agent's or researcher's commit history
designs, err := client.GetCommitsByAuthor(ctx, authorAddress)
```

---

## Sequence type

`types.Sequence` is an immutable DNA sequence with lazy validation. Construct with `NewSequence` (stores as-is) or `NewSequenceNormalized` (uppercases and strips whitespace on construction):

```go
// lazy — no validation at construction time
seq := types.NewSequence("ACGTacgt", meta)

// normalized — uppercase + whitespace stripped
seq := types.NewSequenceNormalized("ACGT ACGT\nACGT", meta)

// validate explicitly before publishing
if err := seq.Validate(); err != nil {
    // *types.SequenceValidationError listing all invalid positions
}

// transformations — return new Sequence values, receiver is unchanged
rc       := seq.ReverseComplement()
sub, err := seq.Subsequence(0, 100)

// properties
seq.Len()        // int
seq.GCContent()  // float64 in [0, 1]

// FASTA
fasta    := seq.FASTA()                    // string
err      := seq.EncodeFASTA(w)             // io.Writer
seq, err := types.DecodeFASTA(r)           // io.Reader, single record
seqs, err := types.DecodeFASTAMulti(r)    // io.Reader, multiple records
```

Valid bases follow the full [IUPAC DNA alphabet](https://www.bioinformatics.org/sms/iupac.html): `A C G T R Y S W K M B D H V N -`.

---

## Package structure

```
biorepo/
├── biorepo.go     # Client, Design, NewDesign — the public surface
├── types/
│   ├── sequence.go      # Sequence type, FASTA encode/decode, validation
│   ├── commit.go        # CommitId, BioCommit — Go mirrors of Solidity types
│   ├── design.go        # Design struct
│   └── tags.go          # TagFromString(), Season 1 tag constants
├── store/
│   ├── store.go         # Store interface, CID type, memory implementation
│   └── filecoin.go      # Filecoin/lighthouse.storage implementation
└── index/
    ├── index.go         # Index interface, CommitRequest, memory implementation
    ├── onchain.go       # Ethereum implementation using abigen bindings
    └── Biorepository.go # abigen-generated contract bindings (do not edit)
```

---

## Error handling

All errors are wrapped with context using `fmt.Errorf("operation: %w", err)`. Sentinel errors for type-checking:

```go
store.ErrNotFound      // CID not present in the store
store.ErrNotSupported  // operation not supported by this backend (e.g. Delete on Filecoin)
store.ErrCIDMismatch   // backend returned a CID that doesn't match expected content
index.ErrNotFound      // CommitId not present in the index
index.ErrParentNotFound // parent CommitId does not exist (mirrors Solidity ParentDoesNotExist)
index.ErrCommitExists  // CommitId already exists (mirrors Solidity CommitExists)
```

Check with `errors.Is`:

```go
design, err := client.GetDesign(ctx, id)
if errors.Is(err, index.ErrNotFound) {
    // handle missing commit
}
```

---

## Contributing

This module is the foundation of the godmode.bio protocol. Changes to the public `Client` surface, the `types` package, or the `Index` interface affect both the autoresearch agent and the UI — coordinate before making breaking changes.

The `store.Store` and `index.Index` interfaces are the extension points. To add a new backend (e.g. a local IPFS node, a different L2 chain), implement the relevant interface and inject it into `New`.