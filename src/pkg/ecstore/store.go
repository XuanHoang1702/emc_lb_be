package ecstore

import (
	"context"
	"errors"
	"fmt"
)

// EcStore stores immutable objects as k data + m parity Reed-Solomon shards
// spread across independent placement nodes, and reconstructs them from any
// k surviving shards.
//
// Write order guarantees crash safety: shards are placed first, the manifest
// last — a torn Put leaves orphaned shards but never a readable half-object.
type EcStore struct {
	codec  Codec
	placer ShardPlacer
}

// NewEcStore wires a codec and a shard placement backend together.
func NewEcStore(codec Codec, placer ShardPlacer) (*EcStore, error) {
	if codec == nil {
		return nil, errors.New("ecstore: codec must not be nil")
	}
	if placer == nil {
		return nil, errors.New("ecstore: placer must not be nil")
	}
	return &EcStore{codec: codec, placer: placer}, nil
}

// manifestIndex is the placement slot reserved for the object manifest.
func (s *EcStore) manifestIndex() int { return s.codec.DataShards() + s.codec.ParityShards() }

// Put encodes data into shards, distributes them, and records the manifest.
func (s *EcStore) Put(ctx context.Context, id string, data []byte) (Manifest, error) {
	if id == "" {
		return Manifest{}, ErrInvalidObjectID
	}
	if len(data) == 0 {
		return Manifest{}, ErrEmptyData
	}

	shards, err := s.codec.Split(data)
	if err != nil {
		return Manifest{}, err
	}

	total := s.manifestIndex()
	for len(shards) < total {
		shards = append(shards, make([]byte, len(shards[0])))
	}

	if err = s.codec.Encode(shards); err != nil {
		return Manifest{}, err
	}

	manifest := Manifest{
		ID:        id,
		Size:      len(data),
		K:         s.codec.DataShards(),
		M:         s.codec.ParityShards(),
		ShardSize: len(shards[0]),
		ShardSums: make([]string, total),
		DataSum:   DataSum(data),
	}
	for i, shard := range shards {
		manifest.ShardSums[i] = ShardSum(shard)
	}

	for i, shard := range shards {
		if err = s.placer.Place(ctx, id, i, shard); err != nil {
			return Manifest{}, fmt.Errorf("ecstore: place shard %d: %w", i, err)
		}
	}

	manifestBytes, err := MarshalManifest(manifest)
	if err != nil {
		return Manifest{}, err
	}
	if err := s.placer.Place(ctx, id, total, manifestBytes); err != nil {
		return Manifest{}, fmt.Errorf("ecstore: place manifest: %w", err)
	}

	return manifest, nil
}

// Get gathers shards, verifies each against the manifest checksums,
// reconstructs any missing or corrupted ones, and returns the original bytes.
func (s *EcStore) Get(ctx context.Context, id string) ([]byte, error) {
	shards, manifest, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}

	data, err := s.codec.Join(shards, manifest.Size)
	if err != nil {
		return nil, err
	}

	if DataSum(data) != manifest.DataSum {
		return nil, ErrChecksumMismatch
	}

	return data, nil
}

// Repair reconstructs missing or corrupted shards and re-places them so the
// cluster returns to full redundancy. It is a no-op on a healthy object.
func (s *EcStore) Repair(ctx context.Context, id string) error {
	shards, _, err := s.load(ctx, id)
	if err != nil {
		return err
	}

	for i, shard := range shards {
		if shard == nil {
			continue
		}
		if err := s.placer.Place(ctx, id, i, shard); err != nil {
			return fmt.Errorf("ecstore: repair shard %d: %w", i, err)
		}
	}

	return nil
}

// Delete removes all shards and the manifest for an object.
func (s *EcStore) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidObjectID
	}
	return s.placer.Delete(ctx, id)
}

// load reads the manifest, gathers every shard behind a checksum gate, and
// reconstructs the missing ones. Shards that are down, absent, or corrupt are
// set to nil — unverified bytes are NEVER handed to Reconstruct.
func (s *EcStore) load(ctx context.Context, id string) ([][]byte, Manifest, error) {
	if id == "" {
		return nil, Manifest{}, ErrInvalidObjectID
	}

	manifestBytes, err := s.placer.Get(ctx, id, s.manifestIndex())
	if err != nil {
		if errors.Is(err, ErrNodeDown) || errors.Is(err, ErrShardNotFound) {
			return nil, Manifest{}, ErrObjectNotFound
		}
		return nil, Manifest{}, fmt.Errorf("ecstore: read manifest: %w", err)
	}

	manifest, err := UnmarshalManifest(manifestBytes)
	if err != nil {
		return nil, Manifest{}, err
	}

	shards, err := s.gatherAndReconstruct(ctx, manifest)
	if err != nil {
		return nil, Manifest{}, err
	}

	return shards, manifest, nil
}

// gatherAndReconstruct collects valid shards for an object; when fewer than
// the full set survives but at least k remain, it rebuilds the rest and
// re-verifies every shard against the manifest checksums.
func (s *EcStore) gatherAndReconstruct(ctx context.Context, manifest Manifest) ([][]byte, error) {
	total := manifest.K + manifest.M
	shards := make([][]byte, total)
	valid := 0

	for i := range shards {
		shard, err := s.placer.Get(ctx, manifest.ID, i)
		if err != nil || !VerifyShard(manifest.ShardSums[i], shard) {
			continue
		}
		shards[i] = shard
		valid++
	}

	if valid < manifest.K {
		return nil, fmt.Errorf("%w: have %d valid of %d needed", ErrInsufficientShards, valid, manifest.K)
	}

	if valid == total {
		return shards, nil
	}

	if err := s.codec.Reconstruct(shards); err != nil {
		return nil, err
	}
	if ok, err := s.codec.Verify(shards); err != nil || !ok {
		return nil, ErrChecksumMismatch
	}
	for i := range shards {
		if !VerifyShard(manifest.ShardSums[i], shards[i]) {
			return nil, fmt.Errorf("%w: reconstructed shard %d", ErrChecksumMismatch, i)
		}
	}

	return shards, nil
}
