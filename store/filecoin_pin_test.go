package store

import (
	"context"
	"testing"
)

// ── Parser unit tests ─────────────────────────────────────────────────────────
// These test only the output-parsing logic and require no binary or network.

func TestParseRootCID(t *testing.T) {
	// Realistic output from `filecoin-pin add demo.txt --auto-fund`
	output := `
Uploading to Filecoin...

  Root CID:    bafybeibh422kjvgfmymx6nr7jandwngrown6ywomk4vplayl4de2x553t4
  Piece CID:   bafkzcibcfab4grpgq6e6rva4kfuxfcvibdzx3kn2jdw6q3zqgwt5cou7j6k4wfq
  Piece ID:    0
  Data Set ID: 325
  Transaction: 0xc85e49d2ed745cc8c5d7115e7c45a1243ec25da7e73e224a744887783afea42b

✅ Your file is now stored on Filecoin with ongoing proof of possession!
`
	cid, err := parseRootCID(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "bafybeibh422kjvgfmymx6nr7jandwngrown6ywomk4vplayl4de2x553t4"
	if cid != want {
		t.Errorf("got %q, want %q", cid, want)
	}
}

func TestParseRootCID_Missing(t *testing.T) {
	_, err := parseRootCID("no CID here at all")
	if err == nil {
		t.Fatal("expected error for missing Root CID, got nil")
	}
}

func TestParseDatasetIDs(t *testing.T) {
	// Realistic output from `filecoin-pin data-set --ls`
	output := `
━━━ Data Sets ━━━
│
│  Address: 0x5a0...B7B0B
│  Network: calibration
│
│  #325 • live • managed
│    Provider: infrafolio-calib (ID 4)
│    Pieces stored: 3
│
│  #412 • live • managed
│    Provider: ezpdpz-calib (ID 3)
│    Pieces stored: 1
│
└  Data set inspection complete
`
	ids, err := parseDatasetIDs(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 dataset IDs, got %d: %v", len(ids), ids)
	}
	if ids[0] != 325 || ids[1] != 412 {
		t.Errorf("got %v, want [325 412]", ids)
	}
}

func TestParseDatasetIDs_Empty(t *testing.T) {
	output := `
━━━ Data Sets ━━━
│  Address: 0x5a0...B7B0B
│  No data sets found.
└  Data set inspection complete
`
	ids, err := parseDatasetIDs(output)
	if err != nil {
		t.Fatalf("unexpected error on empty dataset list: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty slice, got %v", ids)
	}
}

func TestParseDatasetRootCIDs(t *testing.T) {
	// Realistic output from `filecoin-pin data-set 325`
	output := `
Filecoin Onchain Cloud Data Sets

━━━ Data Sets ━━━

Data Set #325 • live
  Pieces stored: 2

Pieces
  Total pieces: 2

  #0
    CommP: bafkzcibcfab4grpgq6e6rva4kfuxfcvibdzx3kn2jdw6q3zqgwt5cou7j6k4wfq
    Root CID: bafybeibh422kjvgfmymx6nr7jandwngrown6ywomk4vplayl4de2x553t4
  #1
    CommP: bafkzcibcjmcnyio2ocxhmtq34uh5ct425xzpnor532zku7tjvqf5toodbxtsqhi
    Root CID: bafybeig27btater5fpt3l67gbme3sebqk3ynwdhlbrbuk3q7espiyplan4

Data set inspection complete
`
	cids, err := parseDatasetRootCIDs(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cids) != 2 {
		t.Fatalf("expected 2 CIDs, got %d: %v", len(cids), cids)
	}
	want := []string{
		"bafybeibh422kjvgfmymx6nr7jandwngrown6ywomk4vplayl4de2x553t4",
		"bafybeig27btater5fpt3l67gbme3sebqk3ynwdhlbrbuk3q7espiyplan4",
	}
	for i, w := range want {
		if cids[i] != w {
			t.Errorf("[%d] got %q, want %q", i, cids[i], w)
		}
	}
}

func TestParseDatasetRootCIDs_Empty(t *testing.T) {
	cids, err := parseDatasetRootCIDs("Data set #99 • live\n  Pieces stored: 0\n")
	if err != nil {
		t.Fatalf("unexpected error on empty pieces: %v", err)
	}
	if len(cids) != 0 {
		t.Errorf("expected empty slice, got %v", cids)
	}
}

// ── Constructor / option tests ────────────────────────────────────────────────

func TestNewFilecoinPinStore_Defaults(t *testing.T) {
	s := NewFilecoinPinStore()
	if s.cliPath != "filecoin-pin" {
		t.Errorf("default cliPath: got %q, want %q", s.cliPath, "filecoin-pin")
	}
	if s.gatewayURL != "https://ipfs.io/ipfs/" {
		t.Errorf("default gatewayURL: got %q, want %q", s.gatewayURL, "https://ipfs.io/ipfs/")
	}
}

func TestNewFilecoinPinStore_Options(t *testing.T) {
	s := NewFilecoinPinStore(
		WithCLIPath("/usr/local/bin/filecoin-pin"),
		WithGateway("https://gateway.lighthouse.storage/ipfs"),
		WithEnv("PRIVATE_KEY=0xdead"),
	)
	if s.cliPath != "/usr/local/bin/filecoin-pin" {
		t.Errorf("cliPath: got %q", s.cliPath)
	}
	// WithGateway should append trailing slash if missing
	if s.gatewayURL != "https://gateway.lighthouse.storage/ipfs/" {
		t.Errorf("gatewayURL: got %q", s.gatewayURL)
	}
	if len(s.extraEnv) != 1 || s.extraEnv[0] != "PRIVATE_KEY=0xdead" {
		t.Errorf("extraEnv: got %v", s.extraEnv)
	}
}

func TestFilecoinPinStore_Delete_NotSupported(t *testing.T) {
	s := NewFilecoinPinStore()
	err := s.Delete(context.Background(), CID("bafybeianything"))
	if err != ErrNotSupported {
		t.Errorf("Delete should return ErrNotSupported, got %v", err)
	}
}
