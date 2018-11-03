package controllers

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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

func (c *infoCache) removeOne(address, ip string) {
	c.Lock()
	c.Unlock()

	delete(c.addressMap, address)
	delete(c.ipMap, ip)
}

func (c *infoCache) clean() {
	c.Lock()
	c.Unlock()

	now := time.Now()
	for address, t := range c.addressMap {
		// over 60 minute, clean
		diff := now.Sub(t).Hours()
		if diff >= 6 {
			delete(c.addressMap, address)
		}
	}

	for ip, t := range c.ipMap {
		// over 60 minute, clean
		diff := now.Sub(t).Hours()
		if diff >= 6 {
			delete(c.ipMap, ip)
		}
	}

	logrus.Info("clear addresses cache")
}
