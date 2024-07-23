package main

import (
	"darwin-cache/core"
	"fmt"
	"log"
	"net/http"
)

var addr = "localhost:9999"

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	getter := core.FetchFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	})
	core.NewCacheGroup("scores", 2<<10, getter)
	peers := core.NewHTTPPool(addr)
	log.Println("darwin-cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
