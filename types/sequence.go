package types

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

// Strand represents the directionality of a DNA sequence.
type Strand int8

const (
	// StrandSense is the 5'→3' coding strand. This is the default and the
	// format Evo 2 expects for all generation and scoring calls.
	StrandSense Strand = iota

	// StrandAntisense is the 3'→5' template strand, stored as the reverse
	// complement of the sense sequence.
	StrandAntisense
)

func (s Strand) String() string {
	switch s {
	case StrandSense:
		return "sense (5'→3')"
	case StrandAntisense:
		return "antisense (3'→5')"
	default:
		return "unknown"
	}
}

// SequenceMeta holds the descriptive fields that travel with a sequence.
// These are sourced from the FASTA header line on parse, and written back
// to the header on encode. All fields are optional.
type SequenceMeta struct {
	// Description is the free-text label from the FASTA header (the part
	// after the '>identifier ' prefix). Example: "drought-tolerance locus, Chr3"
	Description string

	// Organism is the source species in binomial notation, if known.
	// Example: "Zea mays"
	Organism string

	// Strand is the directionality of the bases in Bases.
	// Defaults to StrandSense.
	Strand Strand
}

// Sequence is the core data type for a DNA sequence in the biorepository.
//
// Construction is lazy — no validation is performed on the bases at creation
// time. Call Validate() explicitly when correctness must be guaranteed (e.g.
// before publishing to Filecoin or passing to Evo 2).
//
// Sequences are immutable after construction. Methods that transform the
// sequence (ReverseComplement, Subsequence) return new Sequence values.
type Sequence struct {
	bases []byte
	meta  SequenceMeta
}

// NewSequence constructs a Sequence from a raw bases string and metadata.
// The bases are stored as-is (no normalisation, no validation). Use
// NewSequenceNormalized if you want uppercase + whitespace stripping applied
// at construction time.
func NewSequence(bases string, meta SequenceMeta) Sequence {
	return Sequence{
		bases: []byte(bases),
		meta:  meta,
	}
}

// NewSequenceNormalized constructs a Sequence after uppercasing all bases
// and stripping whitespace. Useful when ingesting sequences from external
// sources that may have mixed case or line breaks.
func NewSequenceNormalized(bases string, meta SequenceMeta) Sequence {
	var b strings.Builder
	b.Grow(len(bases))
	for _, r := range bases {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r - 32)
		} else {
			b.WriteRune(r)
		}
	}
	return Sequence{
		bases: []byte(b.String()),
		meta:  meta,
	}
}

// Bases returns the raw base string. This is what gets passed to Evo 2.
func (s Sequence) Bases() string {
	return string(s.bases)
}

// Meta returns the sequence metadata.
func (s Sequence) Meta() SequenceMeta {
	return s.meta
}

// Len returns the number of bases in the sequence.
func (s Sequence) Len() int {
	return len(s.bases)
}

// IsEmpty reports whether the sequence has no bases.
func (s Sequence) IsEmpty() bool {
	return len(s.bases) == 0
}

// Hash returns the SHA-256 digest of the raw bases. This is used to
// derive the contentHash field written to the BioRepository contract —
// it must be computed identically here and in the Solidity verifier.
//
// Note: the hash is over the normalised uppercase bases only, not the
// metadata, so that two representations of the same biological sequence
// hash identically regardless of description or organism fields.
func (s Sequence) Hash() [32]byte {
	upper := bytes.ToUpper(s.bases)
	return sha256.Sum256(upper)
}

// Validate checks that every base is a valid IUPAC DNA character.
// Returns a ValidationError listing all offending positions if any are found.
//
// Valid IUPAC DNA characters (case-insensitive):
//
//	A C G T           — unambiguous bases
//	R Y S W K M       — two-base ambiguity codes
//	B D H V           — three-base ambiguity codes
//	N                 — any base
//	-                 — gap
func (s Sequence) Validate() error {
	var errs []ValidationError
	for i, b := range s.bases {
		if !isValidIUPACDNA(b) {
			errs = append(errs, ValidationError{
				Position: i,
				Got:      b,
			})
			if len(errs) >= 20 {
				// cap error list to avoid flooding the caller
				break
			}
		}
	}
	if len(errs) > 0 {
		return &SequenceValidationError{Errors: errs}
	}
	return nil
}

