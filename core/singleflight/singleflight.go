package singleflight

import "sync"

// call 代表正在进行中，或者已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 是 singleflight 的主数据结构，管理不同 key 的请求(call)
type Group struct {
	mutex       sync.Mutex
	cachedCalls map[string]*call //lazy initialization
}

// 针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mutex.Lock()
	// 如果 g.cachedCalls == nil，说明是第一次调用 Do，需要初始化 g.cachedCalls
	if g.cachedCalls == nil {
		g.cachedCalls = make(map[string]*call)
	}
	// 如果 key 对应的请求正在进行中，则等待
	if c, ok := g.cachedCalls[key]; ok {
		g.mutex.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	// 如果 key 对应的请求还没有进行，则新建一个请求
	c := new(call)
	c.wg.Add(1)
	g.cachedCalls[key] = c
	g.mutex.Unlock()

	// 调用 fn，获取返回值
	c.val, c.err = fn()
	c.wg.Done()

	// 删除 key 对应的请求
	g.mutex.Lock()
	delete(g.cachedCalls, key)
	g.mutex.Unlock()

	return c.val, c.err
}
