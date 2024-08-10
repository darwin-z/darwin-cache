package main

import (
	"darwin-cache/core"
	"flag"
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

func createGroup() *core.CacheGroup {
	getter := core.FetchFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	})
	return core.NewCacheGroup("scores", 2<<10, getter)
}

func startCacheServer(addr string, peerAddrs []string, group *core.CacheGroup) {
	peers := core.NewHTTPPool(addr)
	peers.SetPeers(peerAddrs...)
	group.RegisterPeers(peers)
	log.Println("darwin-cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, group *core.CacheGroup) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "darwin-cache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	peerAddrs := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var peers []string
	for _, v := range peerAddrs {
		peers = append(peers, v)
	}

	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(peerAddrs[port], peers, group)
}
