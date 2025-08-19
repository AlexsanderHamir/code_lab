package benchs

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Test data structure
type testObject struct {
	ID         int
	Data       string
	shardIndex int
}

// Ring buffer data structure
type RingBuffer[T any] struct {
	buffer   []T
	head     int
	tail     int
	size     int
	capacity int
}

func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buffer:   make([]T, capacity),
		capacity: capacity,
		size:     0,
		head:     0,
		tail:     0,
	}
}

func (rb *RingBuffer[T]) Push(item T) bool {
	if rb.size >= rb.capacity {
		return false // Buffer is full
	}

	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.capacity
	rb.size++
	return true
}

func (rb *RingBuffer[T]) Pop() (T, bool) {
	var zero T
	if rb.size <= 0 {
		return zero, false // Buffer is empty
	}

	item := rb.buffer[rb.head]
	rb.head = (rb.head + 1) % rb.capacity
	rb.size--
	return item, true
}

func (rb *RingBuffer[T]) IsEmpty() bool {
	return rb.size == 0
}

func (rb *RingBuffer[T]) IsFull() bool {
	return rb.size == rb.capacity
}

func (rb *RingBuffer[T]) Size() int {
	return rb.size
}

// Mutex-protected ring buffer implementation
type MutexRingBufferPool struct {
	ringBuffer *RingBuffer[*testObject]
	mu         sync.Mutex
	allocator  func() *testObject
	cleaner    func(*testObject)
}

func NewMutexRingBufferPool(capacity int, allocator func() *testObject, cleaner func(*testObject)) *MutexRingBufferPool {
	pool := &MutexRingBufferPool{
		ringBuffer: NewRingBuffer[*testObject](capacity),
		allocator:  allocator,
		cleaner:    cleaner,
	}

	// Pre-populate the pool
	for range capacity {
		pool.ringBuffer.Push(allocator())
	}

	return pool
}

func (p *MutexRingBufferPool) Get() *testObject {
	p.mu.Lock()
	defer p.mu.Unlock()

	if obj, ok := p.ringBuffer.Pop(); ok {
		return obj
	}

	return p.allocator()
}

func (p *MutexRingBufferPool) Put(obj *testObject) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ringBuffer.Push(obj)
}

// Channel-based implementation for comparison
type ChannelBasedPool struct {
	objects   chan *testObject
	allocator func() *testObject
	cleaner   func(*testObject)
}

func NewChannelBasedPool(capacity int, allocator func() *testObject, cleaner func(*testObject)) *ChannelBasedPool {
	pool := &ChannelBasedPool{
		objects:   make(chan *testObject, capacity),
		allocator: allocator,
		cleaner:   cleaner,
	}

	// Pre-populate the pool
	for range capacity {
		pool.objects <- allocator()
	}

	return pool
}

func (p *ChannelBasedPool) Get() *testObject {
	select {
	case obj := <-p.objects:
		return obj
	default:
		return p.allocator()
	}
}

func (p *ChannelBasedPool) Put(obj *testObject) {
	p.cleaner(obj)

	// Select isn't used because we don't want to drop objects.
	p.objects <- obj
}

// Mutex-protected slice-based implementation for comparison
type MutexBasedPool struct {
	objects   []*testObject
	mu        sync.Mutex
	allocator func() *testObject
	cleaner   func(*testObject)
}

func NewMutexBasedPool(capacity int, allocator func() *testObject, cleaner func(*testObject)) *MutexBasedPool {
	pool := &MutexBasedPool{
		objects:   make([]*testObject, 0, capacity),
		allocator: allocator,
		cleaner:   cleaner,
	}

	// Pre-populate the pool
	for range capacity {
		pool.objects = append(pool.objects, allocator())
	}

	return pool
}

func (p *MutexBasedPool) Get() *testObject {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.objects) == 0 {
		return p.allocator()
	}

	// Get last element (more efficient than shifting)
	obj := p.objects[len(p.objects)-1]
	p.objects = p.objects[:len(p.objects)-1]
	return obj
}

