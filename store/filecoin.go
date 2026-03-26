package store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/lighthouse-web3/lighthouse-go-sdk/lighthouse"
)

type FilecoinStore struct {
	c *lighthouse.Client
}

func NewFilecoinStore() *FilecoinStore {
	client := lighthouse.NewClient(nil,
		lighthouse.WithAPIKey(os.Getenv("LIGHTHOUSE_API_KEY")),
	)
	return &FilecoinStore{c: client}
}

// Put stores data and returns its CID.
// Uploads the data as a temporary file, then uploads to Lighthouse.
func (s *FilecoinStore) Put(ctx context.Context, data []byte) (CID, error) {
	// Create a temporary file to upload
	tmpFile, err := os.CreateTemp("", "lighthouse-upload-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Upload the file
	upload, err := s.c.Storage().UploadFile(ctx, tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to upload to lighthouse: %w", err)
	}

	return CID(upload.Hash), nil
}

// Get retrieves the raw bytes for a given CID.
func (s *FilecoinStore) Get(ctx context.Context, cid CID) ([]byte, error) {
	// First verify the CID exists
	_, err := s.c.Files().Info(ctx, string(cid))
	if err != nil {
		// Check if error indicates not found
		// Lighthouse SDK may return specific errors, this is a best-effort check
		return nil, ErrNotFound
	}

	// Download from IPFS gateway
	ipfsGatewayURL := fmt.Sprintf("https://gateway.lighthouse.storage/ipfs/%s", cid)

	req, err := http.NewRequestWithContext(ctx, "GET", ipfsGatewayURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gateway returned status: %s", resp.Status)
	}

	// Read all data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// Exists reports whether the given CID is present in the store.
func (s *FilecoinStore) Exists(ctx context.Context, cid CID) (bool, error) {
	_, err := s.c.Files().Info(ctx, string(cid))
	if err != nil {
		// Check if error indicates not found
		// Lighthouse returns errors for non-existent files
		// We'll assume any error means not found for existence check
		return false, nil
	}
	return true, nil
}

// List returns all CIDs currently held in the store.
// Note: Lighthouse's List returns files with pagination. This implementation
// collects all pages and extracts CIDs.
func (s *FilecoinStore) List(ctx context.Context) ([]CID, error) {
	var allCIDs []CID
	var lastKey *string

	for {
		// List files with current pagination key
		listUploads, err := s.c.Files().List(ctx, lastKey)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		// Extract CIDs from the current page
		for _, file := range listUploads.Data {
			if file.CID != "" {
				allCIDs = append(allCIDs, CID(file.CID))
			}
		}

		// Check if there are more pages
		if listUploads.LastKey == nil {
			break
		}
		lastKey = listUploads.LastKey
	}

	return allCIDs, nil
}

// Delete removes the content for the given CID.
// Note: Lighthouse uses file IDs rather than CIDs for deletion. This creates a
// limitation - we need to find the file ID for a given CID first.
func (s *FilecoinStore) Delete(ctx context.Context, cid CID) error {
	// We need to search through the list to find the file ID for this CID.
	// This is inefficient but necessary given the API structure.

	// For large stores, we might want to cache CID->ID mappings.
	// Let's implement a search through paginated results.
	var lastKey *string

	for {
		listUploads, err := s.c.Files().List(ctx, lastKey)
		if err != nil {
			return fmt.Errorf("failed to list files for deletion: %w", err)
		}

		// Search for the file with matching CID
		for _, file := range listUploads.Data {
			if file.CID == string(cid) {
				// Found the file, delete by ID
				if err := s.c.Files().Delete(ctx, file.ID); err != nil {
					return fmt.Errorf("failed to delete file: %w", err)
				}
				return nil
			}
		}

		// Check if there are more pages
		if listUploads.LastKey == nil {
			break
		}
		lastKey = listUploads.LastKey
	}

	// If we get here, CID wasn't found
	return ErrNotFound
}
