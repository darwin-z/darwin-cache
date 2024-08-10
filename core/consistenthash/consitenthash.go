package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 是一个哈希函数的接口
type Hash func(data []byte) uint32

// Map 是一致性哈希的主数据结构
type Map struct {
	hash     Hash           // hash函数
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 虚拟节点与真实节点的映射表
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	// 如果没有传入Hash函数，就使用默认的crc32.ChecksumIEEE
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 方法用来添加缓存节点，参数为真实节点的名称，比如使用IP
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 对每一个真实节点key，对应创建m.replicas个虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) // 通过哈希函数计算虚拟节点的哈希值
			m.keys = append(m.keys, hash)                      // 将虚拟节点放入环上
			m.hashMap[hash] = key                              // 虚拟节点与真实节点的映射关系
		}
	}
	sort.Ints(m.keys) // 环上的哈希值排序
}

// Get 方法根据给定的对象获取最靠近它的那个节点 
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	// 计算key的哈希值
	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点的下标idx
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 通过虚拟节点的下标找到真实节点的下标
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
