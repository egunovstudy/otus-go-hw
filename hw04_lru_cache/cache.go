package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value any) bool
	Get(key Key) (any, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value any
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem // key -> элемент списка, внутри лежит cacheItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value any) bool {
	if c.capacity <= 0 {
		return false
	}

	if node, ok := c.items[key]; ok {
		node.Value.(*cacheItem).value = value
		c.queue.MoveToFront(node)
		return true
	}

	item := &cacheItem{key: key, value: value}
	node := c.queue.PushFront(item)
	c.items[key] = node

	if c.queue.Len() > c.capacity {
		last := c.queue.Back()
		if last != nil {
			ci := last.Value.(*cacheItem)
			delete(c.items, ci.key)
			c.queue.Remove(last)
		}
	}

	return false
}

func (c *lruCache) Get(key Key) (any, bool) {
	if c.capacity <= 0 {
		return nil, false
	}

	node, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.queue.MoveToFront(node)
	return node.Value.(*cacheItem).value, true
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}