func (p *MutexBasedPool) Put(obj *testObject) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.objects) < cap(p.objects) {
		p.objects = append(p.objects, obj)
	}
}

// Atomic-based implementation using atomic.Int64
type AtomicBasedPool struct {
	objects   []*testObject
	index     atomic.Int64
	capacity  int64
	allocator func() *testObject
	cleaner   func(*testObject)
}

func NewAtomicBasedPool(capacity int, allocator func() *testObject, cleaner func(*testObject)) *AtomicBasedPool {
	pool := &AtomicBasedPool{
		objects:   make([]*testObject, capacity),
		capacity:  int64(capacity),
		allocator: allocator,
		cleaner:   cleaner,
	}

	// Pre-populate the pool
	for i := range capacity {
		pool.objects[i] = allocator()
	}
	// initialize index to full
	pool.index.Store(int64(capacity))

	return pool
}

func (p *AtomicBasedPool) Get() *testObject {
	for {
		idx := p.index.Load()
		if idx <= 0 {
			return p.allocator()
		}
		if p.index.CompareAndSwap(idx, idx-1) {
			return p.objects[idx-1]
		}
	}
}

func (p *AtomicBasedPool) Put(obj *testObject) {
	for {
		idx := p.index.Load()
		if idx >= p.capacity {
			return
		}
		if p.index.CompareAndSwap(idx, idx+1) {
			p.objects[idx] = obj
			return
		}
	}
}

// Condition variable-based implementation for comparison
type CondBasedPool struct {
	objects   []*testObject
	mu        sync.Mutex
	cond      *sync.Cond
	allocator func() *testObject
	cleaner   func(*testObject)
}

func NewCondBasedPool(capacity int, allocator func() *testObject, cleaner func(*testObject)) *CondBasedPool {
	pool := &CondBasedPool{
		objects:   make([]*testObject, 0, capacity),
		allocator: allocator,
		cleaner:   cleaner,
	}
	pool.cond = sync.NewCond(&pool.mu)

	// Pre-populate the pool
	for range capacity {
		pool.objects = append(pool.objects, allocator())
	}

	return pool
}

func (p *CondBasedPool) Get() *testObject {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Wait until there's an object available
	for len(p.objects) == 0 {
		p.cond.Wait()
	}

	// Get last element
	obj := p.objects[len(p.objects)-1]
	p.objects = p.objects[:len(p.objects)-1]
	return obj
}

func (p *CondBasedPool) Put(obj *testObject) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.objects) < cap(p.objects) {
		p.objects = append(p.objects, obj)
		p.cond.Signal() // Wake up one waiting goroutine
	}
}

// Ring buffer with condition variables implementation
type RingBufferCondPool struct {
	ringBuffer *RingBuffer[*testObject]
	mu         sync.Mutex
	cond       *sync.Cond
	allocator  func() *testObject
	cleaner    func(*testObject)
}

func NewRingBufferCondPool(capacity int, allocator func() *testObject, cleaner func(*testObject)) *RingBufferCondPool {
	pool := &RingBufferCondPool{
		ringBuffer: NewRingBuffer[*testObject](capacity),
		allocator:  allocator,
		cleaner:    cleaner,
	}
	pool.cond = sync.NewCond(&pool.mu)

	// Pre-populate the pool
	for range capacity {
		pool.ringBuffer.Push(allocator())
	}

	return pool
}

func (p *RingBufferCondPool) Get() *testObject {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Wait until there's an object available
	for p.ringBuffer.IsEmpty() {
		p.cond.Wait()
	}

	obj, _ := p.ringBuffer.Pop()
	return obj
}

func (p *RingBufferCondPool) Put(obj *testObject) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.ringBuffer.IsFull() {
		p.ringBuffer.Push(obj)
		p.cond.Signal() // Wake up one waiting goroutine
	}
}

