package benchs

import (
	"sync"
	"sync/atomic"
)

// Test data structure
type testObject struct {
	ID   int
	Data string
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

// RWMutex-protected ring buffer implementation
type RWMutexRingBufferPool struct {
	ringBuffer *RingBuffer[*testObject]
	mu         sync.RWMutex
	allocator  func() *testObject
	cleaner    func(*testObject)
}

func NewRWMutexRingBufferPool(capacity int, allocator func() *testObject, cleaner func(*testObject)) *RWMutexRingBufferPool {
	pool := &RWMutexRingBufferPool{
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

func (p *RWMutexRingBufferPool) Get() *testObject {
	p.mu.Lock()
	defer p.mu.Unlock()

	if obj, ok := p.ringBuffer.Pop(); ok {
		return obj
	}
	return p.allocator()
}

func (p *RWMutexRingBufferPool) Put(obj *testObject) {
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
	if p.cleaner != nil {
		p.cleaner(obj)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.objects) < cap(p.objects) {
		p.objects = append(p.objects, obj)
	}
}

// RWMutex-based implementation for comparison
type RWMutexBasedPool struct {
	objects   []*testObject
	mu        sync.RWMutex
	allocator func() *testObject
	cleaner   func(*testObject)
}

func NewRWMutexBasedPool(capacity int, allocator func() *testObject, cleaner func(*testObject)) *RWMutexBasedPool {
	pool := &RWMutexBasedPool{
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

func (p *RWMutexBasedPool) Get() *testObject {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.objects) == 0 {
		return p.allocator()
	}

	obj := p.objects[len(p.objects)-1]
	p.objects = p.objects[:len(p.objects)-1]
	return obj
}

func (p *RWMutexBasedPool) Put(obj *testObject) {
	if p.cleaner != nil {
		p.cleaner(obj)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.objects) < cap(p.objects) {
		p.objects = append(p.objects, obj)
	}
}

// Atomic-based implementation using atomic.Value
type AtomicBasedPool struct {
	objects   []*testObject
	index     int64
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

	return pool
}

func (p *AtomicBasedPool) Get() *testObject {
	// Try to get an object using atomic operations
	for {
		idx := atomic.LoadInt64(&p.index)
		if idx <= 0 {
			return p.allocator()
		}

		newIdx := idx - 1
		if atomic.CompareAndSwapInt64(&p.index, idx, newIdx) {
			return p.objects[newIdx]
		}
	}
}

func (p *AtomicBasedPool) Put(obj *testObject) {
	if p.cleaner != nil {
		p.cleaner(obj)
	}

	// Try to put the object back
	for {
		idx := atomic.LoadInt64(&p.index)
		if idx >= p.capacity {
			return // Pool is full
		}

		newIdx := idx + 1
		if atomic.CompareAndSwapInt64(&p.index, idx, newIdx) {
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
	if p.cleaner != nil {
		p.cleaner(obj)
	}

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

// Test functions
var testAllocator = func() *testObject {
	return &testObject{
		ID:   0,
		Data: "test",
	}
}

var testCleaner = func(obj *testObject) {
	obj.ID = 0
	obj.Data = ""
}
