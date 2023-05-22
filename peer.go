package main

import (
	"log"

	"github.com/ivoronin/posse/fsm"
)

var (
	PeerStateInit    fsm.State = "init"
	PeerStateDown    fsm.State = "down"
	PeerStateUp      fsm.State = "up"
	PeerStateUnknown fsm.State = "unknown"
)

var (
	PeerEventBlockReadErr   fsm.Event = "BlockReadErr"
	PeerEventBlockReadStale fsm.Event = "BlockReadStale"
	PeerEventBlockReadNew   fsm.Event = "BlockReadNew"
)

type Peer struct {
	fsm *fsm.FSM
}

func peerStateChanged(from fsm.State, to fsm.State, evt fsm.Event) {
	if from == to {
		return
	}
	log.Printf("peer status: %s -> %s", from, to)
}

func NewPeer(maxStale uint64) *Peer {
	peer := new(Peer)
	AnyState := []fsm.State{PeerStateInit, PeerStateDown, PeerStateUp, PeerStateUnknown}
	peer.fsm = fsm.NewFSM(
		PeerStateInit,
		[]fsm.Transition{
			{
				Evt: PeerEventBlockReadErr,
				Src: AnyState,
				Dst: PeerStateUnknown,
			},
			{
				Evt:      PeerEventBlockReadStale,
				Src:      AnyState,
				Dst:      PeerStateDown,
				MinTimes: uint(maxStale),
			},
			{
				Evt: PeerEventBlockReadNew,
				Src: AnyState,
				Dst: PeerStateUp,
			},
		},
		peerStateChanged,
	)
	return peer
}

func (p *Peer) State() fsm.State {
	return p.fsm.CurrentState
}

func (p *Peer) Event(evt fsm.Event) {
	p.fsm.Event(evt)
}
