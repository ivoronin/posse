package main

import (
	"log"

	"github.com/ivoronin/posse/fsm"
)

var (
	PeerRxStateInit  fsm.State = "init"
	PeerRxStateDown  fsm.State = "down"
	PeerRxStateUp    fsm.State = "up"
	PeerRxStateError fsm.State = "error"

	PeerTxStateUp       fsm.State = "up"
	PeerTxStateDowntime fsm.State = "downtime"
	PeerTxStateError    fsm.State = "error"
)

var (
	PeerRxEventBlockReadErr   fsm.Event = "BlockReadErr"
	PeerRxEventBlockReadStale fsm.Event = "BlockReadStale"
	PeerRxEventBlockReadNew   fsm.Event = "BlockReadNew"

	PeerTxEventBlockWritten  fsm.Event = "BlockWritten"
	PeerTxEventBlockWriteErr fsm.Event = "BlockWriteErr"
	PeerTxEventBlockSkipped  fsm.Event = "BlockSkipped"
)

type Peer struct {
	RxFSM *fsm.FSM
	TxFSM *fsm.FSM
}

func peerRxStateChanged(from fsm.State, to fsm.State, evt fsm.Event) {
	if from == to {
		return
	}
	log.Printf("peer rx status: %s -> %s", from, to)
}

func NewPeer(maxStale uint64) *Peer {
	peer := new(Peer)
	AllRxStates := []fsm.State{PeerRxStateInit, PeerRxStateDown, PeerRxStateUp, PeerRxStateError}
	peer.RxFSM = fsm.NewFSM(
		PeerRxStateInit,
		[]fsm.Transition{
			{
				Evt: PeerRxEventBlockReadErr,
				Src: AllRxStates,
				Dst: PeerRxStateError,
			},
			{
				Evt:      PeerRxEventBlockReadStale,
				Src:      AllRxStates,
				Dst:      PeerRxStateDown,
				MinTimes: uint(maxStale),
			},
			{
				Evt: PeerRxEventBlockReadNew,
				Src: AllRxStates,
				Dst: PeerRxStateUp,
			},
		},
		peerRxStateChanged,
	)

	AllTxStates := []fsm.State{PeerTxStateUp, PeerTxStateDowntime, PeerTxStateError}
	peer.TxFSM = fsm.NewFSM(
		PeerTxStateUp,
		[]fsm.Transition{
			{
				Evt: PeerTxEventBlockWritten,
				Src: AllTxStates,
				Dst: PeerTxStateUp,
			},
			{
				Evt:      PeerTxEventBlockSkipped,
				Src:      AllTxStates,
				Dst:      PeerTxStateDowntime,
				MinTimes: uint(maxStale - 1),
			},
			{
				Evt: PeerTxEventBlockWriteErr,
				Src: AllTxStates,
				Dst: PeerTxStateError,
			},
		},
		nil,
	)
	return peer
}
