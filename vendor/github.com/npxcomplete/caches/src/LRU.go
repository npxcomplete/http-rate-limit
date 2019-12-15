package caches

type lruNode struct {
	next *lruNode
	prev *lruNode

	key   Key
	value Value
}

type lruCache struct {
	// add to head
	head *lruNode

	// fast access
	store map[Key]*lruNode

	// full if stack == nil
	stack *lruNode

	// v_node pool, to prevent GC churn
	nodes []lruNode
}

func NewLRUCache(capacity int) *lruCache {
	memoryPool := make([]lruNode, capacity)
	for i := 0; i < capacity-1; i++ {
		memoryPool[i].next = &memoryPool[i+1]
	}

	// simplified nil checks
	dummy := &lruNode{}

	dummy.next = dummy
	dummy.prev = dummy

	return &lruCache{
		store: make(map[Key]*lruNode),
		stack: &memoryPool[0],
		head:  dummy,
		nodes: memoryPool,
	}
}

func (lru *lruCache) Put(key Key, value Value) (evictedValue Value) {
	var node *lruNode = nil
	var ok bool

	if node, ok = lru.store[key]; ok {
		// key already present, evict present lruNode to replace
	} else if lru.stack == nil {
		// cache full, evict the tail
		node = lru.head.prev
	}

	// do eviction
	if node != nil {
		node.prev.next = node.next
		node.next.prev = node.prev

		node.prev = nil
		node.next = nil

		// stack push
		node.next = lru.stack
		lru.stack = node

		evictedValue = node.value
		delete(lru.store, node.key)
	}

	// stack must be non-empty by now
	node = lru.stack
	lru.stack = lru.stack.next

	node.key = key
	node.value = value

	// insert at head
	node.next = lru.head.next
	node.next.prev = node
	node.prev = lru.head
	lru.head.next = node

	lru.store[key] = node

	return
}

func (lru *lruCache) Get(key Key) (value Value, err error) {
	var node *lruNode
	var ok bool
	if node, ok = lru.store[key]; !ok {
		err = MissingValueError
		return
	}

	value = node.value
	// Put implicitly resets eviction priority
	lru.Put(key, value)
	return
}