// Sharded pool implementations
// These implementations use multiple sub-pools to reduce contention

// Sharded Mutex Ring Buffer Pool
type ShardedMutexRingBufferPool struct {
	shards []*MutexRingBufferPool
}

func NewShardedMutexRingBufferPool(capacity int, numShards int, allocator func() *testObject, cleaner func(*testObject)) *ShardedMutexRingBufferPool {
	if numShards <= 0 {
		numShards = runtime.GOMAXPROCS(0)
	}

	shards := make([]*MutexRingBufferPool, numShards)
	shardCapacity := max(capacity/numShards, 1)

	for i := range numShards {
		shards[i] = NewMutexRingBufferPool(shardCapacity, allocator, cleaner)
	}

	return &ShardedMutexRingBufferPool{
		shards: shards,
	}
}

func (p *ShardedMutexRingBufferPool) Get() *testObject {
	shardIndex := runtimeProcPin()
	shard := p.shards[shardIndex]
	runtimeProcUnpin()

	obj := shard.Get()
	obj.shardIndex = shardIndex
	return obj
}

func (p *ShardedMutexRingBufferPool) Put(obj *testObject) {
	shard := p.shards[obj.shardIndex]
	shard.Put(obj)
}

// Sharded Channel Based Pool
type ShardedChannelBasedPool struct {
	shards []*ChannelBasedPool
}

func NewShardedChannelBasedPool(capacity int, numShards int, allocator func() *testObject, cleaner func(*testObject)) *ShardedChannelBasedPool {
	if numShards <= 0 {
		numShards = runtime.GOMAXPROCS(0)
	}

	shards := make([]*ChannelBasedPool, numShards)
	shardCapacity := capacity / numShards
	if shardCapacity < 1 {
		shardCapacity = 1
	}

	for i := 0; i < numShards; i++ {
		shards[i] = NewChannelBasedPool(shardCapacity, allocator, cleaner)
	}

	return &ShardedChannelBasedPool{
		shards: shards,
	}
}

func (p *ShardedChannelBasedPool) Get() *testObject {
	shardIndex := runtimeProcPin()
	shard := p.shards[shardIndex]
	runtimeProcUnpin()

	obj := shard.Get()
	obj.shardIndex = shardIndex
	return obj
}

func (p *ShardedChannelBasedPool) Put(obj *testObject) {
	// Return the object to the same shard it came from
	shardIndex := obj.shardIndex
	if shardIndex >= 0 && shardIndex < len(p.shards) {
		p.shards[shardIndex].Put(obj)
	}
}

// Sharded Atomic Based Pool
type ShardedAtomicBasedPool struct {
	shards []*AtomicBasedPool
}

func NewShardedAtomicBasedPool(capacity int, numShards int, allocator func() *testObject, cleaner func(*testObject)) *ShardedAtomicBasedPool {
	if numShards <= 0 {
		numShards = runtime.GOMAXPROCS(0)
	}

	shards := make([]*AtomicBasedPool, numShards)
	shardCapacity := max(capacity/numShards, 1)

	for i := 0; i < numShards; i++ {
		shards[i] = NewAtomicBasedPool(shardCapacity, allocator, cleaner)
	}

	return &ShardedAtomicBasedPool{
		shards: shards,
	}
}

func (p *ShardedAtomicBasedPool) Get() *testObject {
	shardIndex := runtimeProcPin()
	shard := p.shards[shardIndex]
	runtimeProcUnpin()

	obj := shard.Get()
	obj.shardIndex = shardIndex
	return obj
}

func (p *ShardedAtomicBasedPool) Put(obj *testObject) {
	shard := p.shards[obj.shardIndex]
	shard.Put(obj)
}

// Sharded Condition Based Pool
type ShardedCondBasedPool struct {
	shards []*RingBufferCondPool
}

