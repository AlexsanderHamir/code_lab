package benchs

import (
	"runtime/debug"
	"testing"
)

// Benchmark: Mutex-protected ring buffer pool
func Benchmark_MutexRingBufferPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewMutexRingBufferPool(1000, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Channel-based pool
func Benchmark_ChannelBasedPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewChannelBasedPool(1000, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Atomic-based pool
func Benchmark_AtomicBasedPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewAtomicBasedPool(1000, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Ring buffer with condition variables pool
func Benchmark_RingBufferCondPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewRingBufferCondPool(1000, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Sharded Mutex Ring Buffer Pool
func Benchmark_ShardedMutexRingBufferPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewShardedMutexRingBufferPool(1000, 0, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Sharded Channel-based pool
func Benchmark_ShardedChannelBasedPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewShardedChannelBasedPool(1000, 0, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Sharded Atomic-based pool
func Benchmark_ShardedAtomicBasedPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewShardedAtomicBasedPool(1000, 0, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Sharded Ring buffer with condition variables pool
func Benchmark_ShardedRingBufferCondPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewShardedRingBufferCondPool(1000, 0, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Sharded Goroutine ID Pool (no proc pinning, no shard index storage)
func Benchmark_ShardedGoroutineIDPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewShardedGoroutineIDPool(1000, 0, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}