// ReverseComplement returns a new Sequence whose bases are the reverse
// complement of this sequence, with the strand direction flipped.
//
// This is a pure transformation — the receiver is not modified.
// IUPAC ambiguity codes are complemented correctly (e.g. R↔Y, K↔M).
func (s Sequence) ReverseComplement() Sequence {
	n := len(s.bases)
	rc := make([]byte, n)
	for i, b := range s.bases {
		rc[n-1-i] = complement(b)
	}
	newStrand := StrandSense
	if s.meta.Strand == StrandSense {
		newStrand = StrandAntisense
	}
	return Sequence{
		bases: rc,
		meta: SequenceMeta{
			Description: s.meta.Description + " (reverse complement)",
			Organism:    s.meta.Organism,
			Strand:      newStrand,
		},
	}
}

// Subsequence returns a new Sequence containing the bases in the half-open
// interval [start, end). The metadata description is annotated with the
// coordinates. Returns an error if the interval is out of bounds.
func (s Sequence) Subsequence(start, end int) (Sequence, error) {
	if start < 0 || end > len(s.bases) || start >= end {
		return Sequence{}, fmt.Errorf(
			"subsequence [%d, %d) out of bounds for sequence of length %d",
			start, end, len(s.bases),
		)
	}
	sub := make([]byte, end-start)
	copy(sub, s.bases[start:end])
	return Sequence{
		bases: sub,
		meta: SequenceMeta{
			Description: fmt.Sprintf("%s [%d:%d]", s.meta.Description, start, end),
			Organism:    s.meta.Organism,
			Strand:      s.meta.Strand,
		},
	}, nil
}

// GCContent returns the fraction of bases that are G or C, in [0, 1].
// Returns 0 for an empty sequence. Ambiguous bases are not counted.
func (s Sequence) GCContent() float64 {
	if len(s.bases) == 0 {
		return 0
	}
	var gc int
	for _, b := range s.bases {
		upper := b | 0x20 // tolower
		if upper == 'g' || upper == 'c' {
			gc++
		}
	}
	return float64(gc) / float64(len(s.bases))
}

// FASTA encoding and decoding.
//
// The FASTA format used here:
//
//	>description | organism=Zea mays | strand=sense
//	ACGTACGTACGT...  (wrapped at 70 characters per line)

const fastaLineWidth = 70

// EncodeFASTA writes the sequence in FASTA format to w.
// The header encodes description, organism, and strand from the metadata.
func (s Sequence) EncodeFASTA(w io.Writer) error {
	header := buildFASTAHeader(s.meta)
	if _, err := fmt.Fprintf(w, ">%s\n", header); err != nil {
		return fmt.Errorf("writing FASTA header: %w", err)
	}
	bases := s.bases
	for len(bases) > 0 {
		end := fastaLineWidth
		if end > len(bases) {
			end = len(bases)
		}
		if _, err := w.Write(bases[:end]); err != nil {
			return fmt.Errorf("writing FASTA bases: %w", err)
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			return fmt.Errorf("writing FASTA newline: %w", err)
		}
		bases = bases[end:]
	}
	return nil
}

// FASTA returns the sequence encoded as a FASTA string.
// Convenience wrapper around EncodeFASTA.
func (s Sequence) FASTA() string {
	var buf bytes.Buffer
	_ = s.EncodeFASTA(&buf)
	return buf.String()
}

// DecodeFASTA parses a single FASTA record from r.
// Multi-record FASTA files should be split before calling this — use
// DecodeFASTAMulti to parse all records from a multi-record file.
//
// The header line is expected in the format written by EncodeFASTA, but
// we degrade gracefully: if only a plain description is present (no pipe
// delimiters), we use the full header text as the description.
func DecodeFASTA(r io.Reader) (Sequence, error) {
	seqs, err := DecodeFASTAMulti(r)
	if err != nil {
		return Sequence{}, err
	}
	if len(seqs) == 0 {
		return Sequence{}, fmt.Errorf("no FASTA records found")
	}
	return seqs[0], nil
}

