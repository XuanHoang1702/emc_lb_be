package ecstore

import "errors"

var (
	// ErrInvalidObjectID is returned when an object id is empty.
	ErrInvalidObjectID = errors.New("ecstore: object id must not be empty")
	// ErrEmptyData is returned when Put receives zero bytes.
	ErrEmptyData = errors.New("ecstore: empty objects are not storable")
	// ErrObjectNotFound is returned when no manifest exists for an id.
	ErrObjectNotFound = errors.New("ecstore: object not found")
	// ErrInsufficientShards is returned when fewer than k valid shards remain.
	ErrInsufficientShards = errors.New("ecstore: insufficient valid shards to reconstruct")
	// ErrChecksumMismatch is returned when reconstructed data fails integrity checks.
	ErrChecksumMismatch = errors.New("ecstore: checksum mismatch")
	// ErrNodeDown is returned by a failed placement node.
	ErrNodeDown = errors.New("ecstore: node is down")
	// ErrShardNotFound is returned when a node holds no copy of a shard.
	ErrShardNotFound = errors.New("ecstore: shard not found")
)
