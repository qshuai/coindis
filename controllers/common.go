package controllers

import (
	"sync"
	"time"
)

type infoCache struct {
	sync.Mutex
	AddressMap map[string]time.Time
	IpMap      map[string]time.Time
}

func newInfoCache() *infoCache {
	return &infoCache{
		AddressMap: make(map[string]time.Time),
		IpMap:      make(map[string]time.Time),
	}
}

func (c *infoCache) isExit(key string) bool {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.AddressMap[key]; ok {
		return true
	}

	if _, ok := c.IpMap[key]; ok {
		return true
	}

	return false
}

func (c *infoCache) insertNew(address, ip string) {
	c.Lock()
	c.Unlock()

	now := time.Now()
	c.AddressMap[address] = now
	c.IpMap[ip] = now
}

func (c *infoCache) clean() {
	c.Lock()
	c.Unlock()

	now := time.Now()
	for address, t := range c.AddressMap {
		// over 1 minute, clean
		if now.Sub(t) > time.Duration(60) {
			delete(c.AddressMap, address)
		}
	}

	for ip, t := range c.AddressMap {
		// over 1 minute, clean
		if now.Sub(t) > time.Duration(60) {
			delete(c.AddressMap, ip)
		}
	}
}
