# Go Primitives Performance Benchmark

This experiment benchmarks the performance characteristics of different Go synchronization primitives when implementing object pools. The goal is to identify which synchronization approach provides the best performance characteristics across various metrics, including `Time/op` and `Throughput`.

## Results

Benchmarks are grouped by **sharded**, **non-sharded**, and **atomic/channel/mutex** pools for clarity. The table also includes throughput differences relative to the previous benchmark.

| Benchmark                                 | Time/op      | Throughput (ops/sec) | Slower by          | % Time Diff | Throughput Diff |
| ----------------------------------------- | ------------ | -------------------- | ------------------ | ----------- | --------------- |
| **Benchmark_ShardedAtomicBasedPool-8**    | **3.699 ns** | **315.9M**           | **0 ns** (fastest) | **0%**      | **0%**          |
| Benchmark_ShardedMutexRingBufferPool-8    | 6.984 ns     | 161.0M               | +3.285 ns          | +88.8%      | -49.0%          |
| Benchmark_ShardedRingBufferCondPool-8     | 12.15 ns     | 100.0M               | +5.166 ns          | +74.0%      | -37.9%          |
| Benchmark_ShardedGoroutineIDPool-8        | 19.21 ns     | 106.4M               | +7.06 ns           | +58.1%      | +6.4%           |
| Benchmark_ShardedChannelBasedPool-8       | 24.94 ns     | 64.7M                | +5.73 ns           | +29.8%      | -39.2%          |
| Benchmark_ChannelBasedPool-8              | 48.94 ns     | 25.2M                | +23.99 ns          | +96.1%      | -61.1%          |
| Benchmark_MutexRingBufferPool-8           | 215.6 ns     | 4.96M                | +0.6 ns            | +0.3%       | -80.3%          |
| Benchmark_ProcPinnedMutexRingBufferPool-8 | 215.9 ns     | 5.64M                | +0.3 ns            | +0.1%       | +13.7%          |
| Benchmark_RingBufferCondPool-8            | 215.0 ns     | 5.51M                | -0.9 ns            | -0.4%       | -2.3%           |
| Benchmark_AtomicBasedPool-8               | 338.6 ns     | 3.77M                | +123.0 ns          | +57.1%      | -31.6%          |

## Optimizations

### 1. Pool Sharding

- **What**: Multiple smaller pools (shards) based on logical processors.
- **How**: Each shard operates independently, reducing contention.
- **Benefits**: Eliminates global locks, improves cache locality, scales with CPU cores.

### 2. Runtime Proc Pinning (`runtime.procPin`)

- **What**: Pins a goroutine to a specific logical processor.
- **How**: Returns P's ID used as a deterministic shard index.
- **Benefits**: Consistent shard assignment, predictable access patterns.

### 3. Shard Index Caching

- **What**: Store shard index in object to avoid recalculation.
- **How**: Cache during `Get()`, reuse during `Put()`.
- **Benefits**: Eliminates repeated `runtime.procPin()` calls, saving cycles.

## Performance Impact

Applying these strategies transforms `AtomicBasedPool` from 338.6 ns/op to `ShardedAtomicBasedPool` at 3.699 ns/op — an **\~91.5× speedup**.

- Pool sharding alone drastically reduces contention.
- Runtime proc pinning stabilizes shard access patterns.
- Shard index caching eliminates repeated calculations.

Combined, these optimizations demonstrate the multiplicative effect of applying thoughtful concurrency strategies.

## Conclusions

Different synchronization mechanisms can offer substantial improvements, but stopping at the first apparent win often leaves significant performance on the table.

For example:

- Sharding a `MutexRingBufferPool` reduces `Time/op` from 215.6 ns to 6.984 ns.
- Further optimizations like shard index caching and runtime proc pinning provide additional gains, pushing `ShardedAtomicBasedPool` to 3.699 ns/op.

**Key takeaway:** synchronization primitives should not be used “raw.” By combining multiple strategies — sharding, caching, and processor pinning — performance can improve by nearly two orders of magnitude. Thoughtful application of primitives and runtime insights unlocks dramatic speedups in concurrent Go programs.

## Notes

1. Even without sharding, **pinning a goroutine to a specific processor** can improve throughput. For example:

| Benchmark                                 | Time/op  | Throughput | Slower by | % Time Diff | Throughput Diff |
| ----------------------------------------- | -------- | ---------- | --------- | ----------- | --------------- |
| Benchmark_MutexRingBufferPool-8           | 215.6 ns | 4.96M      | +0.6 ns   | +0.3%       | NaN             |
| Benchmark_ProcPinnedMutexRingBufferPool-8 | 215.9 ns | 5.64M      | +0.3 ns   | +0.1%       | +13.7%          |

- **Observation**: The per-operation time (`Time/op`) remains almost the same, but throughput increases by **\~13%**.
- This shows that small runtime-level optimizations, like `procPin`, can reduce contention and improve aggregate performance even without changing the pool design. It also illustrates that, even if you’re applying other advanced optimizations, taking a step back to consider **how you use your synchronization mechanisms** can yield additional significant performance gains.
