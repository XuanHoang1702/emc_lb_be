package ecstore

import (
	"crypto/sha256"
	"encoding/hex"
)

// ShardSum returns the hex-encoded SHA-256 digest of a single shard.
//
// The Reed-Solomon encoder cannot tell WHICH shard is corrupt — it only
// checks parity consistency across the whole set. Per-shard checksums are
// therefore mandatory before feeding gathered shards to Reconstruct.
func ShardSum(shard []byte) string {
	sum := sha256.Sum256(shard)
	return hex.EncodeToString(sum[:])
}

// DataSum returns the hex-encoded SHA-256 digest of the reconstructed object.
func DataSum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// VerifyShard reports whether a shard matches its recorded checksum.
func VerifyShard(sum string, shard []byte) bool {
	if sum == "" {
		return false
	}
	return ShardSum(shard) == sum
}
