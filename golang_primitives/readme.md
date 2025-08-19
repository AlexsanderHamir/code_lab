# Go Primitives Performance Benchmark

This experiment benchmarks the performance characteristics of different Go synchronization primitives when implementing object pools. The goal is to identify which synchronization approach provides the best performance characteristics across various metrics including:

## Results

| Benchmark                              | Time/op      | Slower by          | % Diff from Previous |
| -------------------------------------- | ------------ | ------------------ | -------------------- |
| **Benchmark_ShardedAtomicBasedPool-8** | **4.191 ns** | **0 ns** (fastest) | **0%**               |
| Benchmark_ShardedMutexRingBufferPool-8 | 7.208 ns     | +3.017 ns          | +72.0%               |
| Benchmark_ShardedRingBufferCondPool-8  | 7.653 ns     | +0.445 ns          | +6.2%                |
| Benchmark_ShardedGoroutineIDPool-8     | 19.61 ns     | +11.957 ns         | +156.3%              |
| Benchmark_ShardedChannelBasedPool-8    | 26.34 ns     | +6.73 ns           | +34.3%               |
| Benchmark_ChannelBasedPool-8           | 48.51 ns     | +22.17 ns          | +84.2%               |
| Benchmark_AtomicBasedPool-8            | 342.2 ns     | +293.69 ns         | +605.3%              |
| Benchmark_MutexRingBufferPool-8        | 216.0 ns     | -126.2 ns          | -36.9%               |
| Benchmark_RingBufferCondPool-8         | 215.3 ns     | -0.7 ns            | -0.3%                |

## Optimizations

### 1. Pool Sharding

- **What**: Multiple smaller pools (shards) based on logical processors
- **How**: Each shard operates independently, reducing contention
- **Benefits**: Eliminates global locks, improves cache locality, scales with CPU cores

### 2. Runtime Proc Pinning (`runtime.procPin`)

- **What**: Pins goroutine to specific logical processor
- **How**: Returns P's ID used as deterministic shard index
- **Benefits**: Consistent shard assignment, predictable access

### 3. Shard Index Caching

- **What**: Store shard index in object to avoid recalculation
- **How**: Cache during `Get()`, reuse during `Put()`
- **Benefits**: Eliminates repeated `runtime.procPin()` calls

### Performance Impact

Combination transforms `AtomicBasedPool` from 342.2 ns to `ShardedAtomicBasedPool` at 4.191 ns - **87.7x speedup**.

## Conclusions

Different synchronization mechanisms can offer substantial improvements, but stopping at the first apparent win often leaves significant performance on the table. For example, sharding Benchmark_MutexRingBufferPool reduced the time from 216.0 ns to 19.61 ns, a clear win by reducing contention. Many would consider this sufficient, yet further optimizations, such as shard index caching and runtime proc pinning, can yield nearly a 3× additional improvement. The key takeaway is that synchronization primitives should not be used “raw.” Once you choose a mechanism, you can often pull control back into your code or into runtime cheaper areas to mitigate its cost. Combining multiple strategies, as with ShardedAtomicBasedPool achieving an 87.7× speedup, shows the dramatic gains possible when primitives are thoughtfully applied.
