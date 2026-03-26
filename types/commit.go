package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// CommitId mirrors `type CommitId is bytes32` in Solidity
type CommitId [32]byte

var ZeroCommitId CommitId // all zeros — used as parent for root commits

func CommitIdFromHex(s string) (CommitId, error)
func (c CommitId) Hex() string
func (c CommitId) IsZero() bool

// BioCommit mirrors the Solidity struct exactly
// used when reading from chain via getCommit()
type BioCommit struct {
	ContentHash [32]byte // CID hash — keccak256 of the Filecoin CID bytes
	Parent      CommitId // parent commit (ZeroCommitId if root)
	Author      common.Address
	Timestamp   uint64
	ProblemTag  [32]byte // keccak256("drought-resistance") etc
	FunctionTag [32]byte // keccak256("stress-tolerance") etc
	Confidence  uint32   // 0–1_000_000, maps to 0.0–1.0
}

// types/tags.go

// TagFromString converts a human-readable tag to its bytes32 onchain representation
// TagFromString("drought-resistance") → keccak256("drought-resistance")
func TagFromString(s string) [32]byte {
	return crypto.Keccak256Hash([]byte(s))
}

// common problem tags — agents and UI use these constants
var (
	TagDroughtResistance   = TagFromString("drought-resistance")
	TagCarbonSequestration = TagFromString("carbon-sequestration")
	TagNitrogenFixation    = TagFromString("nitrogen-fixation")
	TagPlasticDegradation  = TagFromString("plastic-degradation")
	TagMethaneReduction    = TagFromString("methane-reduction")
)
