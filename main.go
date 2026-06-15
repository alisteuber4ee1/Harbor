package main

import (
	"fmt"
	"os"
	"time"
)

// Blob represents a storage blob.
type Blob struct {
	ID        string
	CreatedAt time.Time
}

// UploadSession represents an active upload session.
type UploadSession struct {
	ID     string
	BlobID string
	Active bool
}

// GarbageCollect performs the GC sweep, returning the list of deleted blob IDs.
func GarbageCollect(blobs []Blob, referencedBlobs map[string]bool, activeSessions []UploadSession, gracePeriod time.Duration) []string {
	now := time.Now()
	activeBlobIDs := make(map[string]bool)
	for _, session := range activeSessions {
		if session.Active {
			activeBlobIDs[session.BlobID] = true
		}
	}

	var deleted []string
	for _, blob := range blobs {
		// Exclude referenced blobs
		if referencedBlobs[blob.ID] {
			continue
		}

		// Exclude blobs within the grace period
		if now.Sub(blob.CreatedAt) < gracePeriod {
			continue
		}

		// Exclude blobs associated with active upload sessions
		if activeBlobIDs[blob.ID] {
			continue
		}

		deleted = append(deleted, blob.ID)
	}

	return deleted
}

func main() {
	gracePeriodStr := os.Getenv("GC_GRACE_PERIOD")
	gracePeriod := 2 * time.Hour
	if gracePeriodStr != "" {
		if d, err := time.ParseDuration(gracePeriodStr); err == nil {
			gracePeriod = d
		}
	}

	fmt.Printf("Starting GC with grace period: %v\n", gracePeriod)

	now := time.Now()
	blobs := []Blob{
		{ID: "blob-referenced", CreatedAt: now.Add(-5 * time.Hour)},
		{ID: "blob-recent", CreatedAt: now.Add(-30 * time.Minute)},
		{ID: "blob-active-session", CreatedAt: now.Add(-3 * time.Hour)},
		{ID: "blob-orphaned-old", CreatedAt: now.Add(-5 * time.Hour)},
	}

	referencedBlobs := map[string]bool{
		"blob-referenced": true,
	}

	activeSessions := []UploadSession{
		{ID: "session-1", BlobID: "blob-active-session", Active: true},
	}

	deleted := GarbageCollect(blobs, referencedBlobs, activeSessions, gracePeriod)

	fmt.Println("Deleted blobs:")
	for _, id := range deleted {
		fmt.Println("-", id)
	}
}
