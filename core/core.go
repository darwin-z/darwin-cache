package core

import (
	"fmt"
	"log"
	"sync"
)

// CacheGroup是缓存的命名空间，每个CacheGroup拥有一个唯一的名称
type CacheGroup struct {
	name        string  //当前缓存的命名空间
	dataFetcher Fetcher //回调函数，当缓存不存在时，调用该函数获取源数据
	mainCache   cache   //缓存容器
}

// Fetcher回调函数，当缓存不存在时，调用该函数获取源数据
type Fetcher interface {
	Get(key string) ([]byte, error)
}

// FetchFunc是一个回调函数类型，满足Getter接口
type FetchFunc func(key string) ([]byte, error)

var (
	mu     sync.RWMutex
	groups = make(map[string]*CacheGroup)
)

// 实现Getter接口
func (f FetchFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 实例化Group对象
func NewCacheGroup(name string, cacheBytes int64, getter Fetcher) *CacheGroup {
	if getter == nil {
		panic("getter is nil")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &CacheGroup{
		name:        name,
		dataFetcher: getter,
		mainCache:   cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// 获取Group对象
func GetCacheGroup(name string) *CacheGroup {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 从 mainCache 中查找缓存，如果存在则返回缓存值。
func (g *CacheGroup) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[darwin-cache] hit key: ", key)
		return v, nil
	}
	return g.loadOtherData(key) // 缓存不存在，调用 load 方法
}

// loadOtherData 方法，loadOtherData 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取）
// getLocally 调用用户回调函数 g.getter.Get() 获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
func (g *CacheGroup) loadOtherData(key string) (value ByteView, err error) {
	return g.fetchLocalData(key)
}

// 从本地获取数据
func (g *CacheGroup) fetchLocalData(key string) (ByteView, error) {
	bytes, err := g.dataFetcher.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.addToMainCache(key, value)
	return value, nil
}

// 将数据添加到缓存
func (g *CacheGroup) addToMainCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
