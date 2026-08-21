package ecstore_test

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"testing"

	"emc_lb/src/pkg/ecstore"
)

func newTestStore(t *testing.T, k, m int) (*ecstore.EcStore, *ecstore.InMemoryCluster) {
	t.Helper()

	codec, err := ecstore.NewRSCoder(k, m)
	if err != nil {
		t.Fatalf("NewRSCoder(%d,%d): %v", k, m, err)
	}

	cluster := ecstore.NewInMemoryCluster(k + m)
	store, err := ecstore.NewEcStore(codec, cluster)
	if err != nil {
		t.Fatalf("NewEcStore: %v", err)
	}

	return store, cluster
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return data
}

// TestXORParityIdentity documents the simplest erasure code by hand:
// with one parity shard C = A XOR B, any one of the three rebuilds another.
// Reed-Solomon generalizes this identity to k-of-n over GF(2^8)/GF(2^16).
func TestXORParityIdentity(t *testing.T) {
	a := randomBytes(t, 32)
	b := randomBytes(t, 32)

	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] ^ b[i]
	}

	rebuiltA := make([]byte, len(a))
	rebuiltB := make([]byte, len(b))
	for i := range c {
		rebuiltA[i] = b[i] ^ c[i]
		rebuiltB[i] = a[i] ^ c[i]
	}

	if !bytes.Equal(rebuiltA, a) {
		t.Fatal("A = B XOR C failed")
	}
	if !bytes.Equal(rebuiltB, b) {
		t.Fatal("B = A XOR C failed")
	}
}

func TestRoundTrip(t *testing.T) {
	sizes := []int{1, 64, 399, 400, 401, 1 << 20} // around k*100 boundary and 1 MiB

	for _, size := range sizes {
		store, _ := newTestStore(t, 4, 2)
		data := randomBytes(t, size)

		if _, err := store.Put(context.Background(), "obj", data); err != nil {
			t.Fatalf("Put(size=%d): %v", size, err)
		}

		got, err := store.Get(context.Background(), "obj")
		if err != nil {
			t.Fatalf("Get(size=%d): %v", size, err)
		}
		if !bytes.Equal(got, data) {
			t.Fatalf("round trip mismatch at size=%d: got %d bytes", size, len(got))
		}
	}
}

func TestLoseOneShard(t *testing.T) {
	for lost := 0; lost < 6; lost++ {
		store, cluster := newTestStore(t, 4, 2)
		data := randomBytes(t, 1000)

		if _, err := store.Put(context.Background(), "obj", data); err != nil {
			t.Fatalf("Put: %v", err)
		}

		cluster.Node(lost).Fail()

		got, err := store.Get(context.Background(), "obj")
		if err != nil {
			t.Fatalf("Get after losing shard %d: %v", lost, err)
		}
		if !bytes.Equal(got, data) {
			t.Fatalf("mismatch after losing shard %d", lost)
		}
	}
}

func TestLoseTwoShards(t *testing.T) {
	for a := 0; a < 6; a++ {
		for b := a + 1; b < 6; b++ {
			store, cluster := newTestStore(t, 4, 2)
			data := randomBytes(t, 1000)

			if _, err := store.Put(context.Background(), "obj", data); err != nil {
				t.Fatalf("Put: %v", err)
			}

			cluster.Node(a).Fail()
			cluster.Node(b).Fail()

			got, err := store.Get(context.Background(), "obj")
			if err != nil {
				t.Fatalf("Get after losing shards %d+%d: %v", a, b, err)
			}
			if !bytes.Equal(got, data) {
				t.Fatalf("mismatch after losing shards %d+%d", a, b)
			}
		}
	}
}

func TestLoseThreeShardsFails(t *testing.T) {
	triples := [][3]int{{0, 1, 2}, {0, 4, 5}, {2, 3, 5}}

	for _, triple := range triples {
		store, cluster := newTestStore(t, 4, 2)

		if _, err := store.Put(context.Background(), "obj", randomBytes(t, 1000)); err != nil {
			t.Fatalf("Put: %v", err)
		}

		for _, idx := range triple {
			cluster.Node(idx).Fail()
		}

		if _, err := store.Get(context.Background(), "obj"); !errors.Is(err, ecstore.ErrInsufficientShards) {
			t.Fatalf("losing shards %v: want ErrInsufficientShards, got %v", triple, err)
		}
	}
}

