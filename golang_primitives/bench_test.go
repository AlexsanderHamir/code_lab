package benchs

import (
	"runtime/debug"
	"testing"
)

// Benchmark: Mutex-protected ring buffer pool
func Benchmark_MutexPool(b *testing.B) {
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

// Benchmark: RWMutex-protected ring buffer pool
func Benchmark_RWMutexPool(b *testing.B) {
	debug.SetGCPercent(-1)
	b.ReportAllocs()

	pool := NewRWMutexRingBufferPool(1000, testAllocator, testCleaner)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

// Benchmark: Channel-based pool
func Benchmark_ChannelPool(b *testing.B) {
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
func Benchmark_AtomicPool(b *testing.B) {
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
func Benchmark_CondPool(b *testing.B) {
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
