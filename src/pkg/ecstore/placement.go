package ecstore

import (
	"context"
	"errors"
	"sync"
)

// ShardPlacer abstracts where shards physically live. Production placements
// would target network hosts or disks; the in-memory implementation below
// exists for testing and failure simulation.
type ShardPlacer interface {
	// Place stores (or overwrites) one shard of an object.
	Place(ctx context.Context, objectID string, shardIndex int, shard []byte) error
	// Get returns one shard; implementations return ErrNodeDown or
	// ErrShardNotFound so the store can treat both as "missing".
	Get(ctx context.Context, objectID string, shardIndex int) ([]byte, error)
	// Delete removes every stored shard and manifest copy for an object.
	Delete(ctx context.Context, objectID string) error
}

// InMemoryNode simulates one independent storage node: a map of objects to
// shards that can be failed or corrupted on demand.
type InMemoryNode struct {
	mu    sync.RWMutex
	down  bool
	store map[string]map[int][]byte
}

func NewInMemoryNode() *InMemoryNode {
	return &InMemoryNode{store: make(map[string]map[int][]byte)}
}

// Fail takes the node down; all reads fail with ErrNodeDown until Recover.
func (n *InMemoryNode) Fail() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.down = true
}

// Recover brings a failed node back online.
func (n *InMemoryNode) Recover() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.down = false
}

// Down reports whether the node is currently failed.
func (n *InMemoryNode) Down() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.down
}

// Corrupt flips the first byte of the given shard, simulating silent bit rot.
func (n *InMemoryNode) Corrupt(objectID string, shardIndex int) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	shards, ok := n.store[objectID]
	if !ok {
		return ErrShardNotFound
	}
	shard, ok := shards[shardIndex]
	if !ok {
		return ErrShardNotFound
	}

	shard[0] ^= 0xFF
	return nil
}

func (n *InMemoryNode) Place(_ context.Context, objectID string, shardIndex int, shard []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	copyShard := append([]byte(nil), shard...)

	shards, ok := n.store[objectID]
	if !ok {
		shards = make(map[int][]byte)
		n.store[objectID] = shards
	}
	shards[shardIndex] = copyShard

	return nil
}

func (n *InMemoryNode) Get(_ context.Context, objectID string, shardIndex int) ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.down {
		return nil, ErrNodeDown
	}

	shards, ok := n.store[objectID]
	if !ok {
		return nil, ErrShardNotFound
	}
	shard, ok := shards[shardIndex]
	if !ok {
		return nil, ErrShardNotFound
	}

	return append([]byte(nil), shard...), nil
}

func (n *InMemoryNode) Delete(_ context.Context, objectID string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.store, objectID)
	return nil
}

// InMemoryCluster routes shard index i to node i. A cluster of n shard nodes
// owns one extra metadata node holding manifests — in production the manifest
// must be replicated and authenticated instead.
type InMemoryCluster struct {
	nodes []*InMemoryNode
}

// NewInMemoryCluster creates shardNodes shard nodes plus one metadata node.
func NewInMemoryCluster(shardNodes int) *InMemoryCluster {
	nodes := make([]*InMemoryNode, shardNodes+1)
	for i := range nodes {
		nodes[i] = NewInMemoryNode()
	}
	return &InMemoryCluster{nodes: nodes}
}

// MetaNode exposes the manifest node (index == shard count).
func (c *InMemoryCluster) MetaNode() *InMemoryNode {
	return c.nodes[len(c.nodes)-1]
}

// Node returns the shard node at the given index.
func (c *InMemoryCluster) Node(i int) *InMemoryNode {
	return c.nodes[i]
}

func (c *InMemoryCluster) Place(ctx context.Context, objectID string, shardIndex int, shard []byte) error {
	return c.nodes[shardIndex].Place(ctx, objectID, shardIndex, shard)
}

func (c *InMemoryCluster) Get(ctx context.Context, objectID string, shardIndex int) ([]byte, error) {
	return c.nodes[shardIndex].Get(ctx, objectID, shardIndex)
}

func (c *InMemoryCluster) Delete(ctx context.Context, objectID string) error {
	var errs []error
	for _, node := range c.nodes {
		if err := node.Delete(ctx, objectID); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
