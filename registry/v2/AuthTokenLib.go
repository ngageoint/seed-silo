package v2

import (
	"fmt"
	"sync"
)

type authtokenmap struct {
	authtokens map[string]APIToken
	mu         sync.RWMutex
}

func (atmap *authtokenmap) Set(key string, data APIToken) {
	atmap.mu.Lock()
	defer atmap.mu.Unlock()
	atmap.authtokens[key] = data
}

func (atmap *authtokenmap) Get(key string) (APIToken, error) {
	atmap.mu.RLock()
	defer atmap.mu.RUnlock()
	item, ok := atmap.authtokens[key]
	if !ok {
		return item, fmt.Errorf("The '%s' is not presented", key)
	}
	return item, nil
}

func (atmap *authtokenmap) Check(key string) bool {
	_, ok := atmap.authtokens[key]
	return ok
}

var (
	atmap *authtokenmap
)

func Authtokenmap() *authtokenmap {
	if atmap == nil {
		atmap = &authtokenmap{
			authtokens: make(map[string]APIToken),
		}
	}
	return atmap
}
