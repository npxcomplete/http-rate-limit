package mesh

import (
	"context"

	"github.com/twitchtv/twirp"
)

// PeerGroup implements a simple twirp service for managing peers.
type PeerGroup struct {
	hooks *twirp.ServerHooks
}

// NewPeerGroup creates a PeerGroup with default hooks.
func NewPeerGroup() *PeerGroup {
	return &PeerGroup{
		hooks: &twirp.ServerHooks{},
	}
}

// JoinPeerGroup registers a peer in the group.
func (pg *PeerGroup) JoinPeerGroup(ctx context.Context, req *JoinPeerGroupRequest) (*JoinPeerGroupResponse, error) {
	// TODO: add join logic
	return &JoinPeerGroupResponse{Joined: true}, nil
}

// BroadcastUpdates sends updates to all members of the group.
func (pg *PeerGroup) BroadcastUpdates(ctx context.Context, req *BroadcastUpdatesRequest) (*BroadcastUpdatesResponse, error) {
	// TODO: add broadcast logic
	return &BroadcastUpdatesResponse{Count: int32(len(req.Updates))}, nil
}

// JoinPeerGroupRequest represents a join request.
type JoinPeerGroupRequest struct {
	PeerId string
}

// JoinPeerGroupResponse represents a join response.
type JoinPeerGroupResponse struct {
	Joined bool
}

// BroadcastUpdatesRequest carries updates to send.
type BroadcastUpdatesRequest struct {
	Updates []string
}

// BroadcastUpdatesResponse contains acknowledgement info.
type BroadcastUpdatesResponse struct {
	Count int32
}

var _ interface {
	JoinPeerGroup(context.Context, *JoinPeerGroupRequest) (*JoinPeerGroupResponse, error)
	BroadcastUpdates(context.Context, *BroadcastUpdatesRequest) (*BroadcastUpdatesResponse, error)
} = (*PeerGroup)(nil)
