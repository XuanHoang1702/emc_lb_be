// Package ecstore provides fault-tolerant storage for immutable byte objects
// using Reed-Solomon erasure coding (k data shards + m parity shards).
//
// WARNING: never use ecstore for mutable financial state (wallet balances,
// payments, orders, ledger records). The primary database remains the single
// source of truth for transactional data. ecstore targets immutable artifacts
// only (documents, images, backup archives).
package ecstore

import (
	"bytes"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

// Codec abstracts the erasure-coding engine. Split returns equally sized,
// zero-padded data shards; Encode fills the parity slots in place;
// Reconstruct rebuilds nil (missing) shards when at least k are present.
type Codec interface {
	Split(data []byte) ([][]byte, error)
	Encode(shards [][]byte) error
	Verify(shards [][]byte) (bool, error)
	Reconstruct(shards [][]byte) error
	Join(shards [][]byte, dataSize int) ([]byte, error)
	DataShards() int
	ParityShards() int
}

// rsCodec is a Reed-Solomon implementation of Codec backed by
// github.com/klauspost/reedsolomon — the same engine used by MinIO.
type rsCodec struct {
	enc reedsolomon.Encoder
	k   int
	m   int
}

// NewRSCoder creates a systematic Reed-Solomon coder with k data shards and
// m parity shards. Any k of the k+m shards reconstruct the original data.
func NewRSCoder(k, m int) (Codec, error) {
	if k < 1 || m < 1 {
		return nil, fmt.Errorf("ecstore: shard counts must be >= 1, got k=%d m=%d", k, m)
	}

	enc, err := reedsolomon.New(k, m)
	if err != nil {
		return nil, fmt.Errorf("ecstore: create reed-solomon encoder: %w", err)
	}

	return &rsCodec{enc: enc, k: k, m: m}, nil
}

func (c *rsCodec) DataShards() int   { return c.k }
func (c *rsCodec) ParityShards() int { return c.m }

func (c *rsCodec) Split(data []byte) ([][]byte, error) {
	shards, err := c.enc.Split(data)
	if err != nil {
		return nil, fmt.Errorf("ecstore: split data into %d shards: %w", c.k, err)
	}
	return shards, nil
}

func (c *rsCodec) Encode(shards [][]byte) error {
	if err := c.enc.Encode(shards); err != nil {
		return fmt.Errorf("ecstore: encode parity shards: %w", err)
	}
	return nil
}

func (c *rsCodec) Verify(shards [][]byte) (bool, error) {
	ok, err := c.enc.Verify(shards)
	if err != nil {
		return false, fmt.Errorf("ecstore: verify shard set: %w", err)
	}
	return ok, nil
}

func (c *rsCodec) Reconstruct(shards [][]byte) error {
	if err := c.enc.Reconstruct(shards); err != nil {
		return fmt.Errorf("ecstore: reconstruct missing shards: %w", err)
	}
	return nil
}

func (c *rsCodec) Join(shards [][]byte, dataSize int) ([]byte, error) {
	var buf bytes.Buffer
	if err := c.enc.Join(&buf, shards, dataSize); err != nil {
		return nil, fmt.Errorf("ecstore: join %d shards into %d bytes: %w", len(shards), dataSize, err)
	}
	return buf.Bytes(), nil
}
