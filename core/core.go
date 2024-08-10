package core

import (
	pb "darwin-cache/core/rpc"
	"darwin-cache/core/singleflight"
	"fmt"
	"log"
	"sync"
)

// CacheGroup是缓存的命名空间，每个CacheGroup拥有一个唯一的名称
type CacheGroup struct {
	name        string              //当前缓存的命名空间
	dataFetcher Fetcher             //回调函数，当缓存不存在时，调用该函数获取源数据
	mainCache   cache               //缓存容器
	peers       PeerPicker          //节点选择器
	loader      *singleflight.Group // singleflight.Group 用来确保每个 key 只会被回调函数调用一次
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
		loader:      &singleflight.Group{},
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

func (g *CacheGroup) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// loadOtherData 方法，loadOtherData 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取）
// getLocally 调用用户回调函数 g.getter.Get() 获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
func (g *CacheGroup) loadOtherData(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.fetchFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[darwin-cache] Failed to get from peer", err)
			}
		}
		return g.fetchFromLocal(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return
}

// 从本地获取数据
func (g *CacheGroup) fetchFromLocal(key string) (ByteView, error) {
	bytes, err := g.dataFetcher.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.addToMainCache(key, value)
	return value, nil
}

// 从其他节点获取数据
func (g *CacheGroup) fetchFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

// 将数据添加到缓存
func (g *CacheGroup) addToMainCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
