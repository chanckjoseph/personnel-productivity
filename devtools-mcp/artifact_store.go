package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ArtifactStore manages persistence of Query artifacts to disk.
type ArtifactStore struct {
	cacheDir  string
	queriesDir string
	manifestPath string
	manifest  Manifest
}

// Manifest tracks all saved queries for indexing and reference.
type Manifest struct {
	Version   string          `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Queries   []ManifestEntry `json:"queries"`
}

// ManifestEntry represents a single query in the manifest.
type ManifestEntry struct {
	QueryID     string    `json:"query_id"`
	Filename    string    `json:"filename"`
	OriginalText string   `json:"original_text"`
	CreatedAt   time.Time `json:"created_at"`
	Intent      string    `json:"intent"`
	Domain      string    `json:"domain"`
	Urgency     string    `json:"urgency"`
	Confidence  float64   `json:"confidence"`
	TaskCount   int       `json:"task_count"`
	Status      string    `json:"status"`
}

// NewArtifactStore creates a new artifact store with given base path.
func NewArtifactStore(baseDir string) (*ArtifactStore, error) {
	cacheDir := filepath.Join(baseDir, ".cache")
	queriesDir := filepath.Join(cacheDir, "queries")

	// Create directories if they don't exist
	if err := os.MkdirAll(queriesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	manifestPath := filepath.Join(queriesDir, "manifest.json")

	store := &ArtifactStore{
		cacheDir:     cacheDir,
		queriesDir:   queriesDir,
		manifestPath: manifestPath,
	}

	// Load or create manifest
	if err := store.loadManifest(); err != nil {
		return nil, err
	}

	return store, nil
}

// SaveQuery persists a Query to disk and updates manifest.
func (as *ArtifactStore) SaveQuery(q *Query) error {
	// Generate filename based on query ID
	filename := fmt.Sprintf("%s.json", q.ID)
	filepath := filepath.Join(as.queriesDir, filename)

	// Marshal Query to JSON
	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write query file: %w", err)
	}

	// Update manifest
	entry := ManifestEntry{
		QueryID:     q.ID,
		Filename:    filename,
		OriginalText: q.OriginalText,
		CreatedAt:   q.Metadata.CreatedAt,
		Intent:      q.Intent.Primary,
		Domain:      q.Intent.Domain,
		Urgency:     q.Intent.Urgency,
		Confidence:  q.Metadata.Confidence,
		TaskCount:   q.Metadata.TaskCount,
		Status:      q.Metadata.Status,
	}

	as.manifest.Queries = append(as.manifest.Queries, entry)
	as.manifest.UpdatedAt = time.Now()

	// Save manifest to disk
	if err := as.saveManifest(); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

// LoadQuery retrieves a saved Query by ID.
func (as *ArtifactStore) LoadQuery(queryID string) (*Query, error) {
	filepath := filepath.Join(as.queriesDir, fmt.Sprintf("%s.json", queryID))

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read query file: %w", err)
	}

	var q Query
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("failed to unmarshal query: %w", err)
	}

	return &q, nil
}

// ListQueries returns all saved queries sorted by creation time (newest first).
func (as *ArtifactStore) ListQueries(limit int) []ManifestEntry {
	// Sort by creation time (newest first)
	queries := make([]ManifestEntry, len(as.manifest.Queries))
	copy(queries, as.manifest.Queries)

	sort.Slice(queries, func(i, j int) bool {
		return queries[i].CreatedAt.After(queries[j].CreatedAt)
	})

	// Apply limit
	if limit > 0 && len(queries) > limit {
		queries = queries[:limit]
	}

	return queries
}

// FindQueries searches for queries matching criteria.
func (as *ArtifactStore) FindQueries(domain, urgency string) []ManifestEntry {
	var results []ManifestEntry

	for _, entry := range as.manifest.Queries {
		match := true

		if domain != "" && entry.Domain != domain {
			match = false
		}

		if urgency != "" && entry.Urgency != urgency {
			match = false
		}

		if match {
			results = append(results, entry)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	return results
}

// SearchQueries searches queries by text pattern in original prompt.
func (as *ArtifactStore) SearchQueries(pattern string) []ManifestEntry {
	var results []ManifestEntry

	for _, entry := range as.manifest.Queries {
		// Simple substring search (could be enhanced with regex)
		if runesContains([]rune(entry.OriginalText), []rune(pattern)) {
			results = append(results, entry)
		}
	}

	return results
}

// DeleteQuery removes a query artifact and updates manifest.
func (as *ArtifactStore) DeleteQuery(queryID string) error {
	filepath := filepath.Join(as.queriesDir, fmt.Sprintf("%s.json", queryID))

	// Remove file
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete query file: %w", err)
	}

	// Update manifest by removing entry
	for i, entry := range as.manifest.Queries {
		if entry.QueryID == queryID {
			as.manifest.Queries = append(as.manifest.Queries[:i], as.manifest.Queries[i+1:]...)
			break
		}
	}

	as.manifest.UpdatedAt = time.Now()

	// Save manifest
	if err := as.saveManifest(); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

// GetStats returns statistics about stored queries.
func (as *ArtifactStore) GetStats() map[string]interface{} {
	domainCount := make(map[string]int)
	urgencyCount := make(map[string]int)
	totalTasks := 0
	totalConfidence := 0.0

	for _, entry := range as.manifest.Queries {
		domainCount[entry.Domain]++
		urgencyCount[entry.Urgency]++
		totalTasks += entry.TaskCount
		totalConfidence += entry.Confidence
	}

	avgConfidence := 0.0
	if len(as.manifest.Queries) > 0 {
		avgConfidence = totalConfidence / float64(len(as.manifest.Queries))
	}

	return map[string]interface{}{
		"total_queries":   len(as.manifest.Queries),
		"by_domain":       domainCount,
		"by_urgency":      urgencyCount,
		"total_tasks":     totalTasks,
		"avg_confidence":  avgConfidence,
		"manifest_path":   as.manifestPath,
		"queries_dir":     as.queriesDir,
	}
}

// loadManifest reads the manifest file if it exists.
func (as *ArtifactStore) loadManifest() error {
	as.manifest = Manifest{
		Version:   "1.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Queries:   []ManifestEntry{},
	}

	// If manifest doesn't exist yet, that's ok
	if _, err := os.Stat(as.manifestPath); os.IsNotExist(err) {
		return nil
	}

	data, err := ioutil.ReadFile(as.manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	if err := json.Unmarshal(data, &as.manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	return nil
}

// saveManifest writes the manifest file to disk.
func (as *ArtifactStore) saveManifest() error {
	data, err := json.MarshalIndent(as.manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := ioutil.WriteFile(as.manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// containsRunes checks if pattern is a substring (simple implementation).
func runesContains(haystack, needle []rune) bool {
	if len(needle) == 0 {
		return true
	}
	if len(needle) > len(haystack) {
		return false
	}

	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}