func NewShardedCondBasedPool(capacity int, numShards int, allocator func() *testObject, cleaner func(*testObject)) *ShardedCondBasedPool {
	if numShards <= 0 {
		numShards = runtime.GOMAXPROCS(0)
	}

	shards := make([]*RingBufferCondPool, numShards)
	shardCapacity := capacity / numShards
	if shardCapacity < 1 {
		shardCapacity = 1
	}

	for i := 0; i < numShards; i++ {
		shards[i] = NewRingBufferCondPool(shardCapacity, allocator, cleaner)
	}

	return &ShardedCondBasedPool{
		shards: shards,
	}
}

func (p *ShardedCondBasedPool) Get() *testObject {
	shardIndex := runtimeProcPin()
	shard := p.shards[shardIndex]
	runtimeProcUnpin()

	obj := shard.Get()
	obj.shardIndex = shardIndex
	return obj
}

func (p *ShardedCondBasedPool) Put(obj *testObject) {
	shard := p.shards[obj.shardIndex]
	shard.Put(obj)
}

// Sharded Ring Buffer Condition Pool
type ShardedRingBufferCondPool struct {
	shards []*RingBufferCondPool
}

func NewShardedRingBufferCondPool(capacity int, numShards int, allocator func() *testObject, cleaner func(*testObject)) *ShardedRingBufferCondPool {
	if numShards <= 0 {
		numShards = runtime.GOMAXPROCS(0)
	}

	shards := make([]*RingBufferCondPool, numShards)
	shardCapacity := capacity / numShards
	if shardCapacity < 1 {
		shardCapacity = 1
	}

	for i := 0; i < numShards; i++ {
		shards[i] = NewRingBufferCondPool(shardCapacity, allocator, cleaner)
	}

	return &ShardedRingBufferCondPool{
		shards: shards,
	}
}

func (p *ShardedRingBufferCondPool) Get() *testObject {
	shardIndex := runtimeProcPin()
	shard := p.shards[shardIndex]
	runtimeProcUnpin()

	obj := shard.Get()
	obj.shardIndex = shardIndex
	return obj
}

func (p *ShardedRingBufferCondPool) Put(obj *testObject) {
	shard := p.shards[obj.shardIndex]
	shard.Put(obj)
}

// Sharded Pool using Goroutine ID (no proc pinning, no shard index storage)
type ShardedGoroutineIDPool struct {
	shards    []*MutexRingBufferPool
	numShards int
}

func NewShardedGoroutineIDPool(capacity int, numShards int, allocator func() *testObject, cleaner func(*testObject)) *ShardedGoroutineIDPool {
	if numShards <= 0 {
		numShards = runtime.GOMAXPROCS(0)
	}

	shards := make([]*MutexRingBufferPool, numShards)
	shardCapacity := max(capacity/numShards, 1)

	for i := range numShards {
		shards[i] = NewMutexRingBufferPool(shardCapacity, allocator, cleaner)
	}

	return &ShardedGoroutineIDPool{
		shards:    shards,
		numShards: numShards,
	}
}

func (p *ShardedGoroutineIDPool) getShard() *MutexRingBufferPool {
	var dummy byte
	addr := uintptr(unsafe.Pointer(&dummy))
	id := int(addr>>12) & (p.numShards - 1)
	return p.shards[id]
}

func (p *ShardedGoroutineIDPool) Get() *testObject {
	shard := p.getShard()
	return shard.Get()
}

func (p *ShardedGoroutineIDPool) Put(obj *testObject) {
	shard := p.getShard()
	shard.Put(obj)
}

//go:linkname runtimeProcPin runtime.procPin
func runtimeProcPin() int

//go:linkname runtimeProcUnpin runtime.procUnpin
func runtimeProcUnpin()

// Test functions
var testAllocator = func() *testObject {
	return &testObject{
		ID:         0,
		Data:       "test",
		shardIndex: -1, // -1 indicates not yet assigned to a shard
	}
}

var testCleaner = func(obj *testObject) {
	obj.ID = 0
	obj.Data = ""
	obj.shardIndex = -1
}
