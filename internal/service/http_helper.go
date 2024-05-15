package service

import (
	"leicache/utils/logger"
	"log"
	"net/http"
)

func StartHTTPCacheServer(addr string, addrs []string, leicache *Group) {
	peers := NewHTTPPool(addr)
	peers.UpdatePeers(addrs...)
	leicache.RegisterServer(peers)
	logger.LogrusObj.Infof("service is running at %v", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func StartHTTPAPIServer(apiAddr string, leicache *Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := leicache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			_, err = w.Write(view.ByteSlice())
			if err != nil {
				return
			}
		}))
	logger.LogrusObj.Infof("fontend server is running at %v", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