// DecodeFASTAMulti parses all FASTA records from r. Returns at least one
// record or an error.
func DecodeFASTAMulti(r io.Reader) ([]Sequence, error) {
	var seqs []Sequence
	var currentMeta SequenceMeta
	var currentBases bytes.Buffer
	inRecord := false

	scanner := bufio.NewScanner(r)
	// allow up to 10 MB per line to handle very long sequences on one line
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	flush := func() {
		if inRecord {
			seqs = append(seqs, Sequence{
				bases: currentBases.Bytes(),
				meta:  currentMeta,
			})
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ">") {
			flush()
			currentMeta = parseFASTAHeader(strings.TrimPrefix(line, ">"))
			currentBases.Reset()
			inRecord = true
			continue
		}
		if inRecord {
			currentBases.WriteString(strings.TrimSpace(line))
		}
	}
	flush()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning FASTA input: %w", err)
	}
	if len(seqs) == 0 {
		return nil, fmt.Errorf("no FASTA records found")
	}
	return seqs, nil
}

// buildFASTAHeader serialises SequenceMeta into a single header line.
// Format: "description | organism=Zea mays | strand=sense"
// Fields with zero values are omitted.
func buildFASTAHeader(m SequenceMeta) string {
	parts := []string{}
	if m.Description != "" {
		parts = append(parts, m.Description)
	}
	if m.Organism != "" {
		parts = append(parts, "organism="+m.Organism)
	}
	parts = append(parts, "strand="+m.Strand.String())
	return strings.Join(parts, " | ")
}

// parseFASTAHeader parses a header line produced by buildFASTAHeader,
// degrading gracefully for plain descriptions without pipe delimiters.
func parseFASTAHeader(header string) SequenceMeta {
	parts := strings.Split(header, " | ")
	meta := SequenceMeta{Strand: StrandSense}
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "organism=") {
			meta.Organism = strings.TrimPrefix(part, "organism=")
		} else if strings.HasPrefix(part, "strand=") {
			v := strings.TrimPrefix(part, "strand=")
			if strings.Contains(v, "antisense") {
				meta.Strand = StrandAntisense
			}
		} else if i == 0 {
			meta.Description = part
		}
	}
	return meta
}

// Validation internals.

// isValidIUPACDNA reports whether b is a valid IUPAC DNA base (case-insensitive).
func isValidIUPACDNA(b byte) bool {
	switch b | 0x20 { // tolower
	case 'a', 'c', 'g', 't', // unambiguous
		'r', 'y', 's', 'w', 'k', 'm', // two-base ambiguity
		'b', 'd', 'h', 'v', // three-base ambiguity
		'n', // any base
		'-': // gap
		return true
	}
	return false
}

// complement returns the IUPAC complement of a DNA base.
// Handles both upper and lowercase; preserves case of input.
func complement(b byte) byte {
	upper := b &^ 0x20 // toupper
	var c byte
	switch upper {
	case 'A':
		c = 'T'
	case 'T':
		c = 'A'
	case 'C':
		c = 'G'
	case 'G':
		c = 'C'
	case 'R':
		c = 'Y'
	case 'Y':
		c = 'R'
	case 'S':
		c = 'S'
	case 'W':
		c = 'W'
	case 'K':
		c = 'M'
	case 'M':
		c = 'K'
	case 'B':
		c = 'V'
	case 'V':
		c = 'B'
	case 'D':
		c = 'H'
	case 'H':
		c = 'D'
	case 'N':
		c = 'N'
	case '-':
		c = '-'
	default:
		c = b // unknown: pass through
	}
	// preserve original case
	if b >= 'a' && b <= 'z' {
		return c | 0x20
	}
	return c
}

// Error types.

// ValidationError records a single invalid base at a specific position.
type ValidationError struct {
	Position int
	Got      byte
}

// SequenceValidationError is returned by Validate() when invalid bases
// are found. It implements the error interface.
type SequenceValidationError struct {
	Errors []ValidationError
}

func (e *SequenceValidationError) Error() string {
	if len(e.Errors) == 1 {
		v := e.Errors[0]
		return fmt.Sprintf("invalid IUPAC DNA base %q at position %d", v.Got, v.Position)
	}
	return fmt.Sprintf(
		"%d invalid IUPAC DNA bases (first at position %d: %q)",
		len(e.Errors), e.Errors[0].Position, e.Errors[0].Got,
	)
}
