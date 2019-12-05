package caches

type iKeyTypeValueType_NODE struct {
	next *iKeyTypeValueType_NODE
	prev *iKeyTypeValueType_NODE

	key   KeyType
	value ValueType
}

type iKeyTypeValueTypeLRU struct {
	// add to head
	head *iKeyTypeValueType_NODE

	// fast access
	store map[KeyType]*iKeyTypeValueType_NODE

	// full if stack == nil
	stack *iKeyTypeValueType_NODE

	// v_node pool, to prevent GC churn
	nodes []iKeyTypeValueType_NODE
}

func NewKeyTypeValueTypeLRU(capacity int) *iKeyTypeValueTypeLRU {
	memoryPool := make([]iKeyTypeValueType_NODE, capacity)
	for i := 0; i < capacity-1; i++ {
		memoryPool[i].next = &memoryPool[i+1]
	}

	// simplified nil checks
	dummy := &iKeyTypeValueType_NODE{}

	dummy.next = dummy
	dummy.prev = dummy

	return &iKeyTypeValueTypeLRU{
		store: make(map[KeyType]*iKeyTypeValueType_NODE),
		stack: &memoryPool[0],
		head:  dummy,
		nodes: memoryPool,
	}
}

func (lru *iKeyTypeValueTypeLRU) Put(key KeyType, value ValueType) (evictedValue ValueType) {
	var node *iKeyTypeValueType_NODE = nil
	var ok bool

	if node, ok = lru.store[key]; ok {
		// key already present, evict present node to replace
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

func (lru *iKeyTypeValueTypeLRU) Get(key KeyType) (value ValueType, err error) {
	var node *iKeyTypeValueType_NODE
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
