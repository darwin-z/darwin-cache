package lru

import "container/list"

/*
LRU缓存淘汰策略
*/

// 缓存对象必须实现 Value 接口，即 Len() int 方法，返回其所占的内存大小。
type Value interface {
	Len() int //返回值所占用的字节数
}

type entry struct {
	key   string
	value Value
}

// 一个LRUCacheManager对象表示一个LRU缓存,并发访问不安全
type LRUCacheManager struct {
	maxBytes  int64                       //最大允许存储的字节数
	nbytes    int64                       //当前已存储的字节数
	dl        *list.List                  //双向链表，用于存储缓存项的顺序，实现LRU策略
	cache     map[string]*list.Element    //键是字符串，值是双向链表中对应节点的指针，用于快速查找缓存项
	OnEvicted func(key string, val Value) // 某个缓存项被移除时的回调函数
}

// 实例Cache对象
func New(maxBytes int64, onEvicted func(string, Value)) *LRUCacheManager {
	return &LRUCacheManager{
		maxBytes:  maxBytes,
		dl:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 获取当前缓存项个数
func (c *LRUCacheManager) Len() int {
	return c.dl.Len()
}

// 删除最近最少访问的缓存项
func (c *LRUCacheManager) RemoveOldest() {
	ele := c.dl.Back() //队尾元素就是当前最少访问的
	if ele == nil {
		return
	}
	c.dl.Remove(ele) //从链表中删除
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key)                                //从字典中删除
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) //更新当前所占用的字节数
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value) //调用回调函数
	}
}

// 查找缓存项
func (c *LRUCacheManager) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { //如果键存在，则更新对应节点的值，并将该节点移到队首
		c.dl.MoveToFront(ele)                                  //更新这个缓存的访问
		kv := ele.Value.(*entry)                               //取出节点的值
		c.nbytes += int64(value.Len()) - int64(kv.value.Len()) //更新这个key的value所占用的字节数
		kv.value = value                                       //更新这个key的value
	} else { //不存在则添加
		ele := c.dl.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	//如果当前所占用的字节数大于最大允许的字节数，则移除最少访问的缓存项
	//maxBytes为0表示不限制存储空间
	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// 查找缓存项
func (c *LRUCacheManager) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.dl.MoveToFront(ele) //更新访问时间
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}
