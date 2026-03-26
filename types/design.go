package types

// Design is the full object a caller works with —
// it combines the onchain BioCommit metadata with the actual sequence
// data retrieved from Filecoin. This is what gets passed around in Go code;
// BioCommit is only used when reading/writing the chain directly.
type Design struct {
	CommitId    CommitId
	Commit      BioCommit // the onchain record
	Sequence    Sequence  // the actual DNA data from Filecoin
	FilecoinCID string    // the raw Filecoin CID string
}
