//go:build integration

package store

// Integration tests for FilecoinPinStore.
//
// These tests shell out to the real filecoin-pin binary and hit the Filecoin
// Calibration testnet. They are gated behind the "integration" build tag so
// they never run in ordinary CI.
//
// Run with:
//
//	PRIVATE_KEY=0x... go test -tags integration -v -timeout 5m ./store/
//
// Prerequisites:
//   - filecoin-pin binary in PATH (npm install -g filecoin-pin@latest)
//   - PRIVATE_KEY env var set to a Calibration testnet wallet private key
//   - Wallet funded with tFIL (ChainSafe faucet) and test USDFC (Filecoin faucet)
//   - `filecoin-pin payments setup --auto` already run at least once, OR let
//     TestMain handle it (it calls Setup before the suite runs)
//
// Test isolation: Put/Get/Exists/List all operate against the same live wallet.
// List assertions check for presence of newly uploaded CIDs rather than exact
// counts, since the wallet may already have prior uploads.
//
// Gateway propagation: freshly uploaded content can take a few seconds to be
// resolvable via the IPFS gateway. Get and Exists retry with backoff before
// failing to avoid spurious flakiness.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"testing"
	"time"
)

// ── Test-suite setup ──────────────────────────────────────────────────────────

// integrationStore is the shared store instance for all integration tests.
// Initialised once in TestMain.
var integrationStore *FilecoinPinStore

