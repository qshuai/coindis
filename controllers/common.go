package controllers

import (
	"sync"
	"sync/atomic"
	"time"
)

type infoCache struct {
	sync.Mutex
	addressMap map[string]time.Time
	ipMap      map[string]time.Time
}

func newInfoCache() *infoCache {
	return &infoCache{
		addressMap: make(map[string]time.Time),
		ipMap:      make(map[string]time.Time),
	}
}

func (c *infoCache) isExit(key string) bool {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.addressMap[key]; ok {
		return true
	}

	if _, ok := c.ipMap[key]; ok {
		return true
	}

	return false
}

func (c *infoCache) insertNew(address, ip string) {
	c.Lock()
	c.Unlock()

	now := time.Now()
	c.addressMap[address] = now
	c.ipMap[ip] = now
}

func (c *infoCache) clean() {
	c.Lock()
	c.Unlock()

	now := time.Now()
	for address, t := range c.addressMap {
		// over 1 minute, clean
		if now.Sub(t) > time.Duration(60) {
			delete(c.addressMap, address)
		}
	}

	for ip, t := range c.addressMap {
		// over 1 minute, clean
		if now.Sub(t) > time.Duration(60) {
			delete(c.addressMap, ip)
		}
	}
}

func updateBalance() {
	client := Client()
	amount, err := client.GetBalance("")
	if err == nil {
		atomic.SwapInt64(&balance, int64(amount))
	}
}
