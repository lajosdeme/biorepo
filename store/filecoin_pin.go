package store

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// FilecoinPinStore implements Store by shelling out to the filecoin-pin CLI.
//
// Prerequisites:
//   - filecoin-pin binary must be in PATH (or set via WithCLIPath)
//   - PRIVATE_KEY env var must be set to a funded Filecoin wallet private key
//   - Run `filecoin-pin payments setup --auto` once before first use
//
// Get and Exists use an IPFS gateway (default: https://ipfs.io/ipfs/) since
// there is no filecoin-pin retrieval subcommand. The gateway is configurable
// via WithGateway.
//
// Delete is not supported by the Filecoin network and returns ErrNotSupported.
type FilecoinPinStore struct {
	cliPath    string
	gatewayURL string // must end with "/"
	extraEnv   []string
}

const fname = "data"

type FilecoinPinOption func(*FilecoinPinStore)

// WithCLIPath sets an explicit path to the filecoin-pin binary.
// Default: "filecoin-pin" (resolved from PATH).
func WithCLIPath(path string) FilecoinPinOption {
	return func(s *FilecoinPinStore) { s.cliPath = path }
}

// WithGateway sets the IPFS gateway base URL used for Get/Exists.
// Must end with "/". Default: "https://ipfs.io/ipfs/".
func WithGateway(url string) FilecoinPinOption {
	return func(s *FilecoinPinStore) {
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		s.gatewayURL = url
	}
}

// WithEnv appends extra environment variables (in "KEY=VALUE" form) to every
// CLI invocation. Use this to inject PRIVATE_KEY without relying on the
// process environment.
func WithEnv(env ...string) FilecoinPinOption {
	return func(s *FilecoinPinStore) { s.extraEnv = append(s.extraEnv, env...) }
}