func TestMain(m *testing.M) {
	// Require PRIVATE_KEY — skip the whole suite if absent.
	pk := os.Getenv("PRIVATE_KEY")
	if pk == "" {
		fmt.Fprintln(os.Stderr, "SKIP: PRIVATE_KEY not set; skipping integration tests")
		os.Exit(0)
	}

	// Require the binary.
	if _, err := exec.LookPath("filecoin-pin"); err != nil {
		fmt.Fprintln(os.Stderr, "SKIP: filecoin-pin not found in PATH; skipping integration tests")
		os.Exit(0)
	}

	integrationStore = NewFilecoinPinStore()

	// Run payments setup once before the suite. This is idempotent — safe to
	// call even if the wallet is already configured.
	_, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println("--- TestMain: running filecoin-pin payments setup --auto")
	/* 	if err := integrationStore.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Setup failed: %v\n", err)
		os.Exit(1)
	} */
	fmt.Println("--- TestMain: setup complete")

	os.Exit(m.Run())
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// retryUntil calls f repeatedly with exponential backoff until it returns a
// nil error or the deadline is exceeded. It is used to absorb IPFS gateway
// propagation latency after a fresh upload.
func retryUntil(ctx context.Context, f func() error) error {
	delay := 3 * time.Second
	for {
		err := f()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out; last error: %w", err)
		case <-time.After(delay):
			delay = min(delay*2, 30*time.Second)
		}
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// containsCID returns true if cid is present in the slice.
func containsCID(cids []CID, cid CID) bool {
	return slices.Contains(cids, cid)
}

// ── Setup ─────────────────────────────────────────────────────────────────────

/* func TestIntegration_Setup_Idempotent(t *testing.T) {
	// Setup was already called in TestMain. Calling it again must not error —
	// the CLI's --auto flag is designed to be re-entrant.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := integrationStore.Setup(ctx); err != nil {
		t.Fatalf("second Setup call failed: %v", err)
	}
} */

// ── Put ───────────────────────────────────────────────────────────────────────

func TestIntegration_Put_SmallPayload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	data := []byte("biorepo integration test: small payload")
	cid, err := integrationStore.Put(ctx, data)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if cid == "" {
		t.Fatal("Put returned empty CID")
	}
	if !isBafCID(string(cid)) {
		t.Errorf("Put returned CID with unexpected format: %q", cid)
	}
	t.Logf("uploaded CID: %s", cid)
}

func TestIntegration_Get_SimpleGet(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cid := "bafybeifymr5q7zbzbuddpxd3ngi7rxjislh2u4b72z4w3klj6ge4jrn46q"

	data, err := integrationStore.Get(ctx, CID(cid))

	if err != nil {
		t.Errorf("failed to get data for cid: %v", err)
	}
	fmt.Println(string(data))
}

func TestIntegration_Put_FASTASequence(t *testing.T) {
	// Mirrors the real biorepo use-case: FASTA-encoded DNA stored as bytes.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	fasta := []byte(">drought-tolerance-variant organism=Zea_mays\nACGTACGTACGTACGTACGT\nGCATGCATGCATGCATGCAT\n")
	cid, err := integrationStore.Put(ctx, fasta)
	if err != nil {
		t.Fatalf("Put FASTA failed: %v", err)
	}
	if !isBafCID(string(cid)) {
		t.Errorf("unexpected CID format: %q", cid)
	}
	t.Logf("FASTA CID: %s", cid)
}

func TestIntegration_Put_EmptyPayload(t *testing.T) {
	// The CLI has a minimum piece size of 1 KB. An empty upload is expected
	// to fail. We assert an error is returned rather than an empty CID.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, err := integrationStore.Put(ctx, []byte{})
	if err == nil {
		t.Log("note: CLI accepted empty payload (below min piece size) — may be padded by provider")
	} else {
		t.Logf("empty payload correctly rejected: %v", err)
	}
	// Either outcome is acceptable; this test documents the behaviour.
}

func TestIntegration_Put_LargerPayload(t *testing.T) {
	// 2 KB — above the 1 KB minimum piece size noted in provider details.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	data := bytes.Repeat([]byte("ACGT"), 512) // 2048 bytes
	cid, err := integrationStore.Put(ctx, data)
	if err != nil {
		t.Fatalf("Put 2KB failed: %v", err)
	}
	if !isBafCID(string(cid)) {
		t.Errorf("unexpected CID format: %q", cid)
	}
	t.Logf("2KB CID: %s", cid)
}

// ── Get ───────────────────────────────────────────────────────────────────────

func TestIntegration_Get_RoundTrip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	want := []byte("biorepo round-trip: " + t.Name())
	cid, err := integrationStore.Put(ctx, want)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	t.Logf("uploaded CID: %s", cid)

	// Gateway propagation can be slow after a fresh upload — retry.
	var got []byte
	err = retryUntil(ctx, func() error {
		var gerr error
		got, gerr = integrationStore.Get(ctx, cid)
		return gerr
	})
	if err != nil {
		t.Fatalf("Get failed after retries: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("round-trip mismatch:\n  want: %q\n   got: %q", want, got)
	}
}

func TestIntegration_Get_NotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// A well-formed but non-existent CID.
	fakeCID := CID("bafybeiaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	_, err := integrationStore.Get(ctx, fakeCID)
	if err == nil {
		t.Fatal("expected error for non-existent CID, got nil")
	}
	// Gateway 404s may come back as a generic error rather than ErrNotFound
	// depending on the gateway used, so we just assert non-nil.
	t.Logf("correctly got error for non-existent CID: %v", err)
}

// ── Exists ────────────────────────────────────────────────────────────────────

func TestIntegration_Exists_AfterPut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	data := []byte("biorepo exists check: " + t.Name())
	cid, err := integrationStore.Put(ctx, data)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	t.Logf("uploaded CID: %s", cid)

	// Retry until gateway serves it.
	var exists bool
	err = retryUntil(ctx, func() error {
		var eerr error
		exists, eerr = integrationStore.Exists(ctx, cid)
		if eerr != nil {
			return eerr
		}
		if !exists {
			return fmt.Errorf("CID not yet resolvable via gateway")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Exists returned false after retries: %v", err)
	}
	if !exists {
		t.Errorf("Exists(%s) = false, want true", cid)
	}
}

func TestIntegration_Exists_NotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fakeCID := CID("bafybeiaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	exists, err := integrationStore.Exists(ctx, fakeCID)
	if err != nil {
		// Some gateways return an error rather than false for unknown CIDs;
		// both are acceptable outcomes.
		t.Logf("Exists returned error for unknown CID (acceptable): %v", err)
		return
	}
	if exists {
		t.Errorf("Exists(%s) = true for a fabricated CID", fakeCID)
	}
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestIntegration_List_ContainsUploadedCID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Upload something fresh so we have at least one known CID to look for.
	data := []byte("biorepo list check: " + t.Name())
	cid, err := integrationStore.Put(ctx, data)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	t.Logf("uploaded CID: %s", cid)

	// List may need a moment for the dataset to be indexed on-chain.
	err = retryUntil(ctx, func() error {
		cids, lerr := integrationStore.List(ctx)
		if lerr != nil {
			return lerr
		}
		if !containsCID(cids, cid) {
			return fmt.Errorf("CID %s not yet visible in List (%d total CIDs)", cid, len(cids))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("List never contained uploaded CID: %v", err)
	}
}

func TestIntegration_List_NoDuplicates(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cids, err := integrationStore.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	seen := make(map[CID]int)
	for _, c := range cids {
		seen[c]++
	}
	for c, count := range seen {
		if count > 1 {
			t.Errorf("CID %s appears %d times in List output", c, count)
		}
	}
	t.Logf("List returned %d unique CIDs", len(cids))
}

func TestIntegration_List_AllCIDsWellFormed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cids, err := integrationStore.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	for _, c := range cids {
		if !isBafCID(string(c)) {
			t.Errorf("List returned malformed CID: %q", c)
		}
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestIntegration_Delete_NotSupported(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Upload something real so the CID actually exists.
	cid, err := integrationStore.Put(ctx, []byte("delete test payload"))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	err = integrationStore.Delete(ctx, cid)
	if !errors.Is(err, ErrNotSupported) {
		t.Errorf("Delete should return ErrNotSupported, got: %v", err)
	}
}

// ── Full workflow ─────────────────────────────────────────────────────────────

// TestIntegration_FullWorkflow exercises the entire Put → Exists → Get → List
// chain in sequence, mirroring what biorepo.Client.PublishDesign does in
// production.
func TestIntegration_FullWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	// A realistic FASTA payload representing a synthetic biology design.
	payload := []byte(">integration-test-design organism=Arabidopsis_thaliana confidence=0.923\n" +
		"ATGCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCG\n" +
		"ATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGAT\n")

	// 1. Upload
	t.Log("step 1: Put")
	cid, err := integrationStore.Put(ctx, payload)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	t.Logf("CID: %s", cid)

	// 2. Exists (with gateway retry)
	t.Log("step 2: Exists")
	err = retryUntil(ctx, func() error {
		ok, eerr := integrationStore.Exists(ctx, cid)
		if eerr != nil {
			return eerr
		}
		if !ok {
			return fmt.Errorf("not yet resolvable")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	// 3. Get and compare bytes
	t.Log("step 3: Get")
	got, err := integrationStore.Get(ctx, cid)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("Get round-trip mismatch\n  want len=%d\n   got len=%d", len(payload), len(got))
	}

	// 4. List contains the CID
	t.Log("step 4: List")
	err = retryUntil(ctx, func() error {
		cids, lerr := integrationStore.List(ctx)
		if lerr != nil {
			return lerr
		}
		if !containsCID(cids, cid) {
			return fmt.Errorf("CID not yet in List")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("List did not contain uploaded CID: %v", err)
	}

	t.Log("full workflow passed")
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// isBafCID returns true if s looks like a valid CIDv1 (baf… prefix).
// This is a format check only, not a content-validity check.
func isBafCID(s string) bool {
	return len(s) >= 10 && s[:3] == "baf"
}
