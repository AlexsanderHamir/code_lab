# Go Primitives Performance Benchmark

Benchmarking different Go synchronization primitives for object pools to see which performs best.

## Results

### Performance Rankings (ns/op)

| Rank | Implementation        | Performance     | Performance vs Baseline |
| ---- | --------------------- | --------------- | ----------------------- |
| 1    | ChannelBasedPool      | **132.7 ns/op** | 0% (baseline)           |
| 2    | RWMutexRingBufferPool | 197.1 ns/op     | 48.5% slower            |
| 3    | RingBufferCond        | 198.4 ns/op     | 49.5% slower            |
| 4    | MutexRingBufferPool   | 198.7 ns/op     | 49.7% slower            |
| 5    | AtomicBasedPool       | 631.6 ns/op     | 375.8% slower           |

## Key Findings

- **Channels are fastest** by a significant margin
- **Mutex-based solutions** perform similarly (~49-50% slower than channels)
- **Atomic operations** are surprisingly slow in this context (375.8% slower than channels)

## Conclusions

1. **Channels are optimized** for general use and provide best performance
2. **Mutexes give consistent** performance across variants (unless workload changes)
3. **Atomics need careful** implementation around it to perform well
4. **Use channels by default** unless you have specific needs

## Test Setup

- Go 1.24.3, ARM64 (Apple M1)
- Pool size: 1000 objects
- Parallel execution with GC disabled
- Zero memory allocations per operation

---

_Go's built-in primitives are well-optimized and often the best choice._