func NewFilecoinPinStore(opts ...FilecoinPinOption) *FilecoinPinStore {
	s := &FilecoinPinStore{
		cliPath:    "filecoin-pin",
		gatewayURL: "https://ipfs.io/ipfs/",
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *FilecoinPinStore) Setup(ctx context.Context) error {
	_, err := s.run(ctx, "payments setup", "--auto")
	return err
}

// ── Put ──────────────────────────────────────────────────────────────────────

// Put writes data to a temporary file, uploads it with `filecoin-pin add
// --auto-fund`, and returns the Root CID from the command output.
func (s *FilecoinPinStore) Put(ctx context.Context, data []byte) (CID, error) {
	tmp, err := os.Create("data")
	if err != nil {
		return "", fmt.Errorf("filecoin-pin put: create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		return "", fmt.Errorf("filecoin-pin put: write temp file: %w", err)
	}
	// Close so the CLI can read it cleanly.
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("filecoin-pin put: close temp file: %w", err)
	}

	out, err := s.run(ctx, "add", tmp.Name(), "--auto-fund")
	if err != nil {
		return "", fmt.Errorf("filecoin-pin put: %w", err)
	}

	cid, err := parseRootCID(out)
	if err != nil {
		return "", fmt.Errorf("filecoin-pin put: parse output: %w\nraw output:\n%s", err, out)
	}
	return CID(cid), nil
}

// ── Get ──────────────────────────────────────────────────────────────────────

// Get fetches the raw bytes for cid via the configured IPFS gateway.
func (s *FilecoinPinStore) Get(ctx context.Context, cid CID) ([]byte, error) {
	url := s.gatewayURL + string(cid) + "/" + fname
	fmt.Println("URL: ", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("filecoin-pin get: build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("filecoin-pin get: gateway request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("filecoin-pin get: gateway returned %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("filecoin-pin get: read body: %w", err)
	}
	return data, nil
}

// ── Exists ───────────────────────────────────────────────────────────────────

// Exists checks for cid with a HEAD request to the IPFS gateway.
// A 200 means the content is resolvable; anything else is treated as absent.
func (s *FilecoinPinStore) Exists(ctx context.Context, cid CID) (bool, error) {
	url := s.gatewayURL + string(cid)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false, fmt.Errorf("filecoin-pin exists: build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Network error — we can't determine existence.
		return false, fmt.Errorf("filecoin-pin exists: gateway request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// ── List ─────────────────────────────────────────────────────────────────────

// List returns all Root CIDs across all datasets owned by the wallet.
//
// Implementation:
//  1. `filecoin-pin data-set --ls`  → parse dataset IDs
//  2. `filecoin-pin data-set <id>`  → parse Root CIDs from each dataset
//
// This is O(n datasets) CLI invocations. Fine for dev scale.
func (s *FilecoinPinStore) List(ctx context.Context) ([]CID, error) {
	lsOut, err := s.run(ctx, "data-set", "--ls")
	if err != nil {
		return nil, fmt.Errorf("filecoin-pin list: %w", err)
	}

	ids, err := parseDatasetIDs(lsOut)
	if err != nil {
		return nil, fmt.Errorf("filecoin-pin list: parse dataset list: %w\nraw output:\n%s", err, lsOut)
	}

	seen := make(map[CID]struct{})
	var cids []CID

	for _, id := range ids {
		dsOut, err := s.run(ctx, "data-set", strconv.Itoa(id))
		if err != nil {
			// A single dataset failing shouldn't abort the whole list.
			continue
		}
		rootCIDs, err := parseDatasetRootCIDs(dsOut)
		if err != nil {
			continue
		}
		for _, rc := range rootCIDs {
			c := CID(rc)
			if _, ok := seen[c]; !ok {
				seen[c] = struct{}{}
				cids = append(cids, c)
			}
		}
	}

	return cids, nil
}

// ── Delete ───────────────────────────────────────────────────────────────────

// Delete is not supported by the Filecoin network.
// Filecoin storage is append-only; data persists until the deal expires.
func (s *FilecoinPinStore) Delete(_ context.Context, _ CID) error {
	return ErrNotSupported
}

// ── CLI runner ───────────────────────────────────────────────────────────────

// run executes the filecoin-pin binary with args and returns combined stdout+stderr.
// A non-zero exit code is wrapped as an error that includes the output.
func (s *FilecoinPinStore) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, s.cliPath, args...)

	// Inherit the process environment so PRIVATE_KEY et al. are available,
	// then layer any caller-supplied overrides on top.
	cmd.Env = append(os.Environ(), s.extraEnv...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command %q failed: %w\noutput:\n%s",
			strings.Join(append([]string{s.cliPath}, args...), " "),
			err, buf.String())
	}
	return buf.String(), nil
}

// ── Output parsers ───────────────────────────────────────────────────────────

// Example relevant lines from `filecoin-pin add`:
//
//	Root CID:   bafybeibh422kjvgfmymx6nr7jandwngrown6ywomk4vplayl4de2x553t4
//
// We grab the first occurrence since a single-file upload produces exactly one.
var reRootCID = regexp.MustCompile(`Root CID:\s+(baf[a-zA-Z0-9]+)`)

func parseRootCID(output string) (string, error) {
	m := reRootCID.FindStringSubmatch(output)
	if len(m) < 2 {
		return "", fmt.Errorf("Root CID not found in output")
	}
	return m[1], nil
}

// Example relevant lines from `filecoin-pin data-set --ls`:
//
//	# 325 • live • managed
//
// We extract the numeric IDs.
var reDatasetID = regexp.MustCompile(`#(\d+)\s+[•·]`)

func parseDatasetIDs(output string) ([]int, error) {
	matches := reDatasetID.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		// Empty store is valid — not an error.
		return nil, nil
	}
	ids := make([]int, 0, len(matches))
	for _, m := range matches {
		id, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Example relevant lines from `filecoin-pin data-set <id>`:
//
//	#0
//	  CommP: bafkzcib...
//	  Root CID: bafybeibh422kjvgfmymx6nr7jandwngrown6ywomk4vplayl4de2x553t4
//
// We collect all Root CID values in the Pieces section.
var rePieceRootCID = regexp.MustCompile(`Root CID:\s+(baf[a-zA-Z0-9]+)`)

func parseDatasetRootCIDs(output string) ([]string, error) {
	matches := rePieceRootCID.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	cids := make([]string, 0, len(matches))
	for _, m := range matches {
		cids = append(cids, m[1])
	}
	return cids, nil
}
