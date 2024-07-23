package core

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGetter(t *testing.T) {
	var f Fetcher = FetchFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("callback failed")
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	//如果缓存不存在，则调用该函数,该函数会去db中查找
	getter := FetchFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key]++ //统计缓存命中次数
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	})
	cacheGroup := NewCacheGroup("scores", 2<<10, getter)
	for k, v := range db {
		//缓存未命中测试
		if view, err := cacheGroup.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		//缓存命中测试
		if _, err := cacheGroup.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}
	//缓存不存在测试
	if view, err := cacheGroup.Get("unknown"); err == nil {
		t.Fatalf("the value of unknown should be empty, but %s got", view)
	}
}
