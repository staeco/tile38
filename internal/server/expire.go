package server

import (
	"log"
	"math/rand"
	"time"

	"github.com/tidwall/btree"
)

type exitem struct {
	key, id string
	at      time.Time
}

func (a *exitem) Less(v btree.Item, ctx interface{}) bool {
	b := v.(*exitem)
	if a.at.Before(b.at) {
		return true
	}
	if a.at.After(b.at) {
		return false
	}
	if a.key < b.key {
		return true
	}
	if a.key > b.key {
		return false
	}
	return a.id < b.id
}

// fillExpiresList occurs once at startup
func (c *Server) fillExpiresList() {
	c.exlistmu.Lock()
	c.exlist = c.exlist[:0]
	for key, m := range c.expires {
		for id, at := range m {
			c.exlist = append(c.exlist, exitem{key, id, at})
		}
	}
	c.exlistmu.Unlock()
}

// clearIDExpires clears a single item from the expires list.
func (c *Server) clearIDExpires(key, id string) (cleared bool) {
	if len(c.expires) == 0 {
		return false
	}
	m, ok := c.expires[key]
	if !ok {
		return false
	}
	_, ok = m[id]
	if !ok {
		return false
	}
	delete(m, id)
	return true
}

// clearKeyExpires clears all items that are marked as expires from a single key.
func (c *Server) clearKeyExpires(key string) {
	delete(c.expires, key)
}

// expireAt marks an item as expires at a specific time.
func (c *Server) expireAt(key, id string, at time.Time) {
	m := c.expires[key]
	if m == nil {
		m = make(map[string]time.Time)
		c.expires[key] = m
	}
	m[id] = at
	c.exlistmu.Lock()
	c.exlist = append(c.exlist, exitem{key, id, at})
	c.exlistmu.Unlock()
}

// getExpires returns the when an item expires.
func (c *Server) getExpires(key, id string) (at time.Time, ok bool) {
	if len(c.expires) == 0 {
		return at, false
	}
	m, ok := c.expires[key]
	if !ok {
		return at, false
	}
	at, ok = m[id]
	return at, ok
}

// hasExpired returns true if an item has expired.
func (c *Server) hasExpired(key, id string) bool {
	at, ok := c.getExpires(key, id)
	if !ok {
		return false
	}
	return time.Now().After(at)
}

// backgroundExpiring watches for when items that have expired must be purged
// from the database. It's executes 10 times a seconds.
func (c *Server) backgroundExpiring() {
	rand.Seed(time.Now().UnixNano())
	var purgelist []exitem
	for {
		if c.stopBackgroundExpiring.on() {
			return
		}
		now := time.Now()
		purgelist = purgelist[:0]
		c.exlistmu.Lock()
		for i := 0; i < 20 && len(c.exlist) > 0; i++ {
			ix := rand.Int() % len(c.exlist)
			if now.After(c.exlist[ix].at) {
				// purge from exlist
				purgelist = append(purgelist, c.exlist[ix])
				c.exlist[ix] = c.exlist[len(c.exlist)-1]
				c.exlist = c.exlist[:len(c.exlist)-1]
			}
		}
		c.exlistmu.Unlock()
		if len(purgelist) > 0 {
			c.mu.Lock()
			for _, item := range purgelist {
				if c.hasExpired(item.key, item.id) {
					// purge from database
					msg := &Message{}
					msg.Args = []string{"del", item.key, item.id}
					_, d, err := c.cmdDel(msg)
					if err != nil {
						c.mu.Unlock()
						log.Fatal(err)
						continue
					}
					if err := c.writeAOF(msg.Args, &d); err != nil {
						c.mu.Unlock()
						log.Fatal(err)
						continue
					}
				}
			}
			c.mu.Unlock()
			if len(purgelist) > 5 {
				continue
			}
		}
		time.Sleep(time.Second / 10)
	}
}
