package algo

import (
	"jinx/pkg/util/types"
	"sync"
)

func RoundRobin(servers []types.UpStreamServer, currentServer int, lock *sync.Mutex) types.UpStreamServer {
	lock.Lock()
	defer lock.Unlock()
	nextServerIndex := (currentServer + 1) % len(servers)
	return servers[nextServerIndex]
}