func TestCorruptShardDetectedAndRecovered(t *testing.T) {
	t.Run("single corruption", func(t *testing.T) {
		store, cluster := newTestStore(t, 4, 2)
		data := randomBytes(t, 1000)

		if _, err := store.Put(context.Background(), "obj", data); err != nil {
			t.Fatalf("Put: %v", err)
		}

		if err := cluster.Node(1).Corrupt("obj", 1); err != nil {
			t.Fatalf("Corrupt: %v", err)
		}

		got, err := store.Get(context.Background(), "obj")
		if err != nil {
			t.Fatalf("Get after corruption: %v", err)
		}
		if !bytes.Equal(got, data) {
			t.Fatal("mismatch after corrupt-shard recovery")
		}
	})

	t.Run("corruption plus loss still recovers", func(t *testing.T) {
		store, cluster := newTestStore(t, 4, 2)
		data := randomBytes(t, 1000)

		if _, err := store.Put(context.Background(), "obj", data); err != nil {
			t.Fatalf("Put: %v", err)
		}

		if err := cluster.Node(1).Corrupt("obj", 1); err != nil {
			t.Fatalf("Corrupt: %v", err)
		}
		cluster.Node(4).Fail()

		got, err := store.Get(context.Background(), "obj")
		if err != nil {
			t.Fatalf("Get after corruption+loss: %v", err)
		}
		if !bytes.Equal(got, data) {
			t.Fatal("mismatch after corruption+loss recovery")
		}
	})
}

func TestRepairRestoresCluster(t *testing.T) {
	store, cluster := newTestStore(t, 4, 2)
	data := randomBytes(t, 1000)

	if _, err := store.Put(context.Background(), "obj", data); err != nil {
		t.Fatalf("Put: %v", err)
	}

	cluster.Node(0).Fail()
	cluster.Node(5).Fail()

	if err := store.Repair(context.Background(), "obj"); err != nil {
		t.Fatalf("Repair: %v", err)
	}

	cluster.Node(0).Recover()
	cluster.Node(5).Recover()

	got, err := store.Get(context.Background(), "obj")
	if err != nil {
		t.Fatalf("Get after repair: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("mismatch after repair")
	}
}

func TestTamperedManifestFails(t *testing.T) {
	store, cluster := newTestStore(t, 4, 2)

	if _, err := store.Put(context.Background(), "obj", randomBytes(t, 1000)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Overwrite the manifest with structurally valid JSON carrying a wrong sum.
	forged := `{"id":"obj","size":1000,"k":4,"m":2,"shard_size":250,
		"shard_sums":["bad","bad","bad","bad","bad","bad"],
		"data_sum":"bad"}`
	if err := cluster.MetaNode().Place(context.Background(), "obj", 6, []byte(forged)); err != nil {
		t.Fatalf("forge manifest: %v", err)
	}

	if _, err := store.Get(context.Background(), "obj"); !errors.Is(err, ecstore.ErrInsufficientShards) {
		t.Fatalf("want ErrInsufficientShards from forged sums, got %v", err)
	}
}

func TestInputValidation(t *testing.T) {
	store, _ := newTestStore(t, 4, 2)
	ctx := context.Background()

	if _, err := store.Put(ctx, "", []byte{1}); !errors.Is(err, ecstore.ErrInvalidObjectID) {
		t.Fatalf("empty id: want ErrInvalidObjectID, got %v", err)
	}
	if _, err := store.Put(ctx, "obj", nil); !errors.Is(err, ecstore.ErrEmptyData) {
		t.Fatalf("empty data: want ErrEmptyData, got %v", err)
	}
	if _, err := store.Get(ctx, "missing"); !errors.Is(err, ecstore.ErrObjectNotFound) {
		t.Fatalf("missing object: want ErrObjectNotFound, got %v", err)
	}
}

func TestDeleteRemovesObject(t *testing.T) {
	store, _ := newTestStore(t, 4, 2)
	ctx := context.Background()

	if _, err := store.Put(ctx, "obj", randomBytes(t, 100)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := store.Delete(ctx, "obj"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get(ctx, "obj"); !errors.Is(err, ecstore.ErrObjectNotFound) {
		t.Fatalf("Get after Delete: want ErrObjectNotFound, got %v", err)
	}
}
