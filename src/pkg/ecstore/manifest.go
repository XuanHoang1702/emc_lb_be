package ecstore

import (
	"encoding/json"
	"fmt"
)

// Manifest is the root of trust for one stored object. It records the coding
// parameters, the exact original size (Split zero-pads the last shard), and
// the SHA-256 digest of every shard plus the whole object.
//
// The manifest is written LAST during Put so a crash mid-write can never
// expose a half-recorded object: shards without a manifest are orphans that
// garbage collection may reclaim.
type Manifest struct {
	ID        string   `json:"id"`
	Size      int      `json:"size"`
	K         int      `json:"k"`
	M         int      `json:"m"`
	ShardSize int      `json:"shard_size"`
	ShardSums []string `json:"shard_sums"`
	DataSum   string   `json:"data_sum"`
}

// Validate checks structural invariants of a decoded manifest.
func (m Manifest) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("ecstore: manifest has empty object id")
	}
	if m.Size <= 0 {
		return fmt.Errorf("ecstore: manifest size must be > 0, got %d", m.Size)
	}
	if m.K < 1 || m.M < 1 {
		return fmt.Errorf("ecstore: manifest shard counts invalid, got k=%d m=%d", m.K, m.M)
	}
	if len(m.ShardSums) != m.K+m.M {
		return fmt.Errorf("ecstore: manifest has %d shard sums, want %d", len(m.ShardSums), m.K+m.M)
	}
	if m.DataSum == "" {
		return fmt.Errorf("ecstore: manifest has empty data sum")
	}
	return nil
}

// MarshalManifest serializes a manifest for placement on the metadata node.
func MarshalManifest(m Manifest) ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("ecstore: marshal manifest: %w", err)
	}

	return data, nil
}

// UnmarshalManifest decodes and validates a manifest.
func UnmarshalManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("ecstore: unmarshal manifest: %w", err)
	}

	if err := m.Validate(); err != nil {
		return Manifest{}, err
	}

	return m, nil
}
