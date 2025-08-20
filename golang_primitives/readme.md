# Go Primitives Performance Benchmark

This experiment benchmarks the performance characteristics of different Go synchronization primitives when implementing object pools. The goal is to identify which synchronization approach provides the best performance characteristics across various metrics, including `Time/op` and `Throughput`.

## Results

Benchmarks are grouped by **sharded**, **non-sharded**, and **atomic/channel/mutex** pools for clarity. The table also includes throughput differences relative to the previous benchmark.

| Benchmark                                                | Time/op      | Throughput (ops/sec) | Slower by          | % Time Diff | Throughput Diff |
| -------------------------------------------------------- | ------------ | -------------------- | ------------------ | ----------- | --------------- |
| **Benchmark_ShardedAtomicBasedPool-8**                   | **4.332 ns** | **328.6M**           | **0 ns** (fastest) | **0%**      | **0%**          |
| Benchmark_ShardedMutexRingBufferPool-8                   | 8.202 ns     | 167.97M              | +3.87 ns           | +89.3%      | -48.9%          |
| Benchmark_ShardedRingBufferCondPool-8                    | 7.796 ns     | 152.5M               | +3.46 ns           | +44.8%      | -9.3%           |
| Benchmark_ShardedGoroutineIDPool-8                       | 24.24 ns     | 101.7M               | +16.44 ns          | +211.0%     | -33.3%          |
| Benchmark_ShardedChannelBasedPool-8                      | 26.05 ns     | 52.6M                | +1.81 ns           | +7.5%       | -48.3%          |
| Benchmark_ChannelBasedPool-8                             | 48.22 ns     | 25.7M                | +22.17 ns          | +85.1%      | -51.1%          |
| Benchmark_MutexRingBufferPoolHighConcurrency-8           | 206.5 ns     | 5.80M                | -9.0 ns            | -4.2%       | +22.9%          |
| Benchmark_MutexRingBufferPool-8                          | 215.5 ns     | 5.06M                | +9.0 ns            | +4.4%       | -12.8%          |
| Benchmark_ProcPinnedMutexRingBufferPool-8                | 210.0 ns     | 5.70M                | -5.5 ns            | -2.6%       | +12.6%          |
| Benchmark_ProcPinnedMutexRingBufferPoolHighConcurrency-8 | 210.1 ns     | 5.75M                | +0.1 ns            | +0.05%      | +0.9%           |
| Benchmark_RingBufferCondPool-8                           | 211.7 ns     | 6.16M                | +1.6 ns            | +0.8%       | +7.1%           |
| Benchmark_AtomicBasedPool-8                              | 349.0 ns     | 3.51M                | +137.3 ns          | +64.8%      | -43.0%          |

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

Applying these strategies transforms `AtomicBasedPool` from 349.0 ns/op to `ShardedAtomicBasedPool` at 4.332 ns/op — an **\~80.5× speedup**.

- Pool sharding alone drastically reduces contention.
- Runtime proc pinning stabilizes shard access patterns.
- Shard index caching eliminates repeated calculations.

Combined, these optimizations demonstrate the multiplicative effect of applying thoughtful concurrency strategies.

## Conclusions

Different synchronization mechanisms can offer substantial improvements, but stopping at the first apparent win often leaves significant performance on the table.

For example:

- Sharding a `MutexRingBufferPool` reduces `Time/op` from 215.5 ns to 8.202 ns.
- Further optimizations like shard index caching and runtime proc pinning provide additional gains, pushing `ShardedAtomicBasedPool` to 4.332 ns/op.

**Key takeaway:** synchronization primitives should not be used “raw.” By combining multiple strategies — sharding, caching, and processor pinning — performance can improve by nearly two orders of magnitude. Thoughtful application of primitives and runtime insights unlocks dramatic speedups in concurrent Go programs.

## Notes

1. Even without sharding, **pinning a goroutine to a specific processor** can slightly improve throughput:

| Benchmark                                 | Time/op  | Throughput | 
| ----------------------------------------- | -------- | ---------- |
| Benchmark_MutexRingBufferPool-8           | 215.5 ns | 4.64M      | 
| Benchmark_ProcPinnedMutexRingBufferPool-8 | 210.0 ns | 5.76M      |

- **Observation**: `Time/op` improves slightly (from 215.5 ns to 210.0 ns), and throughput increases by **\~2.6%**.
- **Takeaway**: Small runtime-level optimizations like `procPin` reduce contention and improve aggregate performance even without changing pool design. This demonstrates that, even if advanced optimizations are already applied, revisiting **how you use synchronization primitives** can yield further gains.

2. **Caveat**: The benefit of `procPin` diminishes under very high concurrency. Once thousands of goroutines are involved, pinning without sharding can actually hurt performance, as illustrated below:

\| Benchmark_MutexRingBufferPoolHighConcurrency-8 | 206.5 ns | 5.80M | -9.0 ns | -4.2% | +22.9% |
\| Benchmark_ProcPinnedMutexRingBufferPoolHighConcurrency-8 | 210.1 ns | 5.75M | +0.1 ns | +0.05% | +0.9% |

- **Observation**: At high concurrency, the proc-pinned variant slightly increases `Time/op` (210.1 ns vs 206.5 ns) and barely improves throughput (\~0.9%).
- **Takeaway**: Without sharding, `procPin` alone is insufficient to scale under extreme concurrency. Combining it with sharding or other contention-reducing strategies is necessary for high parallel workloads.
