package lru

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

// 测试添加缓存
func TestAdd(t *testing.T) {
	cache := New(int64(0), nil)
	cache.Add("hello", String("world"))
	if v, ok := cache.Get("hello"); !ok || string(v.(String)) != "world" {
		t.Fatalf("cache hit hello=world failed")
	}
	if _, ok := cache.Get("test"); ok {
		t.Fatalf("缓存键 test 不存在")
	}
}

// 测试触发LRU策略
func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "v1", "v1", "v2"
	cap := len((k1 + v1) + (k2 + v2))
	cache := New(int64(cap), nil)
	cache.Add(k1, String(v1))
	cache.Add(k2, String(v2))
	cache.Add(k3, String(v3))
	//添加 key1,key2当添加到key3时,存储空间超过了maxBytes,触发LRU淘汰策略,删除队尾元素
	if _, ok := cache.Get("key1"); ok || cache.Len() != 2 {
		t.Fatalf("RemoveOldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	deletedKeys := make([]string, 0)
	cache := New(int64(10), func(key string, value Value) {
		deletedKeys = append(deletedKeys, key)
	})
	cache.Add("key1", String("123456"))
	cache.Add("k2", String("k2"))
	cache.Add("k3", String("k3"))
	cache.Add("k4", String("k4"))
	//期待被删除的
	expect := []string{"key1", "k2"}
	if !reflect.DeepEqual(expect, deletedKeys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
