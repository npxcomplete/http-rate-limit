package ratelimit

import (
	"sync"
	"time"
)

// Capacity tracks rate limiting information for a peer.
type Capacity struct {
	Burst       uint32
	Rate        uint32
	Available   int64
	LastUpdated time.Time
	RecentSpend uint32
}

// PeerGroup manages capacity information for a set of peers using an
// internal map keyed by peer identifier.
type PeerGroup struct {
	mu    sync.RWMutex
	peers map[string]*Capacity
}

// NewPeerGroup creates a PeerGroup with no peers.
func NewPeerGroup() *PeerGroup {
	return &PeerGroup{peers: make(map[string]*Capacity)}
}

// Get returns the Capacity for the given peer id. The boolean return value
// indicates whether a Capacity was present.
func (g *PeerGroup) Get(id string) (*Capacity, bool) {
	g.mu.RLock()
	c, ok := g.peers[id]
	g.mu.RUnlock()
	return c, ok
}

// Set assigns the given Capacity to the peer id.
func (g *PeerGroup) Set(id string, c *Capacity) {
	g.mu.Lock()
	g.peers[id] = c
	g.mu.Unlock()
}

// Delete removes the peer from the group.
func (g *PeerGroup) Delete(id string) {
	g.mu.Lock()
	delete(g.peers, id)
	g.mu.Unlock()
}
