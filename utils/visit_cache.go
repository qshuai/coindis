package utils

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type InfoCache struct {
	sync.Mutex
	addressMap map[string]time.Time
	ipMap      map[string]time.Time
}

func (c *InfoCache) IsExit(key string) bool {
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

func (c *InfoCache) InsertNew(address, ip string) {
	c.Lock()
	c.Unlock()

	now := time.Now()
	c.addressMap[address] = now
	c.ipMap[ip] = now
}

func (c *InfoCache) RemoveOne(address, ip string) {
	c.Lock()
	c.Unlock()

	delete(c.addressMap, address)
	delete(c.ipMap, ip)
}

func (c *InfoCache) Clean() {
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

	logrus.Debug("clear addresses cache")
}

func New(capacity int) *InfoCache {
	return &InfoCache{
		addressMap: make(map[string]time.Time, capacity),
		ipMap:      make(map[string]time.Time, capacity),
	}
}
