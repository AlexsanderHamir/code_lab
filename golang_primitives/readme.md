# Go Primitives Performance Benchmark

## Overview

Benchmarking different Go synchronization primitives for object pools to see which performs best.

## Results

### Performance Rankings (ns/op)

| Rank | Implementation        | Performance     | Performance Ratio |
| ---- | --------------------- | --------------- | ----------------- |
| 1    | ChannelBasedPool      | **132.7 ns/op** | 1.0x (baseline)   |
| 2    | RWMutexRingBufferPool | 197.1 ns/op     | 1.5x slower       |
| 3    | RingBufferCond        | 198.4 ns/op     | 1.5x slower       |
| 4    | MutexRingBufferPool   | 198.7 ns/op     | 1.5x slower       |
| 5    | AtomicBasedPool       | 631.6 ns/op     | 4.8x slower       |

## Key Findings

- **Channels are fastest** by a significant margin
- **Mutex-based solutions** perform similarly (~198 ns/op)
- **Atomic operations** are surprisingly slow in this context

## Conclusions

1. **Channels are optimized** for general use and provide best performance
2. **Mutexes give consistent** performance across variants
3. **Atomics need careful** implementation around it to perform well
4. **Use channels by default** unless you have specific needs

## Test Setup

- Go 1.24.3, ARM64 (Apple M1)
- Pool size: 1000 objects
- Parallel execution with GC disabled
- Zero memory allocations per operation

---

_Go's built-in primitives are well-optimized and often the best choice._
